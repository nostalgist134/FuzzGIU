package fuzz

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCommon"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/inputHandler"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageReact"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageSend"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/input"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/rp"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var JQ fuzzCommon.JobQueue = make([]*fuzzTypes.Fuzz, 0)
var SendMetaPool = sync.Pool{
	New: func() any { return new(fuzzTypes.SendMeta) },
}

// Rp 协程池指针
var Rp *rp.RoutinePool

// trySubmit 尝试提交任务，若提交失败，则先从队列中取出所有结果并处理，再提交
func trySubmit(task rp.Task, fuzz1 *fuzzTypes.Fuzz) bool {
	for !Rp.Submit(task, time.Millisecond*10) {
		// 处理外部输入
		if handleInputStack(fuzz1) {
			return true
		}
		// 若处于暂停状态，则不消耗结果
		if Rp.Status() == rp.StatPause {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		// 将结果队列全部消耗而不是取一个，避免陷入handleReaction->trySubmit->handleReaction->...的无限递归
		for r := Rp.GetSingleResult(); r != nil; r = Rp.GetSingleResult() {
			// 若确定jobStop，就可以不用再取结果了，直接返回上一层直到doFuzz，然后退出
			if jobStop, _ := handleReaction(r, fuzz1); jobStop {
				return jobStop
			}
		}
	}
	return false
}

// tryGetUrlScheme 尝试获取url scheme，若整个fuzz过程中url的scheme不会变化（不包含任何fuzz keyword）则可将其缓存
// 从而避免在SendRequest中反复调用url.Parse消耗资源
func tryGetUrlScheme(req *fuzzTypes.Req, keywords []string) string {
	u, err := url.Parse(req.URL)
	if err != nil {
		return ""
	}
	scheme := u.Scheme
	for _, k := range keywords {
		if strings.Index(scheme, k) != -1 {
			return ""
		}
	}
	return scheme
}

func handleInputStack(fuzz1 *fuzzTypes.Fuzz) bool {
	for inp, hasInput := input.GetSingleInput(); hasInput; inp, hasInput = input.GetSingleInput() {
		drainRp(fuzz1)
		if err := inputHandler.HandleInput(inp); err != nil {
			// 停止当前任务
			if errors.Is(err, fuzzCommon.ErrJobStop) {
				output.Logf(common.OutputToWhere, "job stopped by %v", inp.Peer.RemoteAddr())
				Rp.Clear()
				return true
			}
			output.Logf(common.OutputToWhere, "input error: %v", err)
		}
	}
	return false
}

// handleReaction 根据fuzz设置处理反应
func handleReaction(r *fuzzTypes.Reaction, fuzz1 *fuzzTypes.Fuzz) (bool, bool) {
	defer common.PutReaction(r)
	stopJob := false
	addReq := false

	if r.Flag&fuzzTypes.ReactAddJob != 0 && r.NewJob != nil {
		k, p := stageReact.GetReactTraceInfo(r)
		if k != nil && p != nil {
			output.Logf(common.OutputToWhere, "task with %s:%s added job", k, p)
		}
		JQ.AddJob(r.NewJob)
		// job 总数加1
		output.SetJobTotal(output.GetCounterValue(output.TotalJob) + 1)
	}
	if r.Flag&fuzzTypes.ReactStopJob != 0 {
		output.Log(common.OutputToWhere, "job stopped by react")
		stopJob = true
	}
	if r.Flag&fuzzTypes.ReactAddReq != 0 && r.NewReq != nil {
		addReq = true
		k, p := stageReact.GetReactTraceInfo(r)
		newSend := (SendMetaPool.Get()).(*fuzzTypes.SendMeta)
		newTask := func() *fuzzTypes.Reaction {
			newSend.Timeout = fuzz1.Send.Timeout
			newSend.Retry = fuzz1.Send.Retry
			newSend.RetryRegex = fuzz1.Send.RetryRegex
			newSend.RetryCode = fuzz1.Send.RetryCode
			newSend.HttpFollowRedirects = fuzz1.Send.HttpFollowRedirects
			newSend.Request = r.NewReq
			resp := stageSend.SendRequest(newSend, "")
			reaction := stageReact.React(fuzz1, newSend.Request, resp, []string{""},
				[]string{fmt.Sprintf("add via react by %s:%s", k, p)}, nil)
			SendMetaPool.Put(newSend)
			// task数加1
			output.AddTaskCounter()
			return reaction
		}
		stopJob = trySubmit(newTask, fuzz1)
		// task总数加1
		output.SetTaskTotal(output.GetCounterValue(output.TotalTask) + 1)
	}
	if r.Flag&fuzzTypes.ReactExit != 0 {
		output.FinishOutput(common.OutputToWhere)
		if common.OutputToWhere&output.OutToScreen != 0 {
			output.ScreenClose()
		}
		fmt.Println("exit by react")
		os.Exit(0)
	}
	return stopJob, addReq
}

// drainRp 消耗协程池中的所有任务和结果
func drainRp(fuzz1 *fuzzTypes.Fuzz) bool {
	for {
		canStop := true // canStop 标记了结果是否已经消耗完毕
		// 循环1：跑到Rp等待不阻塞（也就是任务队列为空）为止
		for !Rp.Wait(time.Millisecond * 10) {
			for r := Rp.GetSingleResult(); r != nil; r = Rp.GetSingleResult() {
				stopJob, addReq := handleReaction(r, fuzz1)
				if stopJob {
					Rp.Clear()
					return true
				}
				if addReq {
					canStop = false
				}
			}
		}
		// 循环2：在确保任务队列为空之后，再把结果队列的结果全部消耗完毕
		for r := Rp.GetSingleResult(); r != nil; r = Rp.GetSingleResult() {
			stopJob, addReq := handleReaction(r, fuzz1)
			if stopJob {
				Rp.Clear()
				return true
			}
			if addReq {
				canStop = false
			}
		}
		// 若上面两个循环都跑完了，也没有添加新请求，这种情况下任务队列和结果队列均为空，没可能再有新请求，因此视作结果消耗完毕
		if canStop {
			break
		}
	}
	return false
}

// doFuzz fuzz任务实际执行的函数
func doFuzz(fuzz1 *fuzzTypes.Fuzz, jobId int) time.Duration {
	fuzzCommon.SetCurFuzz(fuzz1) // 设置当前任务，用于外部控制获取

	timeStart := time.Now()

	// 预加载插件
	if err := preLoadJobPlugin(fuzz1); err != nil {
		output.Logf(common.OutputToWhere, "job#%d preload plugins failed: %v\njob will be skipped", jobId, err)
		return time.Since(timeStart)
	}

	// 递归边界（虽然在react中也有判断，但是在这里也防备一下）
	if fuzz1.React.RecursionControl.RecursionDepth > fuzz1.React.RecursionControl.MaxRecursionDepth {
		return time.Since(timeStart)
	}

	// 初始化协程池
	if Rp == nil {
		Rp = rp.New(fuzz1.Misc.PoolSize)
		Rp.Start()
	} else {
		Rp.Resize(fuzz1.Misc.PoolSize)
	}

	fuzz1 = stagePreprocess.Preprocess(fuzz1, fuzz1.Preprocess.Preprocessors)

	if len(fuzz1.Preprocess.PlTemp) == 0 {
		output.Logf(common.OutputToWhere, "job#%d has no fuzz keyword, skip", jobId)
		return time.Since(timeStart)
	}

	// fuzz关键字的处理
	keywords := make([]string, 0)
	loopLen := int64(1)
	// 计算长度(loopLen)
	for keyword, pt := range fuzz1.Preprocess.PlTemp {
		keywords = append(keywords, keyword)
		// sniper模式
		if fuzz1.Preprocess.Mode == "sniper" || fuzz1.React.RecursionControl.MaxRecursionDepth > 0 {
			// 如果采用递归扫描或者sniper模式，则只使用一个关键词
			loopLen = int64(len(pt.PlList))
			if fuzz1.Preprocess.Mode == "sniper" {
				loopLen *= int64(common.GetKeywordNum(&fuzz1.Preprocess.ReqTemplate, keyword))
			}
			break
		}
		switch fuzz1.Preprocess.Mode {
		// clusterbomb模式：遍历每个关键词对应payload列表的所有组合
		case "clusterbomb":
			loopLen *= int64(len(pt.PlList))
		// pitchfork模式：每个关键字的payload列表在遍历时下标会同步替换，因此以最小的payload列表为准
		case "pitchfork":
			if int64(len(pt.PlList)) < loopLen {
				loopLen = int64(len(pt.PlList))
			}
		// pitchfork-cycle模式：以最大的payload列表为准，每个关键字的payload列表在遍历时下标会同步替换，较短的列表遍历完了则循环遍历
		case "pitchfork-cycle":
			if int64(len(pt.PlList)) > loopLen {
				loopLen = int64(len(pt.PlList))
			}
		default:
			fmt.Println("unsupported mode", fuzz1.Preprocess.Mode)
			os.Exit(1)
		}
	}

	output.SetTaskTotal(loopLen)
	output.ClearTaskCounter()

	var task func() *fuzzTypes.Reaction // task函数：实际放入池中执行的任务

	reqTemplate := common.ParseReqTemplate(&fuzz1.Preprocess.ReqTemplate, keywords) // 请求模板

	var plProcessorPlugins = make([][]fuzzTypes.Plugin, len(keywords)) // payload处理器插件
	for i, keyword := range keywords {
		plProcessorPlugins[i] = fuzz1.Preprocess.PlTemp[keyword].Processors
	}

	jobStop := false

	uScheme := tryGetUrlScheme(&fuzz1.Preprocess.ReqTemplate, keywords)

	payloadEachKeyword := make([]string, len(keywords))

	// 主循环
	for i := int64(0); i < loopLen; i++ {

		curInd := int64(len(fuzz1.Preprocess.PlTemp[keywords[0]].PlList))
		if fuzz1.Preprocess.Mode == "clusterbomb" && fuzz1.React.RecursionControl.MaxRecursionDepth <= 0 {
			curInd = i
		}

		// 根据模式生成任务
		if fuzz1.Preprocess.Mode != "sniper" && fuzz1.React.RecursionControl.MaxRecursionDepth <= 0 {
			for j := 0; j < len(keywords); j++ { // 遍历keywords列表，根据i选出每个关键字对应的payload
				switch fuzz1.Preprocess.Mode {
				// clusterbomb模式，遍历所有的payload组合
				case "clusterbomb":
					d := int64(len(fuzz1.Preprocess.PlTemp[keywords[len(keywords)-j-1]].PlList))
					r := curInd % d
					curInd /= d
					payloadEachKeyword[len(keywords)-j-1] =
						fuzz1.Preprocess.PlTemp[keywords[len(keywords)-j-1]].PlList[r]
				// pitchfork模式：每个关键字使用一样的payload下标
				case "pitchfork":
					payloadEachKeyword[j] = fuzz1.Preprocess.PlTemp[keywords[j]].PlList[i]
				// pitchfork-cycle模式：每次i循环下标都同步更新1，但payload列表到尾部后会从头再次开始
				case "pitchfork-cycle":
					payloadEachKeyword[j] = fuzz1.Preprocess.PlTemp[keywords[j]].PlList[i%
						int64(len(fuzz1.Preprocess.PlTemp[keywords[j]].PlList))]
				}
			}

			interior := common.GetStringSlice(len(payloadEachKeyword))
			copy(interior, payloadEachKeyword) // 将payloadEachPayload复制一份再用于闭包中

			task = func() *fuzzTypes.Reaction {
				// sendMeta对象
				send := (SendMetaPool.Get()).(*fuzzTypes.SendMeta)
				send.Timeout = fuzz1.Send.Timeout
				send.Retry = fuzz1.Send.Retry
				send.RetryRegex = fuzz1.Send.RetryRegex
				send.RetryCode = fuzz1.Send.RetryCode
				send.HttpFollowRedirects = fuzz1.Send.HttpFollowRedirects
				send.Proxy = ""

				// 代理轮询
				if len(fuzz1.Send.Proxies) > 0 {
					send.Proxy = fuzz1.Send.Proxies[i%int64(len(fuzz1.Send.Proxies))]
				}

				var cacheId int32

				processedPayloads := common.GetStringSlice(len(interior))
				for j, plugins := range plProcessorPlugins {
					processedPayloads[j] = stagePreprocess.PayloadProcessor(interior[j], plugins)
				}

				send.Request, cacheId = common.ReplacePayloadsByTemplate(reqTemplate, processedPayloads, -1)
				send.Request.HttpSpec.ForceHttps = fuzz1.Preprocess.ReqTemplate.HttpSpec.ForceHttps

				resp := stageSend.SendRequest(send, uScheme)
				reaction := stageReact.React(fuzz1, send.Request, resp, keywords, processedPayloads, nil)

				common.PutStringSlice(processedPayloads)
				common.PutStringSlice(interior)
				SendMetaPool.Put(send)
				common.ReleaseReqCache(cacheId)
				output.AddTaskCounter()
				return reaction
			}
		} else { // sniper模式或者递归模式
			keyword := keywords[0]
			payload := fuzz1.Preprocess.PlTemp[keyword].PlList[i%curInd]

			task = func() *fuzzTypes.Reaction {
				// sendMeta对象
				send := (SendMetaPool.Get()).(*fuzzTypes.SendMeta)
				send.Timeout = fuzz1.Send.Timeout
				send.Retry = fuzz1.Send.Retry
				send.RetryRegex = fuzz1.Send.RetryRegex
				send.RetryCode = fuzz1.Send.RetryCode
				send.HttpFollowRedirects = fuzz1.Send.HttpFollowRedirects
				send.Proxy = ""

				// 代理轮询
				if len(fuzz1.Send.Proxies) > 0 {
					send.Proxy = fuzz1.Send.Proxies[i%int64(len(fuzz1.Send.Proxies))]
				}

				processedPayload := payload
				processedPayload = stagePreprocess.PayloadProcessor(processedPayload, plProcessorPlugins[0])
				var recPos []int = nil
				var cacheId int32

				// payload替换
				if fuzz1.Preprocess.Mode == "sniper" &&
					fuzz1.React.RecursionControl.RecursionDepth <= fuzz1.React.RecursionControl.MaxRecursionDepth {
					// 同时启用sniper和递归
					send.Request, recPos, cacheId =
						common.ReplacePayloadTrackTemplate(reqTemplate, payload, int(i/curInd))
				} else if fuzz1.React.RecursionControl.RecursionDepth <=
					fuzz1.React.RecursionControl.MaxRecursionDepth {
					// 只启用递归
					send.Request, recPos, cacheId =
						common.ReplacePayloadTrackTemplate(reqTemplate, payload, -1)
				} else { // 只启用sniper
					send.Request, cacheId =
						common.ReplacePayloadsByTemplate(reqTemplate, []string{payload}, int(i/curInd))
				}

				send.Request.HttpSpec.ForceHttps = fuzz1.Preprocess.ReqTemplate.HttpSpec.ForceHttps

				resp := stageSend.SendRequest(send, uScheme)

				reaction := stageReact.React(fuzz1, send.Request, resp, []string{keyword},
					[]string{processedPayload}, recPos)

				SendMetaPool.Put(send)
				common.ReleaseReqCache(cacheId)
				output.AddTaskCounter()
				return reaction
			}
		}
		if trySubmit(task, fuzz1) {
			Rp.Clear()
			return time.Since(timeStart)
		}
		time.Sleep(fuzz1.Misc.DelayGranularity * time.Duration(fuzz1.Misc.Delay))
		// 任务提交后，从结果队列中取出所有结果并处理
		for r := Rp.GetSingleResult(); r != nil; r = Rp.GetSingleResult() {
			jobStop, _ = handleReaction(r, fuzz1)
			if jobStop {
				Rp.Clear()
				return time.Since(timeStart)
			}
		}
		// 处理外部输入
		if handleInputStack(fuzz1) {
			Rp.Clear()
			return time.Since(timeStart)
		}
	}
	drainRp(fuzz1)
	return time.Since(timeStart)
}

// DoSingleJob 执行单个fuzz任务及其衍生出的所有任务
func DoSingleJob(fuzz1 *fuzzTypes.Fuzz) {
	defer output.ScreenClose()

	if JQ == nil {
		JQ = make([]*fuzzTypes.Fuzz, 0)
	} else if len(JQ) > 0 { // 将jq清零，避免DoJobs执行已经执行过的任务
		JQ = JQ[:0]
	}

	outToMem := common.OutputToWhere & output.OutToMem
	// 初始化输出
	if fuzz1.React.OutSettings.NativeStdout {
		common.OutputToWhere = output.OutToNativeStdout
	} else {
		common.OutputToWhere = output.OutToScreen
	}
	if fuzz1.React.OutSettings.OutputFile != "" {
		common.OutputToWhere |= output.OutToFile
	}
	common.OutputToWhere |= outToMem
	output.InitOutput(fuzz1, common.OutputToWhere)

	jobId := int(output.GetCounterValue(output.TotalJob))
	timeLapsed := doFuzz(fuzz1, jobId)

	output.Logf(common.OutputToWhere, "Job#%d completed, time %v", jobId, timeLapsed)
	output.FinishOutput(common.OutputToWhere)
	output.AddJobCounter()

	// 若执行单个任务后添加了新任务，需要把新任务也全部执行
	if len(JQ) != 0 {
		DoJobs()
	}
}

func DoJobs() {
	// 仅当job总数为0时设置
	if output.GetCounterValue(output.TotalJob) == 0 {
		output.SetJobTotal(int64(len(JQ)))
	}
	defer output.ScreenClose()

	fuzzCommon.SetJQ(&JQ)

	i := 0
	toWhereShadow := int32(0)
	for ; i < len(JQ); i++ {
		// 根据OutSettings选则输出模式（termui界面、原生stdout）并初始化
		if JQ[i].React.OutSettings.NativeStdout {
			common.OutputToWhere = output.OutToNativeStdout
		} else {
			common.OutputToWhere = output.OutToScreen
		}
		if JQ[i].React.OutSettings.OutputFile != "" {
			common.OutputToWhere |= output.OutToFile
		}
		output.InitOutput(JQ[i], common.OutputToWhere) // 初始化

		timeLapsed := doFuzz(JQ[i], i) // 执行任务

		toWhereShadow = common.OutputToWhere

		// 如果下一个任务仍然使用同样文件以及同样输出格式，则不结束文件输出，追加到同一文件
		if i+1 < len(JQ) && JQ[i+1].React.OutSettings.OutputFile == JQ[i].React.OutSettings.OutputFile &&
			JQ[i+1].React.OutSettings.OutputFormat == JQ[i].React.OutSettings.OutputFormat {
			toWhereShadow &= ^output.OutToFile
		}
		output.FinishOutput(toWhereShadow)

		output.AddJobCounter()
		output.Logf(toWhereShadow, "Job#%d completed, time %v", i, timeLapsed)
	}
	output.Log(toWhereShadow, "All jobs completed")
	output.WaitForScreenQuit()
}

func GetCurrentJob() *fuzzTypes.Fuzz {
	return fuzzCommon.GetCurFuzz()
}

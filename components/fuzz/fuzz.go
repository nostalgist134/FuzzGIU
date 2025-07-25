package fuzz

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageReact"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageSend"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"github.com/nostalgist134/FuzzGIU/components/wp"
	"net/http"
	"os"
	"sync"
	"time"
)

type JobQueue []*fuzzTypes.Fuzz

type safeSlice struct {
	l   []*fuzzTypes.Reaction
	end int
	mu  sync.Mutex
}

var JQ JobQueue = make([]*fuzzTypes.Fuzz, 0)
var SendMetaPool = sync.Pool{
	New: func() any { return new(fuzzTypes.SendMeta) },
}

// Wp 协程池指针
var Wp *wp.WorkerPool

// handleReaction 根据fuzz设置处理反应
func handleReaction(r *fuzzTypes.Reaction, fuzz1 *fuzzTypes.Fuzz, reactPlugin plugin.Plugin) bool {
	stopJob := false
	if r.Flag&fuzzTypes.ReactFlagAddJob != 0 && r.NewJob != nil {
		JQ.AddJob(r.NewJob)
		output.SetJobCounter(output.GetCounterSingle(3) + 1)
	}
	if r.Flag&fuzzTypes.ReactFlagStopJob != 0 {
		output.Log("job stopped by react", common.OutputToWhere)
		stopJob = true
	}
	if r.Flag&fuzzTypes.ReactAddReq != 0 && r.NewReq != nil {
		newSend := (SendMetaPool.Get()).(*fuzzTypes.SendMeta)
		newTask := func() *fuzzTypes.Reaction {
			newSend.Timeout = fuzz1.Send.Timeout
			newSend.Retry = fuzz1.Send.Retry
			newSend.RetryRegex = fuzz1.Send.RetryRegex
			newSend.RetryCode = fuzz1.Send.RetryCode
			newSend.HttpFollowRedirects = fuzz1.Send.HttpFollowRedirects
			newSend.Request = r.NewReq
			resp := stageSend.SendRequest(newSend)
			reaction := stageReact.React(fuzz1, newSend.Request, resp, reactPlugin,
				[]string{""}, []string{"added via react"}, nil)
			SendMetaPool.Put(newSend)
			// task数加1
			output.AddTaskCounter()
			return reaction
		}
		// task总数加1
		output.SetTaskCounter(output.GetCounterSingle(1) + 1)
		Wp.Submit(newTask)
	}
	if r.Flag&fuzzTypes.ReactFlagExit != 0 {
		output.FinishOutput(common.OutputToWhere)
		if common.OutputToWhere&output.OutToScreen != 0 {
			output.ScreenClose()
		}
		fmt.Println("exit by react")
		os.Exit(0)
	}
	return stopJob
}

// doFuzz 程序实际执行的函数 生成payload->预处理->分配->返回处理->输出
func doFuzz(fuzz1 *fuzzTypes.Fuzz) time.Duration {
	timeStart := time.Now()
	// 判断递归深度
	if fuzz1.React.RecursionControl.RecursionDepth > fuzz1.React.RecursionControl.MaxRecursionDepth {
		return time.Since(timeStart)
	}

	// 初始化协程池
	if Wp == nil {
		Wp = wp.New(fuzz1.Misc.PoolSize)
		Wp.Start()
	} else {
		Wp.Resize(fuzz1.Misc.PoolSize)
	}

	fuzz1 = stagePreprocess.Preprocess(fuzz1, fuzz1.Preprocess.Preprocessors)
	// 多个fuzz关键字的处理
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
				loopLen *= int64(common.GetKeywordNum(&fuzz1.Send.Request, keyword))
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
	output.SetTaskCounter(loopLen)
	output.ClearTaskCounter()
	// 任务
	var task func() *fuzzTypes.Reaction
	// req模板解析
	reqTemplate := common.ParseReqTemplate(&fuzz1.Send.Request, keywords)
	// 反应器插件
	var reactPlugin plugin.Plugin
	if fuzz1.React.Reactor != "" {
		reactPlugin = plugin.ParsePluginsStr(fuzz1.React.Reactor)[0]
	}
	// payload处理插件
	var plProcessorPlugins = make([][]plugin.Plugin, len(keywords))
	// 用于接收handleReaction标记当前任务是否结束
	jobStop := false
	for i, keyword := range keywords {
		plProcessorPlugins[i] = plugin.ParsePluginsStr(fuzz1.Preprocess.PlTemp[keyword].Processors)
	}
	// 主循环
	for i := int64(0); i < loopLen; i++ {
		send := (SendMetaPool.Get()).(*fuzzTypes.SendMeta)
		send.Timeout = fuzz1.Send.Timeout
		send.Retry = fuzz1.Send.Retry
		send.RetryRegex = fuzz1.Send.RetryRegex
		send.RetryCode = fuzz1.Send.RetryCode
		send.HttpFollowRedirects = fuzz1.Send.HttpFollowRedirects

		payloadEachKeyword := make([]string, 0)
		curInd := int64(len(fuzz1.Preprocess.PlTemp[keywords[0]].PlList))
		send.Proxy = ""
		// 代理轮询
		if len(fuzz1.Send.Proxies) > 0 {
			send.Proxy = fuzz1.Send.Proxies[i%int64(len(fuzz1.Send.Proxies))]
		}
		if fuzz1.Preprocess.Mode == "clusterbomb" && fuzz1.React.RecursionControl.MaxRecursionDepth <= 0 {
			curInd = i
		}
		if fuzz1.Preprocess.Mode != "sniper" && fuzz1.React.RecursionControl.MaxRecursionDepth <= 0 {
			for j := 0; j < len(keywords); j++ { // 遍历keywords列表，根据i选出每个关键字对应的payload
				switch fuzz1.Preprocess.Mode {
				// clusterbomb模式，遍历所有的payload组合
				case "clusterbomb":
					d := int64(len(fuzz1.Preprocess.PlTemp[keywords[len(keywords)-j-1]].PlList))
					r := curInd % d
					curInd /= d
					payloadEachKeyword = append([]string{fuzz1.Preprocess.PlTemp[keywords[len(keywords)-j-1]].PlList[r]},
						payloadEachKeyword...)
				// pitchfork模式：每个关键字使用一样的payload下标
				case "pitchfork":
					payloadEachKeyword = append(payloadEachKeyword, fuzz1.Preprocess.PlTemp[keywords[j]].PlList[i])
				// pitchfork-cycle模式：每次i循环下标都同步更新1，但payload列表到尾部后会从头再次开始
				case "pitchfork-cycle":
					payloadEachKeyword = append(payloadEachKeyword, fuzz1.Preprocess.PlTemp[keywords[j]].PlList[i%
						int64(len(fuzz1.Preprocess.PlTemp[keywords[j]].PlList))])
				}
			}
			task = func() *fuzzTypes.Reaction {
				processedPayloads := make([]string, len(payloadEachKeyword))
				for j, plugins := range plProcessorPlugins {
					processedPayloads[j] = stagePreprocess.PayloadProcessor(payloadEachKeyword[j], plugins)
				}
				send.Request = common.ReplacePayloadsByTemplate(reqTemplate, processedPayloads, -1)
				resp := stageSend.SendRequest(send)
				reaction := stageReact.React(fuzz1, send.Request, resp, reactPlugin,
					keywords, processedPayloads, nil)
				SendMetaPool.Put(send)
				output.AddTaskCounter()
				return reaction
			}
		} else { // sniper模式或者递归模式
			keyword := keywords[0]
			payload := fuzz1.Preprocess.PlTemp[keyword].PlList[i%curInd]
			task = func() *fuzzTypes.Reaction {
				processedPayload := payload
				processedPayload = stagePreprocess.PayloadProcessor(processedPayload, plProcessorPlugins[0])
				var recPos []int = nil
				// payload替换
				if fuzz1.Preprocess.Mode == "sniper" &&
					fuzz1.React.RecursionControl.RecursionDepth <= fuzz1.React.RecursionControl.MaxRecursionDepth {
					// 同时启用sniper和递归
					send.Request, recPos = common.ReplacePayloadTrackTemplate(reqTemplate, payload, int(i/curInd))
				} else if fuzz1.React.RecursionControl.RecursionDepth <= fuzz1.React.RecursionControl.MaxRecursionDepth {
					// 只启用递归
					send.Request, recPos = common.ReplacePayloadTrackTemplate(reqTemplate, payload, -1)
				} else { // 只启用sniper
					send.Request = common.ReplacePayloadsByTemplate(reqTemplate, []string{payload}, int(i/curInd))
				}
				resp := stageSend.SendRequest(send)
				reaction := stageReact.React(fuzz1, send.Request, resp, reactPlugin,
					[]string{keyword}, []string{processedPayload}, recPos)
				SendMetaPool.Put(send)
				output.AddTaskCounter()
				return reaction
			}
		}
		Wp.Submit(task)
		time.Sleep(time.Millisecond * time.Duration(fuzz1.Misc.Delay))
		// 重复100次获取结果队列中的结果
		maxTry := 8192
		for {
			if maxTry == 0 {
				break
			}
			if r := Wp.GetSingleResult(); r != nil {
				jobStop = handleReaction(r, fuzz1, reactPlugin)
				if jobStop {
					return time.Since(timeStart)
				}
			}
			maxTry--
		}
	}
	for !Wp.Wait(time.Millisecond * 10) {
		for r := Wp.GetSingleResult(); r != nil; r = Wp.GetSingleResult() {
			if handleReaction(r, fuzz1, reactPlugin) {
				return time.Since(timeStart)
			}
		}
	}
	for r := Wp.GetSingleResult(); r != nil; r = Wp.GetSingleResult() {
		if handleReaction(r, fuzz1, reactPlugin) {
			return time.Since(timeStart)
		}
	}
	return time.Since(timeStart)
}

func (jq *JobQueue) AddJob(fuzz *fuzzTypes.Fuzz) {
	*jq = append(*jq, fuzz)
}

func DoJobs() {
	output.SetJobCounter(int64(len(JQ)))
	i := 0
	toWhereShadow := int32(0)
	for ; i < len(JQ); i++ {
		// 根据OutSettings选则输出模式（termui界面、原生stdout）
		if !JQ[i].React.OutSettings.NativeStdout {
			common.OutputToWhere = output.OutToScreen
		} else {
			common.OutputToWhere = output.OutToNativeStdout
		}
		if JQ[i].React.OutSettings.OutputFile != "" {
			common.OutputToWhere |= output.OutToFile
		}
		output.InitOutput(JQ[i], common.OutputToWhere)
		timeLapsed := doFuzz(JQ[i])
		toWhereShadow = common.OutputToWhere
		// 如果下一个任务仍然使用同样文件以及同样输出格式，则不结束文件输出，追加到同一文件
		if i+1 < len(JQ) && JQ[i+1].React.OutSettings.OutputFile == JQ[i].React.OutSettings.OutputFile &&
			JQ[i+1].React.OutSettings.OutputFormat == JQ[i].React.OutSettings.OutputFormat {
			toWhereShadow &= ^output.OutToFile
		}
		output.FinishOutput(toWhereShadow)
		output.AddJobCounter()
		output.Log(fmt.Sprintf("Job#%d completed, time %v", i, timeLapsed), toWhereShadow)
	}
	output.Log("All jobs completed", toWhereShadow)
	output.WaitForScreenQuit()
}

func Debug(fuzz1 *fuzzTypes.Fuzz) {
	kw := ""
	for k, _ := range fuzz1.Preprocess.PlTemp {
		kw = k
		break
	}
	r := fuzz1.Send.Request
	t := common.ParseReqTemplate(&r, []string{kw})
	newReq, trackPos := common.ReplacePayloadTrackTemplate(t, "1milaogiu", -1)
	resp := &fuzzTypes.Resp{HttpResponse: &http.Response{StatusCode: 404}}
	fmt.Println(newReq, trackPos)
	reaction := stageReact.React(fuzz1, newReq, resp, plugin.Plugin{}, []string{}, []string{}, trackPos)
	fmt.Println(reaction.NewJob.Send.Request)
}

package fuzz

import (
	"FuzzGIU/components/fuzz/common"
	"FuzzGIU/components/fuzz/stagePreprocess"
	"FuzzGIU/components/fuzz/stageReact"
	"FuzzGIU/components/fuzz/stageSend"
	"FuzzGIU/components/fuzzTypes"
	"FuzzGIU/components/output"
	"FuzzGIU/components/plugin"
	"FuzzGIU/components/wp"
	"bytes"
	"fmt"
	"os"
	"sync"
	"time"
)

type JobQueue []fuzzTypes.Fuzz

var JQ JobQueue = make([]fuzzTypes.Fuzz, 0)
var OutputBuffer bytes.Buffer
var SendMetaPool = sync.Pool{
	New: func() interface{} { return new(fuzzTypes.SendMeta) },
}

var outputUI = false

// patchLog#9: 这两个是cachedChan使用的缓存以及mutex，原本是作为cachedChan成员的，但是作为成员的时候，一直不能正常运行，拿出来作为全局
// 变量时就可以
var cache = make([]*fuzzTypes.Reaction, 0)
var mu = &sync.Mutex{}

type cachedChan struct {
	ch <-chan *fuzzTypes.Reaction
}

func cachedChanInit(ch <-chan *fuzzTypes.Reaction) *cachedChan {
	ret := &cachedChan{
		ch: ch,
	}
	go func() {
		for {
			select {
			case reaction, ok := <-ret.ch:
				if !ok {
					return
				}
				mu.Lock()
				cache = append(cache, reaction)
				mu.Unlock()
			}
		}
	}()
	return ret
}

func (cChan *cachedChan) Get() *fuzzTypes.Reaction {
	var ret *fuzzTypes.Reaction
	mu.Lock()
	// 获取通道缓存中的第一个值
	if len(cache) == 0 {
		ret = nil
	} else {
		ret = cache[0]
		// 更新缓存
		cache = cache[1:]
	}
	mu.Unlock()
	return ret
}

// doFuzz 程序实际执行的函数 生成payload->预处理->分配->返回处理->输出
func doFuzz(fuzz *fuzzTypes.Fuzz) {
	// 判断递归深度
	if fuzz.React.RecursionControl.RecursionDepth > fuzz.React.RecursionControl.MaxRecursionDepth {
		return
	}
	/*------------------------------------ 预处理阶段 ------------------------------------*/
	Wp := wp.NewWorkerPool(fuzz.Misc.PoolSize)
	fuzz = stagePreprocess.Preprocess(fuzz, fuzz.Preprocess.Preprocessors)
	// 多个fuzz关键字的处理，有4种模式
	keywords := make([]string, 0)
	plListsLen := int64(1)
	for keyword, pt := range fuzz.Preprocess.PlTemp { // 计算长度(plListsLen)
		keywords = append(keywords, keyword)
		// sniper模式
		if fuzz.Preprocess.Mode == "sniper" || fuzz.React.RecursionControl.MaxRecursionDepth > 0 {
			// 如果采用递归扫描或者sniper模式，则只使用一个关键词
			plListsLen = int64(len(pt.PlList))
			if fuzz.Preprocess.Mode == "sniper" {
				plListsLen *= int64(common.GetKeywordNum(&fuzz.Send.Request, keyword))
			}
			break
		}
		switch fuzz.Preprocess.Mode {
		// clusterbomb模式：遍历每个关键词对应payload列表的所有组合
		case "clusterbomb":
			plListsLen *= int64(len(pt.PlList))
		// pitchfork模式：每个关键字的payload列表在遍历时下标会同步替换，因此以最小的payload列表为准
		case "pitchfork":
			if int64(len(pt.PlList)) < plListsLen {
				plListsLen = int64(len(pt.PlList))
			}
		// pitchfork-cycle模式：以最大的payload列表为准，每个关键字的payload列表在遍历时下标会同步替换，如果
		// 一个payload列表遍历完了，会从这个列表第一个元素重新开始
		case "pitchfork-cycle":
			if int64(len(pt.PlList)) > plListsLen {
				plListsLen = int64(len(pt.PlList))
			}
		default:
			fmt.Println("unsupported mode", fuzz.Preprocess.Mode)
			os.Exit(1)
		}
	}
	output.SetCounter(int(plListsLen), -1)
	var task func() *fuzzTypes.Reaction // 任务

	sendMetaList := make([]*fuzzTypes.SendMeta, 0)
	defer func() { //回收sendMeta
		for _, sendMeta := range sendMetaList {
			SendMetaPool.Put(sendMeta)
		}
	}()

	reqTemplate := common.ParseReqTemplate(&fuzz.Send.Request, keywords) // req模板解析
	wpIsRunning := false                                                 // 协程池运行状态
	// 协程池结果管道以及缓存管道
	var resultsChan <-chan *fuzzTypes.Reaction
	var cChan *cachedChan
	// 反应器插件
	var reactPlugin plugin.Plugin
	if fuzz.React.Reactor != "" {
		reactPlugin = plugin.ParsePluginsStr(fuzz.React.Reactor)[0]
	}
	// 主循环
	for i := int64(0); i < plListsLen; i++ {
		sendMeta := (SendMetaPool.Get()).(*fuzzTypes.SendMeta)
		sendMetaList = append(sendMetaList, sendMeta)
		sendMeta.Timeout = fuzz.Send.Timeout
		sendMeta.Retry = fuzz.Send.Retry
		sendMeta.RetryRegex = fuzz.Send.RetryRegex
		sendMeta.RetryCode = fuzz.Send.RetryCode
		sendMeta.HttpFollowRedirects = fuzz.Send.HttpFollowRedirects

		payloadEachKeyword := make([]string, 0)
		// 由于payloadGenerator生成的逻辑是即使为空也会返回一个空字符串，所以可以不用判断curInd是否为0
		curInd := int64(len(fuzz.Preprocess.PlTemp[keywords[0]].PlList))
		sendMeta.Proxy = ""
		// 目前设计为如果代理池中多于一个代理则会按顺序循环使用
		if len(fuzz.Send.Proxies) > 0 {
			sendMeta.Proxy = fuzz.Send.Proxies[i%int64(len(fuzz.Send.Proxies))]
		}
		if fuzz.Preprocess.Mode == "clusterbomb" {
			curInd = i
		}
		if fuzz.Preprocess.Mode != "sniper" && fuzz.React.RecursionControl.MaxRecursionDepth <= 0 {
			for j := 0; j < len(keywords); j++ { // 遍历keywords列表，根据i选出每个关键字对应的payload
				switch fuzz.Preprocess.Mode {
				// clusterbomb模式，遍历所有的payload组合
				case "clusterbomb":
					// patchLog#8: 修复了clusterbomb模式的逻辑
					d := int64(len(fuzz.Preprocess.PlTemp[keywords[len(keywords)-j-1]].PlList))
					r := curInd % d
					curInd /= d
					payloadEachKeyword = append([]string{fuzz.Preprocess.PlTemp[keywords[len(keywords)-j-1]].PlList[r]},
						payloadEachKeyword...)
				// pitchfork模式：每个关键字使用一样的payload下标
				case "pitchfork":
					payloadEachKeyword = append(payloadEachKeyword, fuzz.Preprocess.PlTemp[keywords[j]].PlList[i])
				// pitchfork-cycle模式：每次i循环下标都同步更新1，但payload列表到尾部后会从头再次开始
				case "pitchfork-cycle":
					payloadEachKeyword = append(payloadEachKeyword,
						fuzz.Preprocess.PlTemp[keywords[j]].PlList[i%int64(len(fuzz.Preprocess.PlTemp[keywords[j]].PlList))])
				}
			}
			sendMeta.Request = common.ReplacePayloadsByTemplate(reqTemplate, payloadEachKeyword, -1)
			task = func() *fuzzTypes.Reaction {
				resp := stageSend.SendRequest(sendMeta)
				return stageReact.React(fuzz, sendMeta.Request, resp, reactPlugin,
					keywords, payloadEachKeyword, nil)
			}
		} else { // sniper模式或者递归模式
			keyword := keywords[0]
			payload := fuzz.Preprocess.PlTemp[keyword].PlList[i%curInd]
			var recursionPos []int = nil
			// payload替换
			if fuzz.Preprocess.Mode == "sniper" &&
				fuzz.React.RecursionControl.RecursionDepth < fuzz.React.RecursionControl.MaxRecursionDepth {
				// 同时启用sniper和递归
				sendMeta.Request = common.ReplacePayloadsByTemplate(reqTemplate, []string{keyword}, int(i/curInd))
				oldRequest := sendMeta.Request
				// todo: 这句还没想好怎么改，因为第一次replace会导致req发生变化，就不能用旧的模板了，所以这里只能用原先的函数
				sendMeta.Request, recursionPos = common.ReplacePayloadTrack(sendMeta.Request, keyword, payload)
				common.PutReq(oldRequest)
			} else if fuzz.React.RecursionControl.RecursionDepth < fuzz.React.RecursionControl.MaxRecursionDepth {
				// 只启用递归
				sendMeta.Request, recursionPos = common.ReplacePayloadTrack(&fuzz.Send.Request, keyword, payload)
			} else { // 只启用sniper
				sendMeta.Request = common.ReplacePayloadsByTemplate(reqTemplate, []string{payload}, int(i/curInd))
			}
			task = func() *fuzzTypes.Reaction {
				resp := stageSend.SendRequest(sendMeta)
				return stageReact.React(fuzz, sendMeta.Request, resp, reactPlugin,
					[]string{keyword}, []string{payload}, recursionPos)
			}
		}
		/*------------------------------------ 发送阶段 ------------------------------------*/
		Wp.Submit(task)
		if !wpIsRunning {
			wpIsRunning = true
			Wp.Start()
		}
		/*------------------------------------ 响应阶段 ------------------------------------*/
		if resultsChan == nil {
			resultsChan = Wp.GetResult()
			cChan = cachedChanInit(resultsChan)
		}
		tries := 10
		// 重复10次获取结果队列中的结果
		for ; tries > 0; tries-- {
			if cChan == nil {
				break
			}
			r := cChan.Get()
			if r == nil {
				continue
			}
			output.UpdateCounterTask()
			if r.Flag&fuzzTypes.ReactFlagAddJob != 0 {
				JQ.AddJob(r.NewJob)
				output.UpdateTotalJob()
			}
			if r.Flag&fuzzTypes.ReactFlagOutput != 0 {
				output.UpdateOutput(r.Output.Msg)
			}
			if r.Flag&fuzzTypes.ReactFlagStopJob != 0 {
				Wp.Stop()
				return
			}
			if r.Flag&fuzzTypes.ReactFlagExit != 0 {
				fmt.Println("Now exiting...")
				os.Exit(0)
			}
			break
		}
		time.Sleep(time.Millisecond * time.Duration(fuzz.Misc.Delay))
	}
	Wp.Wait()
	/*
		patchLog#1: 有时主循环中的响应阶段没办法把协程队列中的所有结果取干净，这是因为getResults是非阻塞的，
		队列为空时返回nil，此循环退出条件也为nil，但是有些任务不会那么快结束，就会漏，解决方法是主循环结束后使
		用wp.wait等待所有协程完成任务后，再获取
	*/
	if cChan != nil {
		for {
			r := cChan.Get()
			if r == nil {
				break
			}
			output.UpdateCounterTask()
			if r.Flag&fuzzTypes.ReactFlagAddJob != 0 {
				JQ.AddJob(r.NewJob)
				output.UpdateTotalJob()
			}
			if r.Flag&fuzzTypes.ReactFlagOutput != 0 {
				output.UpdateOutput(r.Output.Msg)
			}
			if r.Flag&fuzzTypes.ReactFlagStopJob != 0 {
				Wp.Stop()
				return
			}
			if r.Flag&fuzzTypes.ReactFlagExit != 0 {
				fmt.Println("Now exiting...")
				os.Exit(0)
			}
		}
	}
}

func (jq *JobQueue) AddJob(fuzz *fuzzTypes.Fuzz) {
	*jq = append(*jq, *fuzz)
}

func DoJobs(outputFile string) {
	file, err := os.Open(outputFile)
	if err == nil {
		defer file.Close()
	}
	output.SetCounter(-1, len(JQ))
	for i := 0; i < len(JQ); i++ {
		doFuzz(&JQ[i])
		output.UpdateCounterJob()
		if file != nil {
			file.Write(OutputBuffer.Bytes())
			OutputBuffer.Reset()
		}
	}
	output.Finish()
}

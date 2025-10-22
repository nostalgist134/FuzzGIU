package fuzz

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageReact"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageSend"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"github.com/nostalgist134/FuzzGIU/components/tmplReplace"
)

// taskMultiKeyword 多关键字fuzz使用的执行函数
func taskMultiKeyword(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	payloads := c.Payloads
	job := c.JobCtx.Job
	i := c.IterInd
	plProc := c.PlProc
	repTmpl := c.RepTmpl
	keywords := c.Keywords
	uScheme := c.USchemeCache

	defer resourcePool.StringSlices.Put(payloads)

	// nostalgist134他妈是不是脑袋有问题，还专门写一个sync.Pool来管理sendMeta，直接栈分配不就行了
	send := fuzzTypes.SendMeta{
		Retry:               job.Send.Retry,
		HttpFollowRedirects: job.Send.HttpFollowRedirects,
		RetryCode:           job.Send.RetryCode,
		RetryRegex:          job.Send.RetryRegex,
		Timeout:             job.Send.Timeout,
	}

	// 代理轮询
	if len(job.Send.Proxies) > 0 {
		send.Proxy = job.Send.Proxies[i%len(job.Send.Proxies)]
	}

	var cacheId int32

	processedPayloads := resourcePool.StringSlices.Get(len(payloads))
	defer resourcePool.StringSlices.Put(processedPayloads)

	for j, eachPlProc := range plProc {
		processedPayloads[j] = stagePreprocess.PayloadProcessor(c.JobCtx.OutputCtx, payloads[j], eachPlProc)
	}

	send.Request, cacheId = repTmpl.Replace(processedPayloads, -1)
	defer tmplReplace.ReleaseReqCache(cacheId)

	send.Request.HttpSpec.ForceHttps = job.Preprocess.ReqTemplate.HttpSpec.ForceHttps

	resp := stageSend.SendRequest(&send, uScheme)
	reaction := stageReact.React(c.JobCtx, send.Request, resp, keywords, processedPayloads, nil)

	return reaction
}

// taskSingleKeyword 单关键字（sniper模式或者递归模式）使用的任务执行函数
func taskSingleKeyword(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	job := c.JobCtx.Job
	i := c.IterInd
	payloads := c.Payloads
	plProc := c.PlProc
	repTmpl := c.RepTmpl
	snipLen := c.SnipLen
	uScheme := c.USchemeCache
	keywords := c.Keywords

	send := fuzzTypes.SendMeta{
		Retry:               job.Send.Retry,
		HttpFollowRedirects: job.Send.HttpFollowRedirects,
		RetryCode:           job.Send.RetryCode,
		RetryRegex:          job.Send.RetryRegex,
		Timeout:             job.Send.Timeout,
	}

	// 代理轮询
	if len(job.Send.Proxies) > 0 {
		send.Proxy = job.Send.Proxies[i%len(job.Send.Proxies)]
	}

	processedPayload := payloads[0]
	payload := payloads[0]
	processedPayload = stagePreprocess.PayloadProcessor(c.JobCtx.OutputCtx, processedPayload, plProc[0])

	var recPos []int
	var cacheId int32

	// payload替换
	if job.Control.IterCtrl.Iterator.Name == "sniper" && // 同时启用sniper和递归
		job.React.RecursionControl.RecursionDepth <= job.React.RecursionControl.MaxRecursionDepth {
		send.Request, recPos, cacheId = repTmpl.ReplaceTrack(payload, i/snipLen)
	} else if job.React.RecursionControl.RecursionDepth <=
		job.React.RecursionControl.MaxRecursionDepth { // 只启用递归
		send.Request, recPos, cacheId = repTmpl.ReplaceTrack(payload, -1)
	} else { // 只启用sniper
		send.Request, cacheId = repTmpl.Replace([]string{payload}, i/snipLen)
	}
	defer tmplReplace.ReleaseReqCache(cacheId)
	defer resourcePool.IntSlices.Put(recPos) // resourcesPool.Put(nil)会自动被忽略

	send.Request.HttpSpec.ForceHttps = job.Preprocess.ReqTemplate.HttpSpec.ForceHttps

	resp := stageSend.SendRequest(&send, uScheme)

	reaction := stageReact.React(c.JobCtx, send.Request, resp, keywords,
		[]string{processedPayload}, recPos)

	return reaction
}

// taskNoKeywords 用于没有包含payload信息的任务的执行，目前只有handleReaction时发现需要添加新请求时，才使用此函数
func taskNoKeywords(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	job := c.JobCtx.Job
	r := c.ViaReaction
	k, p := stageReact.GetReactTraceInfo(r)

	send := fuzzTypes.SendMeta{
		Request:             r.NewReq,
		Retry:               job.Send.Retry,
		HttpFollowRedirects: job.Send.HttpFollowRedirects,
		RetryCode:           job.Send.RetryCode,
		RetryRegex:          job.Send.RetryRegex,
		Timeout:             job.Send.Timeout,
	}

	resp := stageSend.SendRequest(&send, "")
	reaction := stageReact.React(c.JobCtx, send.Request, resp, []string{""},
		[]string{fmt.Sprintf("add via react by %s:%s", k, p)}, nil)
	return reaction
}

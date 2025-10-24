package fuzz

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageDoReq"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageReact"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"github.com/nostalgist134/FuzzGIU/components/tmplReplace"
)

// taskMultiKeyword 多关键字fuzz使用的执行函数
func taskMultiKeyword(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	defer resourcePool.PutTaskCtx(c)

	payloads := c.Payloads
	job := c.JobCtx.Job
	i := c.IterInd
	plProc := c.PlProc
	repTmpl := c.RepTmpl
	keywords := c.Keywords
	uScheme := c.USchemeCache

	defer resourcePool.StringSlices.Put(payloads)

	rc := resourcePool.GetReqCtx()
	defer resourcePool.PutReqCtx(rc)

	*rc = fuzzTypes.RequestCtx{
		Retry:               job.Send.Retry,
		HttpFollowRedirects: job.Send.HttpFollowRedirects,
		RetryCode:           job.Send.RetryCode,
		RetryRegex:          job.Send.RetryRegex,
		Timeout:             job.Send.Timeout,
	}

	// 代理轮询
	if len(job.Send.Proxies) > 0 {
		rc.Proxy = job.Send.Proxies[i%len(job.Send.Proxies)]
	}

	var cacheId int32

	processedPayloads := resourcePool.StringSlices.Get(len(payloads))
	defer resourcePool.StringSlices.Put(processedPayloads)

	for j, eachPlProc := range plProc {
		processedPayloads[j] = stagePreprocess.PayloadProcessor(c.JobCtx.OutputCtx, payloads[j], eachPlProc)
	}

	rc.Request, cacheId = repTmpl.Replace(processedPayloads, -1)
	defer tmplReplace.ReleaseReqCache(cacheId)

	rc.Request.HttpSpec.ForceHttps = job.Preprocess.ReqTemplate.HttpSpec.ForceHttps

	resp := stageDoReq.DoRequest(rc, uScheme)
	reaction := stageReact.React(c.JobCtx, rc.Request, resp, keywords, processedPayloads, nil)

	return reaction
}

// taskSingleKeyword 单关键字（sniper模式或者递归模式）使用的任务执行函数（单关键字的执行函数居然比多关键字的还复杂，笑死）
func taskSingleKeyword(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	defer resourcePool.PutTaskCtx(c)

	job := c.JobCtx.Job
	i := c.IterInd
	payloads := c.Payloads
	plProc := c.PlProc
	repTmpl := c.RepTmpl
	snipLen := c.SnipLen
	uScheme := c.USchemeCache
	keywords := c.Keywords

	rc := resourcePool.GetReqCtx()
	defer resourcePool.PutReqCtx(rc)

	*rc = fuzzTypes.RequestCtx{
		Retry:               job.Send.Retry,
		HttpFollowRedirects: job.Send.HttpFollowRedirects,
		RetryCode:           job.Send.RetryCode,
		RetryRegex:          job.Send.RetryRegex,
		Timeout:             job.Send.Timeout,
	}

	// 代理轮询
	if len(job.Send.Proxies) > 0 {
		rc.Proxy = job.Send.Proxies[i%len(job.Send.Proxies)]
	}

	processedPayload := payloads[0]
	payload := payloads[0]

	processedPayloads := resourcePool.StringSlices.Get(1)
	defer resourcePool.StringSlices.Put(processedPayloads)

	processedPayload = stagePreprocess.PayloadProcessor(c.JobCtx.OutputCtx, processedPayload, plProc[0])
	processedPayloads[0] = processedPayload

	var recPos []int
	var cacheId int32

	tmp := resourcePool.StringSlices.Get(1)
	defer resourcePool.StringSlices.Put(tmp)

	tmp[0] = payload

	// payload替换
	if job.Control.IterCtrl.Iterator.Name == "sniper" && // 同时启用sniper和递归
		job.React.RecursionControl.RecursionDepth <= job.React.RecursionControl.MaxRecursionDepth {
		rc.Request, recPos, cacheId = repTmpl.ReplaceTrack(payload, i/snipLen)
	} else if job.React.RecursionControl.RecursionDepth <=
		job.React.RecursionControl.MaxRecursionDepth { // 只启用递归
		rc.Request, recPos, cacheId = repTmpl.ReplaceTrack(payload, -1)
	} else { // 只启用sniper
		rc.Request, cacheId = repTmpl.Replace(tmp, i/snipLen)
	}
	defer tmplReplace.ReleaseReqCache(cacheId)
	defer resourcePool.IntSlices.Put(recPos)

	rc.Request.HttpSpec.ForceHttps = job.Preprocess.ReqTemplate.HttpSpec.ForceHttps

	resp := stageDoReq.DoRequest(rc, uScheme)

	reaction := stageReact.React(c.JobCtx, rc.Request, resp, keywords, processedPayloads, recPos)

	return reaction
}

// taskNoKeywords 用于没有包含payload信息的任务的执行，目前只有handleReaction时发现需要添加新请求时，才使用此函数
func taskNoKeywords(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	defer resourcePool.PutTaskCtx(c)

	job := c.JobCtx.Job
	r := c.ViaReaction
	k, p := stageReact.GetReactTraceInfo(r)

	rc := resourcePool.GetReqCtx()
	defer resourcePool.PutReqCtx(rc)

	*rc = fuzzTypes.RequestCtx{
		Retry:               job.Send.Retry,
		HttpFollowRedirects: job.Send.HttpFollowRedirects,
		RetryCode:           job.Send.RetryCode,
		RetryRegex:          job.Send.RetryRegex,
		Timeout:             job.Send.Timeout,
	}

	tmp := resourcePool.StringSlices.Get(1)
	defer resourcePool.StringSlices.Put(tmp)

	addedVia := fmt.Sprintf("add via react by %s:%s", k, p)
	tmp[0] = addedVia

	resp := stageDoReq.DoRequest(rc, "")
	reaction := stageReact.React(c.JobCtx, rc.Request, resp, []string{""},
		tmp, nil)
	return reaction
}

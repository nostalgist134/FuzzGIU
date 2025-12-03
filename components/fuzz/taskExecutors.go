package fuzz

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageReact"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageRequest"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"github.com/nostalgist134/FuzzGIU/components/tmplReplace"
	"math/rand/v2"
)

// taskMultiKeyword 多关键字fuzz使用的执行函数
func taskMultiKeyword(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	defer fuzzCtx.PutTaskCtx(c)

	payloads := c.Payloads
	job := c.JobCtx.Job
	i := c.IterInd
	plProc := c.PlProc
	repTmpl := c.RepTmpl
	keywords := c.Keywords
	uScheme := c.USchemeCache

	defer c.JobCtx.OutputCtx.Counter.Complete(counter.CntrTask)

	defer resourcePool.StringSlices.Put(payloads)

	rc := resourcePool.GetReqCtx()
	defer resourcePool.PutReqCtx(rc)

	*rc = fuzzTypes.RequestCtx{
		Retry:               job.Request.Retry,
		HttpFollowRedirects: job.Request.HttpFollowRedirects,
		RetryCodes:          job.Request.RetryCodes,
		RetryRegex:          job.Request.RetryRegex,
		Timeout:             job.Request.Timeout,
	}

	// 代理轮询
	if len(job.Request.Proxies) > 0 {
		rc.Proxy = job.Request.Proxies[i%len(job.Request.Proxies)]
	}

	var cacheId int32

	for j, eachPlProc := range plProc {
		payloads[j] = stagePreprocess.PayloadProcessor(c.JobCtx.OutputCtx, payloads[j], eachPlProc)
	}

	rc.Request, cacheId = repTmpl.Replace(payloads, -1)
	defer tmplReplace.ReleaseReqCache(cacheId)

	rc.Request.HttpSpec.ForceHttps = job.Preprocess.ReqTemplate.HttpSpec.ForceHttps
	rc.Request.HttpSpec.RandomAgent = job.Preprocess.ReqTemplate.HttpSpec.RandomAgent

	resp := stageRequest.DoRequest(rc, uScheme)
	reaction := stageReact.React(c.JobCtx, rc.Request, resp, keywords, payloads, nil)

	return reaction
}

// taskSingleKeyword 单关键字（sniper模式或者递归模式）使用的任务执行函数（单关键字的执行函数居然比多关键字的还复杂，笑死）
func taskSingleKeyword(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	defer fuzzCtx.PutTaskCtx(c)

	defer c.JobCtx.OutputCtx.Counter.Complete(counter.CntrTask)

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

	defer resourcePool.StringSlices.Put(payloads)

	*rc = fuzzTypes.RequestCtx{
		Retry:               job.Request.Retry,
		HttpFollowRedirects: job.Request.HttpFollowRedirects,
		RetryCodes:          job.Request.RetryCodes,
		RetryRegex:          job.Request.RetryRegex,
		Timeout:             job.Request.Timeout,
	}

	// 代理轮询
	if len(job.Request.Proxies) > 0 {
		rc.Proxy = job.Request.Proxies[i%len(job.Request.Proxies)]
	}

	payloads[0] = stagePreprocess.PayloadProcessor(c.JobCtx.OutputCtx, payloads[0], plProc[0])

	var recPos []int
	var cacheId int32

	// payload替换
	if job.Control.IterCtrl.Iterator.Name == "sniper" && // 同时启用sniper和递归
		job.React.RecursionControl.RecursionDepth <= job.React.RecursionControl.MaxRecursionDepth {
		rc.Request, recPos, cacheId = repTmpl.ReplaceTrack(payloads[0], i/snipLen)
	} else if job.React.RecursionControl.RecursionDepth <=
		job.React.RecursionControl.MaxRecursionDepth { // 只启用递归
		rc.Request, recPos, cacheId = repTmpl.ReplaceTrack(payloads[0], -1)
	} else { // 只启用sniper
		rc.Request, cacheId = repTmpl.Replace(payloads, i/snipLen)
	}
	defer tmplReplace.ReleaseReqCache(cacheId)
	defer resourcePool.IntSlices.Put(recPos)

	rc.Request.HttpSpec.ForceHttps = job.Preprocess.ReqTemplate.HttpSpec.ForceHttps
	rc.Request.HttpSpec.RandomAgent = job.Preprocess.ReqTemplate.HttpSpec.RandomAgent

	resp := stageRequest.DoRequest(rc, uScheme)

	reaction := stageReact.React(c.JobCtx, rc.Request, resp, keywords, payloads, recPos)

	return reaction
}

// taskNoKeywords 用于没有包含payload信息的任务的执行，目前只有handleReaction时发现需要添加新请求时，才使用此函数
func taskNoKeywords(c *fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	defer fuzzCtx.PutTaskCtx(c)

	defer c.JobCtx.OutputCtx.Counter.Complete(counter.CntrTask)

	job := c.JobCtx.Job
	r := c.ViaReaction
	k, p := stageReact.GetReactTraceInfo(r)

	rc := resourcePool.GetReqCtx()
	defer resourcePool.PutReqCtx(rc)

	*rc = fuzzTypes.RequestCtx{
		Retry:               job.Request.Retry,
		HttpFollowRedirects: job.Request.HttpFollowRedirects,
		RetryCodes:          job.Request.RetryCodes,
		RetryRegex:          job.Request.RetryRegex,
		Timeout:             job.Request.Timeout,
		Request:             r.NewReq,
	}

	// 随机代理（因为这里没有i变量，所以不能轮询，就随便选一个）
	if len(job.Request.Proxies) > 0 {
		rc.Proxy = job.Request.Proxies[rand.IntN(len(job.Request.Proxies))]
	}

	resp := stageRequest.DoRequest(rc, "")
	reaction := stageReact.React(c.JobCtx, rc.Request, resp, []string{""},
		[]string{fmt.Sprintf("add via react by %s:%s", k, p)}, nil)
	return reaction
}

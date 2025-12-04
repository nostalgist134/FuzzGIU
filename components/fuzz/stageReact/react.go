package stageReact

import (
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
)

// mergeReaction 将r2的内容归并到r1中
// r1为空的字段赋值为r2相应的字段，不为空的字段保持不变
// r1、r2的flag相或
func mergeReaction(r1 *fuzzTypes.Reaction, r2 *fuzzTypes.Reaction) {
	r1.Flag |= r2.Flag
	if r1.Output.Msg == "" {
		r1.Output.Msg = r2.Output.Msg
	}
	if r1.NewReq == nil {
		r1.NewReq = r2.NewReq
	}
	if r1.NewJob == nil {
		r1.NewJob = r2.NewJob
	}
}

// React 通过req和resp结构根据特定规则生成一个reaction结构，指示fuzz流程的下一步动作
func React(jobCtx *fuzzCtx.JobCtx, reqSend *fuzzTypes.Req, resp *fuzzTypes.Resp,
	keywordsUsed []string, payloadEachKeyword []string, recursionPos []int) *fuzzTypes.Reaction {
	defer resourcePool.PutReq(reqSend)

	reaction := resourcePool.GetReaction()

	fuzz1 := jobCtx.Job
	outCtx := jobCtx.OutputCtx

	var recursionJob *fuzzTypes.Fuzz
	/*
		递归模式通过向任务列表添加新任务完成，新任务的req结构由当前任务的React.RecursionControl控制
		1. recursionPos标记了payload替换后每个替换位置的下一个下标，通过 fuzz1.ReplacePayloadTrack 生成
		2. 根据recursionPos的标记，reqSend（newReq为将关键词替换为payload后的请求）中插入关键词
		3. reaction.Flag中标记AddJob，并设置newJob=recursionJob
	*/
	// 递归模式添加新任务
	if fuzz1.React.RecursionControl.RecursionDepth < fuzz1.React.RecursionControl.MaxRecursionDepth &&
		(resp.HttpResponse != nil &&
			fuzz1.React.RecursionControl.StatCodes.Contains(resp.HttpResponse.StatusCode) ||
			common.RegexMatch(resp.RawResponse, fuzz1.React.RecursionControl.Regex)) && recursionPos != nil {
		recursionJob = deriveRecursionJob(fuzz1, reqSend, recursionPos)
	}

	// 添加递归任务
	if recursionJob != nil {
		reaction.Flag |= fuzzTypes.ReactAddJob
		reaction.NewJob = recursionJob
	}

	// 决定是否过滤，自定义reactor若没有标识响应是否会被过滤，根据Matcher和Filter确定
	filtered := fuzz1.React.Filter.MatchResponse(resp)
	matched := fuzz1.React.Matcher.MatchResponse(resp)
	if filtered {
		reaction.Flag |= fuzzTypes.ReactFiltered
	}
	if matched {
		reaction.Flag |= fuzzTypes.ReactMatch
	}
	// 仅当没被标记为过滤，且被标记为匹配，或者出错时输出
	if !filtered && matched || !fuzz1.React.IgnoreError && resp.ErrMsg != "" {
		reaction.Flag |= fuzzTypes.ReactOutput
	}

	if resp.ErrMsg != "" {
		jobCtx.OutputCtx.Counter.Add(counter.CntrErrors, counter.FieldCompleted, 1)
	}

	// reactor插件调用
	if fuzz1.React.Reactor.Name != "" {
		pluginReaction := plugin.React(fuzz1.React.Reactor, reqSend, resp)
		if pluginReaction.Flag&fuzzTypes.ReactMerge != 0 { // 将默认响应逻辑产生的reaction归并
			mergeReaction(pluginReaction, reaction)
		} else {
			resourcePool.PutReaction(reaction)
			reaction = pluginReaction
		}
	}

	// 生成并输出消息
	if reaction.Flag&fuzzTypes.ReactOutput != 0 {
		o := output.GetOutputObj()
		defer output.PutOutputObj(o)
		o.Msg = reaction.Output.Msg
		if !reaction.Output.Overwrite {
			o.Keywords = keywordsUsed
			o.Payloads = payloadEachKeyword
			o.Request = reqSend
			o.Response = resp
		}
		outCtx.Output(o)
	}
	// 添加新单个请求的reaction，在输出消息后添加追溯信息(keyword:payload对)，易于追踪，由于消息已经输出，所以改Msg段没问题
	if reaction.Flag&fuzzTypes.ReactAddReq != 0 || reaction.Flag&fuzzTypes.ReactAddJob != 0 {
		AppendReactTraceInfo(reaction, keywordsUsed, payloadEachKeyword)
	}
	return reaction
}

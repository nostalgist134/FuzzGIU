package stageReact

import (
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"strings"
)

// todo: 更新递归任务的生成逻辑，添加对req.Fields的支持
// 	最好把递归任务的生成单独弄一个函数出来（已完成，deriveRecursionJob）

// infoMarker 用来标识payload信息出现的位置
var infoMarker = common.RandMarker()

func valInRanges(v int, ranges []fuzzTypes.Range) bool {
	for _, r := range ranges {
		if v <= r.Upper && v >= r.Lower {
			return true
		}
	}
	return false
}

// matchResponse 将响应与match成员进行匹配
func matchResponse(resp *fuzzTypes.Resp, m *fuzzTypes.Match) bool {
	if len(m.Size) == 0 && len(m.Words) == 0 && len(m.Code) == 0 && len(m.Lines) == 0 &&
		len(m.Regex) == 0 && m.Time.Upper == m.Time.Lower {
		return false
	}
	whenToRet := false
	if m.Mode == "or" {
		whenToRet = true
	}
	if len(m.Size) != 0 && valInRanges(resp.Size, m.Size) == whenToRet {
		return whenToRet
	}
	if len(m.Words) != 0 && valInRanges(resp.Words, m.Words) == whenToRet {
		return whenToRet
	}
	if len(m.Code) != 0 && resp.HttpResponse != nil && valInRanges(resp.HttpResponse.StatusCode, m.Code) == whenToRet {
		return whenToRet
	}
	if len(m.Lines) != 0 && valInRanges(resp.Lines, m.Lines) == whenToRet {
		return whenToRet
	}
	if len(m.Regex) != 0 && common.RegexMatch(resp.RawResponse, m.Regex) == whenToRet {
		return whenToRet
	}
	if m.Time.Upper != m.Time.Lower &&
		(resp.ResponseTime < m.Time.Upper && resp.ResponseTime >= m.Time.Lower) == whenToRet {
		return whenToRet
	}
	return !whenToRet
}

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

// React 函数
// patchLog#12: 添加了一个reactPlugin参数，现在reactor插件通过此参数调用，避免每次调用都解析插件字符串导致性能问题
// reactPlugin在doFuzz函数中通过一次plugin.ParsePluginsStr解析得到
func React(jobCtx *fuzzCtx.JobCtx, reqSend *fuzzTypes.Req, resp *fuzzTypes.Resp,
	keywordsUsed []string, payloadEachKeyword []string, recursionPos []int) *fuzzTypes.Reaction {
	defer resourcePool.PutReq(reqSend)
	reaction := resourcePool.GetNewReaction()

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
			valInRanges(resp.HttpResponse.StatusCode, fuzz1.React.RecursionControl.StatCodes) ||
			common.RegexMatch(resp.RawResponse, fuzz1.React.RecursionControl.Regex)) && recursionPos != nil {
		outCtx.LogFmtMsg("job#%d payload %s recursive, add new job", jobCtx.JobId, payloadEachKeyword[0])

		recursionJob = deriveRecursionJob(fuzz1, reqSend, recursionPos)
	}

	// reactDns调用
	if strings.Index(fuzz1.Preprocess.ReqTemplate.URL, "dns://") == 0 {
		resourcePool.PutReaction(reaction)
		reaction = reactDns(reqSend, resp)
	}

	// 添加递归任务（如果自定义reactor没有添加）
	if reaction.NewJob == nil && recursionJob != nil {
		reaction.Flag |= fuzzTypes.ReactAddJob
		reaction.NewJob = recursionJob
	}

	// 决定是否过滤，自定义reactor若没有标识响应是否会被过滤，根据Matcher和Filter确定
	if (reaction.Flag&fuzzTypes.ReactMatch == 0) && (reaction.Flag&fuzzTypes.ReactFiltered == 0) {
		filtered := matchResponse(resp, &fuzz1.React.Filter)
		matched := matchResponse(resp, &fuzz1.React.Matcher)
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
	} else if reaction.Flag&fuzzTypes.ReactMatch != 0 {
		reaction.Flag |= fuzzTypes.ReactOutput
	}

	// reactor插件调用
	if fuzz1.React.Reactor.Name != "" {
		pluginReaction := plugin.React(fuzz1.React.Reactor, reqSend, resp)
		if pluginReaction.Flag&fuzzTypes.ReactMerge != 0 {
			mergeReaction(pluginReaction, reaction)
		} else {
			resourcePool.PutReaction(reaction)
			reaction = pluginReaction
		}
	}

	o := output.OutObj{Msg: reaction.Output.Msg}
	// 生成并输出消息
	if reaction.Flag&fuzzTypes.ReactOutput != 0 {
		if !reaction.Output.Overwrite {
			o.Keywords = keywordsUsed
			o.Payloads = payloadEachKeyword
			o.Request = reqSend
			o.Response = resp
		}
		outCtx.Output(&o)
	}
	// 添加新单个请求的reaction，在输出消息后添加追溯信息(keyword:payload对)，易于追踪
	if reaction.Flag&fuzzTypes.ReactAddReq != 0 || reaction.Flag&fuzzTypes.ReactAddJob != 0 {
		sb := strings.Builder{}
		sb.WriteByte('\n')
		// 写入infoMarker，避免与原先的信息冲突，InfoMarker是随机生成的12位长字符串
		sb.WriteString(infoMarker)
		for i, k := range keywordsUsed {
			sb.WriteString(k)
			sb.WriteString(":")
			sb.WriteString(payloadEachKeyword[i])
			if i != len(keywordsUsed)-1 {
				sb.WriteString(infoMarker)
			}
		}
		reaction.Output.Msg += sb.String()
	}
	return reaction
}

// GetReactTraceInfo 获取reaction结构中的追溯信息
func GetReactTraceInfo(reaction *fuzzTypes.Reaction) ([]string, []string) {
	markerInd := strings.Index(reaction.Output.Msg, infoMarker)
	if markerInd == -1 {
		return nil, nil
	}
	k := make([]string, 0)
	p := make([]string, 0)
	if len(reaction.Output.Msg[markerInd:]) == len(infoMarker) {
		return nil, nil
	}
	for _, kpPair := range strings.Split(reaction.Output.Msg[markerInd+len(infoMarker):], infoMarker) {
		if kpPair != "" {
			if key, payload, ok := strings.Cut(kpPair, ":"); ok {
				k = append(k, key)
				p = append(p, payload)
			}
		}
	}
	return k, p
}

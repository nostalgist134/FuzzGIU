package stageReact

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"strings"
)

// infoMarker 用来标识payload信息出现的位置
var infoMarker = common.GetRandMarker()

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
	mode := func(cond bool) bool {
		if m.Mode == "and" {
			return !cond
		}
		return cond
	}
	retVal := true
	if m.Mode == "and" {
		retVal = false
	}
	if len(m.Size) != 0 {
		if mode(valInRanges(resp.Size, m.Size)) {
			return retVal
		}
	}
	if len(m.Words) != 0 {
		if mode(valInRanges(resp.Words, m.Words)) {
			return retVal
		}
	}
	if len(m.Code) != 0 {
		if mode(resp.HttpResponse != nil && valInRanges(resp.HttpResponse.StatusCode, m.Code)) {
			return retVal
		}
	}
	if len(m.Lines) != 0 {
		if mode(valInRanges(resp.Lines, m.Lines)) {
			return retVal
		}
	}
	if len(m.Regex) != 0 {
		if mode(common.RegexMatch(resp.RawResponse, m.Regex)) {
			return retVal
		}
	}
	if m.Time.Upper != m.Time.Lower {
		if mode(resp.ResponseTime < m.Time.Upper && resp.ResponseTime >= m.Time.Lower) {
			return retVal
		}
	}
	return !retVal
}

func insertRecursionMarker(recKeyword string, splitter string,
	field string, recursionPos []int, currentPos int) (string, int) {
	sb := strings.Builder{}
	ind := 0
	for ; recursionPos[currentPos] > 0; currentPos++ {
		sb.WriteString(field[ind:recursionPos[currentPos]])
		sb.WriteString(splitter)
		sb.WriteString(recKeyword)
		ind = recursionPos[currentPos]
	}
	if -recursionPos[currentPos] <= len(field) {
		sb.WriteString(field[ind:-recursionPos[currentPos]])
		sb.WriteString(splitter)
		sb.WriteString(recKeyword)
		ind = -recursionPos[currentPos]
		if ind < len(field) {
			sb.WriteString(field[ind:])
		}
	} else {
		return field, currentPos + 1
	}
	return sb.String(), currentPos + 1
}

// React 函数
// patchLog#12: 添加了一个reactPlugin参数，现在reactor插件通过此参数调用，避免每次调用都解析插件字符串导致性能问题
// reactPlugin在doFuzz函数中通过一次plugin.ParsePluginsStr解析得到
func React(fuzz1 *fuzzTypes.Fuzz, reqSend *fuzzTypes.Req, resp *fuzzTypes.Resp,
	reactPlugin fuzzTypes.Plugin, keywordsUsed []string, payloadEachKeyword []string,
	recursionPos []int) *fuzzTypes.Reaction {
	defer common.PutReq(reqSend)
	reaction := new(fuzzTypes.Reaction)
	var recursionJob *fuzzTypes.Fuzz = nil
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
		output.Log(fmt.Sprintf("payload %s recursive, add new job", payloadEachKeyword[0]), common.OutputToWhere)
		recKeyword := fuzz1.React.RecursionControl.Keyword
		splitter := fuzz1.React.RecursionControl.Splitter
		recursionJob = common.CopyFuzz(fuzz1)
		recursionJob.Preprocess.Mode = ""
		// 递归深度=当前深度+1
		recursionJob.React.RecursionControl.RecursionDepth++
		recursionJob.Send.Request = *reqSend
		currentPos := 0
		// HttpSpec.Method
		recursionJob.Send.Request.HttpSpec.Method, currentPos = insertRecursionMarker(recKeyword, splitter,
			recursionJob.Send.Request.HttpSpec.Method, recursionPos, 0)
		// URL
		recursionJob.Send.Request.URL, currentPos = insertRecursionMarker(recKeyword, splitter,
			recursionJob.Send.Request.URL, recursionPos, currentPos)
		// HttpSpec.Version
		recursionJob.Send.Request.HttpSpec.Version, currentPos = insertRecursionMarker(recKeyword, splitter,
			recursionJob.Send.Request.HttpSpec.Version, recursionPos, currentPos)
		// HttpSpec.Headers
		for i := 0; i < len(recursionJob.Send.Request.HttpSpec.Headers); i++ {
			recursionJob.Send.Request.HttpSpec.Headers[i], currentPos = insertRecursionMarker(recKeyword, splitter,
				recursionJob.Send.Request.HttpSpec.Headers[i], recursionPos, currentPos)
		}
		// Data
		recursionJob.Send.Request.Data, _ = insertRecursionMarker(recKeyword, splitter,
			recursionJob.Send.Request.Data, recursionPos, currentPos)
	}
	// reactDns调用
	if strings.Index(fuzz1.Send.Request.URL, "dns://") == 0 {
		reaction = reactDns(reqSend, resp)
	}
	// reactor插件调用
	if fuzz1.React.Reactor != "" {
		/*reqJson, _ := json.Marshal(reqSend)
		respJson, _ := json.Marshal(resp)
		reaction = (plugin.Call(plugin.PTypeReactor, reactPlugin, reqJson, respJson)).(*fuzzTypes.Reaction)*/
		reaction = plugin.Reactor(reactPlugin, reqSend, resp)
	}
	// 添加递归任务（如果自定义reactor没有添加）
	if reaction.NewJob == nil && recursionJob != nil {
		reaction.Flag |= fuzzTypes.ReactAddJob
		reaction.NewJob = recursionJob
	}
	// 决定是否输出，自定义reactor若没有标识响应是否会被过滤，根据Matcher和Filter确定
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
		if !filtered && matched || !fuzz1.React.OutSettings.IgnoreError && resp.ErrMsg != "" {
			reaction.Flag |= fuzzTypes.ReactOutput
		}
	} else if reaction.Flag&fuzzTypes.ReactMatch != 0 {
		reaction.Flag |= fuzzTypes.ReactOutput
	}
	o := output.ObjectOutput{Msg: reaction.Output.Msg}
	// 生成并输出消息
	if reaction.Flag&fuzzTypes.ReactOutput != 0 {
		if !reaction.Output.Overwrite {
			o.Keywords = keywordsUsed
			o.Payloads = payloadEachKeyword
			o.Request = reqSend
			o.Response = resp
		}
		output.ObjOutput(&o, common.OutputToWhere)
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
			if kp := strings.Split(kpPair, ":"); len(kp) > 1 {
				k = append(k, kp[0])
				p = append(p, kp[1])
			}
		}
	}
	return k, p
}

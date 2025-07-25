package stageReact

import (
	"FuzzGIU/components/common"
	"FuzzGIU/components/fuzzTypes"
	"FuzzGIU/components/output"
	"FuzzGIU/components/plugin"
	"fmt"
	"strings"
)

func valIn(val int, slice []int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// patchLog#10: 修复了match函数在未指定match成员（成员长度为0或值无效）时仍然比较的问题
func match(resp *fuzzTypes.Resp, m *fuzzTypes.Match) bool {
	if len(m.Size) == 0 && len(m.Words) == 0 && len(m.Code) == 0 && len(m.Lines) == 0 &&
		len(m.Regex) == 0 && m.Time.UpBound == m.Time.DownBound {
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
		if mode(valIn(resp.Size, m.Size)) {
			return retVal
		}
	}
	if len(m.Words) != 0 {
		if mode(valIn(resp.Words, m.Words)) {
			return retVal
		}
	}
	if len(m.Code) != 0 {
		if mode(resp.HttpResponse != nil && valIn(resp.HttpResponse.StatusCode, m.Code)) {
			return retVal
		}
	}
	if len(m.Lines) != 0 {
		if mode(valIn(resp.Lines, m.Lines)) {
			return retVal
		}
	}
	if len(m.Regex) != 0 {
		if mode(common.RegexMatch(resp.RawResponse, m.Regex)) {
			return retVal
		}
	}
	if m.Time.UpBound != m.Time.DownBound {
		if mode(resp.ResponseTime < m.Time.UpBound && resp.ResponseTime >= m.Time.DownBound) {
			return retVal
		}
	}
	return !retVal
}

// insertRecursionMarker This function inserts a recursion marker into a given field at the specified positions
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
	reactPlugin plugin.Plugin, keywordsUsed []string, payloadEachKeyword []string,
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
		(resp.HttpResponse != nil && valIn(resp.HttpResponse.StatusCode, fuzz1.React.RecursionControl.StatCodes) ||
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
		reaction.Flag |= fuzzTypes.ReactFlagAddJob
		reaction.NewJob = recursionJob
	}
	// 决定是否输出，自定义reactor若没有标识响应是否会被过滤，根据Matcher和Filter确定
	if (reaction.Flag&fuzzTypes.ReactFlagMatch == 0) && (reaction.Flag&fuzzTypes.ReactFlagFiltered == 0) {
		filtered := match(resp, &fuzz1.React.Filter)
		matched := match(resp, &fuzz1.React.Matcher)
		if filtered {
			reaction.Flag |= fuzzTypes.ReactFlagFiltered
		}
		if matched {
			reaction.Flag |= fuzzTypes.ReactFlagMatch
		}
		// 仅当没被标记为过滤，且被标记为匹配，或者出错时输出
		if !filtered && matched || !fuzz1.React.OutSettings.IgnoreError && resp.ErrMsg != "" {
			reaction.Flag |= fuzzTypes.ReactFlagOutput
		}
	} else if reaction.Flag&fuzzTypes.ReactFlagMatch != 0 {
		reaction.Flag |= fuzzTypes.ReactFlagOutput
	}
	o := output.ObjectOutput{Msg: reaction.Output.Msg}
	// 生成并输出消息
	if reaction.Flag&fuzzTypes.ReactFlagOutput != 0 {
		if !reaction.Output.Overwrite {
			o.Keywords = keywordsUsed
			o.Payloads = payloadEachKeyword
			o.Request = reqSend
			o.Response = resp
		}
		output.ObjOutput(&o, common.OutputToWhere)
	}
	return reaction
}

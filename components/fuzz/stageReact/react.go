package stageReact

import (
	"FuzzGIU/components/fuzz/common"
	"FuzzGIU/components/fuzzTypes"
	"FuzzGIU/components/plugin"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func valIn(val int, slice []int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func valIn2(val int, slice []int) int8 {
	for _, v := range slice {
		if v == val {
			return 1
		}
	}
	return 0
}

func getFirstLine(b []byte) string {
	if len(b) == 0 || bytes.Index(b, []byte("\n")) == 0 {
		return ""
	} else if bytes.Index(b, []byte("\n")) == -1 {
		return string(b)
	} else {
		return string(b[:bytes.Index(b, []byte("\n"))])
	}
}

type matchMeta struct {
	Code  []int  `json:"code"`
	Lines []int  `json:"lines"`
	Words []int  `json:"words"`
	Size  []int  `json:"size"`
	Regex string `json:"regex"`
	Mode  string `json:"mode"`
	Time  struct {
		DownBound time.Duration `json:"down_bound"`
		UpBound   time.Duration `json:"up_bound"`
	} `json:"time"`
}

func ReactDebug(fuzz1 *fuzzTypes.Fuzz, newReq *fuzzTypes.Req, resp *fuzzTypes.Resp,
	keywordsUsed []string, payloadEachKeyword []string, recursionPos []int) *fuzzTypes.Reaction {
	fmt.Printf("[DEBUG] reacting %s\n", newReq.URL)
	ret := new(fuzzTypes.Reaction)
	ret.Output.Msg = newReq.URL
	return ret
}

// patchLog#10: 修复了match函数在未指定matchMeta成员（成员长度为0或值无效）时仍然比较的问题
func match(resp *fuzzTypes.Resp, meta matchMeta) bool {
	if len(meta.Size) == 0 && len(meta.Words) == 0 && len(meta.Code) == 0 && len(meta.Lines) == 0 &&
		len(meta.Regex) == 0 && meta.Time.UpBound == meta.Time.DownBound {
		return false
	}
	if meta.Mode == "and" {
		if len(meta.Size) != 0 {
			if !valIn(resp.Size, meta.Size) {
				return false
			}
		}
		if len(meta.Words) != 0 {
			if !valIn(resp.Lines, meta.Lines) {
				return false
			}
		}
		if len(meta.Code) != 0 {
			if resp.HttpResponse == nil || !valIn(resp.HttpResponse.StatusCode, meta.Code) {
				return false
			}
		}
		if len(meta.Lines) != 0 {
			if !valIn(resp.Lines, meta.Lines) {
				return false
			}
		}
		if len(meta.Regex) != 0 {
			if !common.RegexMatch(resp.RawResponse, meta.Regex) {
				return false
			}
		}
		if meta.Time.UpBound != meta.Time.DownBound {
			if resp.ResponseTime >= meta.Time.UpBound || resp.ResponseTime < meta.Time.DownBound {
				return false
			}
		}
		return true
	} else {
		if len(meta.Size) != 0 {
			if valIn(resp.Size, meta.Size) {
				return true
			}
		}
		if len(meta.Words) != 0 {
			if valIn(resp.Lines, meta.Lines) {
				return true
			}
		}
		if len(meta.Code) != 0 {
			if resp.HttpResponse != nil && valIn(resp.HttpResponse.StatusCode, meta.Code) {
				return true
			}
		}
		if len(meta.Lines) != 0 {
			if valIn(resp.Lines, meta.Lines) {
				return true
			}
		}
		if len(meta.Regex) != 0 {
			if common.RegexMatch(resp.RawResponse, meta.Regex) {
				return true
			}
		}
		if meta.Time.UpBound != meta.Time.DownBound {
			if resp.ResponseTime < meta.Time.UpBound && resp.ResponseTime >= meta.Time.DownBound {
				return true
			}
		}
		return false
	}
}

// React 函数
func React(fuzz1 *fuzzTypes.Fuzz, newReq *fuzzTypes.Req, resp *fuzzTypes.Resp,
	keywordsUsed []string, payloadEachKeyword []string, recursionPos []int) *fuzzTypes.Reaction {
	defer common.PutReq(newReq) // req结构在replacePayload中是由sync.pool生成的，函数结束后放回
	reaction := new(fuzzTypes.Reaction)
	if resp.RespError != nil {
		reaction.Flag |= fuzzTypes.ReactError
		reaction.Flag |= fuzzTypes.ReactFlagOutput
		reaction.Output.Msg = resp.RespError.Error()
		return reaction
	}
	// Filter/Matcher逻辑
	// matcher的优先级比filter高
	// match/filtered
	// 0 0				-> 不输出
	// 0 1				-> 不输出
	// 1 0				-> 输出
	// 1 1				-> 输出
	respFiltered := match(resp, fuzz1.React.Filter)
	respMatched := match(resp, fuzz1.React.Matcher)
	var recursionJob *fuzzTypes.Fuzz = nil
	/*
		递归模式通过向任务列表添加新任务完成，新任务的req结构由当前任务的React.RecursionControl控制
		1. recursionPos标记了payload替换后每个替换位置的下一个下标，通过 fuzz1.ReplacePayloadTrack 生成
		2. 根据recursionPos的标记，newReq（newReq为将关键词替换为payload后的请求）中插入关键词
		3. reaction.Flag中标记AddJob，并设置newJob=recursionJob
	*/
	// 递归模式添加新任务
	if fuzz1.React.RecursionControl.RecursionDepth < fuzz1.React.RecursionControl.MaxRecursionDepth &&
		(valIn(resp.HttpResponse.StatusCode, fuzz1.React.RecursionControl.StatCodes) ||
			common.RegexMatch(resp.RawResponse, fuzz1.React.RecursionControl.Regex)) && recursionPos != nil {
		recKeyword := fuzz1.React.RecursionControl.Keyword
		splitter := fuzz1.React.RecursionControl.Splitter
		recursionJob = common.CopyFuzz(fuzz1)
		recursionJob.React.RecursionControl.RecursionDepth++ // 递归深度=当前深度+1
		recursionJob.Send.Request = *newReq
		insertRecursionMarker := func(field string, recursionPos []int, currentPos int) (string, int) {
			sb := strings.Builder{}
			ind := 0
			for ; recursionPos[currentPos] > 0; currentPos++ {
				sb.WriteString(field[ind:recursionPos[currentPos]])
				sb.WriteString(splitter)
				sb.WriteString(recKeyword)
				ind = recursionPos[currentPos]
			}
			if recursionPos[currentPos] != -len(field) {
				sb.WriteString(field[ind : recursionPos[currentPos]*-1])
				sb.WriteString(splitter)
				sb.WriteString(recKeyword)
				ind = recursionPos[currentPos] * -1
				sb.WriteString(field[ind:])
			} else {
				return field, currentPos + 1
			}
			return sb.String(), currentPos + 1
		}
		currentPos := 0
		// HttpSpec.Method
		recursionJob.Send.Request.HttpSpec.Method, currentPos = insertRecursionMarker(
			recursionJob.Send.Request.HttpSpec.Method, recursionPos, 0)
		// URL
		recursionJob.Send.Request.URL, currentPos = insertRecursionMarker(
			recursionJob.Send.Request.URL, recursionPos, currentPos)
		// HttpSpec.Version
		recursionJob.Send.Request.HttpSpec.Version, currentPos = insertRecursionMarker(
			recursionJob.Send.Request.HttpSpec.Version, recursionPos, currentPos)
		// HttpSpec.Headers
		for i := 0; i < len(recursionJob.Send.Request.HttpSpec.Headers); i++ {
			recursionJob.Send.Request.HttpSpec.Headers[i], currentPos = insertRecursionMarker(
				recursionJob.Send.Request.HttpSpec.Headers[i], recursionPos, currentPos)
		}
		// Data
		recursionJob.Send.Request.Data, _ = insertRecursionMarker(
			recursionJob.Send.Request.Data, recursionPos, currentPos)
	}
	// reactDns调用
	if strings.Index(fuzz1.Send.Request.URL, "dns://") == 0 {
		reaction = reactDns(newReq, resp)
	}
	if fuzz1.React.Reactors != "" { // reactor调用（只调用一个）
		reactors := plugin.ParsePluginsStr(fuzz1.React.Reactors)
		r := reactors[0]
		reqJson, _ := json.Marshal(newReq)
		respJson, _ := json.Marshal(resp)
		reaction = (plugin.Call(plugin.PTypeReactor, r, reqJson, respJson)).(*fuzzTypes.Reaction)
	}
	// 决定是否输出
	// 自定义reactor没有标识响应是否会被过滤，根据Matcher和Filter来确定
	if (reaction.Flag&fuzzTypes.ReactFlagMatch == 0) && (reaction.Flag&fuzzTypes.ReactFlagFiltered == 0) {
		if (!respFiltered && !respMatched) || (respFiltered && !respMatched) {
			reaction.Flag |= fuzzTypes.ReactFlagFiltered
		} else {
			reaction.Flag |= fuzzTypes.ReactFlagMatch
			reaction.Flag |= fuzzTypes.ReactFlagOutput
		}
	} else {
		if reaction.Flag&fuzzTypes.ReactFlagMatch != 0 {
			reaction.Flag |= fuzzTypes.ReactFlagOutput
		}
	}
	// 如果没有添加新任务，且启用递归模式，则添加任务由默认递归逻辑确定
	if reaction.NewJob == nil {
		reaction.NewJob = recursionJob
	}
	// 生成输出消息
	msgBuilder := strings.Builder{}
	if (reaction.Flag&fuzzTypes.ReactFlagOutput != 0) && !reaction.Output.Overwrite {
		rawRespFirstLine := getFirstLine(resp.RawResponse)
		switch fuzz1.React.Verbosity {
		case 1: // 详细程度为1，只输出使用的关键字和payload, rawResponse的第一行以及http重定向链（如果有）
			if len(keywordsUsed) == 1 {
				msgBuilder.WriteString(fmt.Sprintf("%-63s\t--->\t%s\n", payloadEachKeyword[0],
					rawRespFirstLine))
			} else {
				for i := 0; i < len(keywordsUsed); i++ {
					if i == len(keywordsUsed)-1 {
						msgBuilder.WriteString(fmt.Sprintf(
							"%-10s : %-50s\t--->\t%s\n",
							keywordsUsed[i],
							payloadEachKeyword[i],
							rawRespFirstLine))
						continue
					}
					msgBuilder.WriteString(fmt.Sprintf("%-10s : %-50s\t\n", keywordsUsed[i], payloadEachKeyword[i]))
				}
			}
		case 2: // 详细程度为2，输出r使用的关键字和payload, req.data, req.URL, 重定向链, resp第一行
			msgBuilder.WriteString(newReq.URL)
			msgBuilder.WriteByte('\n')
			msgBuilder.WriteString(newReq.Data)
			msgBuilder.WriteByte('\n')
			if len(keywordsUsed) == 1 {
				msgBuilder.WriteString(fmt.Sprintf("%-63s\t--->\t%s\n", payloadEachKeyword[0],
					rawRespFirstLine))
			} else {
				for i := 0; i < len(keywordsUsed); i++ {
					if i == len(keywordsUsed)-1 {
						msgBuilder.WriteString(fmt.Sprintf(
							"%-10s : %-50s\t--->\t%s\n",
							keywordsUsed[i],
							payloadEachKeyword[i],
							rawRespFirstLine))
						continue
					}
					msgBuilder.WriteString(fmt.Sprintf("%-10s : %-50s\t\n",
						keywordsUsed[i], payloadEachKeyword[i]))
				}
			}
		case 3: // 详细程度为3，输出整个req和rawResponse
			reqJson, _ := json.Marshal(newReq)
			msgBuilder.Write(reqJson)
			msgBuilder.Write([]byte("\n |\n\\/\n"))
			msgBuilder.Write(resp.RawResponse)
			msgBuilder.Write([]byte{'\n'})
		}
		if resp.HttpRedirectChain != "" {
			msgBuilder.WriteString(resp.HttpRedirectChain + "\n")
		}
		reaction.Output.Msg += msgBuilder.String()
	}
	return reaction
}

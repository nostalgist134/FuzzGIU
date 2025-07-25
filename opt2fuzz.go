package main

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/options"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultKeyword = "MILAOGIU"

func appendPayloadTmp(tempMap map[string]fuzzTypes.PayloadTemp, pluginStrs []string, appendType int, genType string) {
	/*
		-w C:/aaa.txt,Q:/az/www.txt::FUZZ1 -> "FUZZ1":{"C:/aaa.txt,Q:/az/www.txt|wordlist", processor, pllist}
		-pl-gen giu1(1,2,3),zzwa(1,"6666412",3)::FUZZ2 -> "FUZZ2":{"giu1(1,2,3),zzwa(1,\"6666412\",3)|plugin", processor, pllist}
		-pl-processor proc1(1,"hello"),proc2("1234214")::FUZZ2 -> "FUZZ2":{giu1(1,2,3),zzwa(1,"6666412",3)|plugin, "proc1(1,\"hello\"),proc2(\"1234214\")", pllist}
	*/
	const (
		appendGen  = 0
		keywordSep = "::"
		gentypeSep = "|"
	)
	for _, tmp := range pluginStrs {
		indSep := strings.LastIndex(tmp, keywordSep)
		var keyword string
		if indSep+len(keywordSep) >= len(tmp) || indSep == -1 { // 未指定keyword，使用默认keyword
			indSep = len(tmp)
			keyword = defaultKeyword
		} else {
			keyword = tmp[indSep+len(keywordSep):]
		}
		pluginExpression := tmp[:indSep]
		originalGenerators := ""
		originalProcessors := ""
		_, keyExist := tempMap[keyword]
		if appendType == appendGen {
			suffix := genType
			// 判断键是否已经存在
			if keyExist {
				originalGenerators = tempMap[keyword].Generators
				originalGentype := originalGenerators[strings.LastIndex(originalGenerators, gentypeSep)+1:]
				if originalGentype != genType { // 如果原先的生成器类型与现有的不符则不修改，直接退出
					return
				}
				originalProcessors = tempMap[keyword].Processors
				suffix = ","
			} else {
				suffix = gentypeSep + suffix
			}
			// 添加新项
			tempMap[keyword] = fuzzTypes.PayloadTemp{
				// 如果键已经存在，则拼接generator+","+原来的generator
				Generators: pluginExpression + suffix + originalGenerators,
				Processors: originalProcessors,
			}
		} else {
			if keyExist {
				originalGenerators = tempMap[keyword].Generators
				originalProcessors = tempMap[keyword].Processors
				tempMap[keyword] = fuzzTypes.PayloadTemp{
					Generators: originalGenerators,
					Processors: pluginExpression + originalProcessors,
				}
			} else {
				return
			}
		}
	}
}

// parseCommaSeparatedStr 将形如1,2,3-9,11的字符串转化为int型切片
func parseCommaSeparatedStr(str string) []int {
	if str == "" {
		return nil
	}
	strSplit := strings.Split(str, ",")
	ret := make([]int, 0)
	for _, s := range strSplit {
		if ranges := strings.Split(s, "-"); len(ranges) == 2 {
			lower, err := strconv.Atoi(ranges[0])
			if err != nil {
				continue
			}
			upper, err := strconv.Atoi(ranges[1])
			if err != nil {
				continue
			}
			for i := lower; i <= upper; i++ {
				ret = append(ret, i)
			}
		} else {
			j, err := strconv.Atoi(s)
			if err != nil {
				continue
			} else {
				ret = append(ret, j)
			}
		}
	}
	return ret
}

func opt2fuzz(opt *options.Opt) *fuzzTypes.Fuzz {
	fuzz := new(fuzzTypes.Fuzz)
	// opt.General
	fuzz.Send.Request.URL = opt.General.URL
	fuzz.Send.Request.Data = opt.General.Data
	fuzz.Send.Timeout = opt.General.Timeout
	fuzz.Misc.PoolSize = opt.General.RoutinePoolSize
	fuzz.Misc.Delay = opt.General.Delay
	// opt.HTTP
	var req *fuzzTypes.Req
	var textReq []byte
	var readErr error
	if opt.General.ReqFile != "" {
		textReq, readErr = os.ReadFile(opt.General.ReqFile)
	}
	if opt.General.ReqFile != "" && !os.IsNotExist(readErr) && readErr == nil { // 有指定reqFile
		req, _ = parseHttpRequest(textReq)
		if req == nil { // 如果req == nil说明req文件不是http请求表单，此时把文件当作data对待
			fuzz.Send.Request.Data = string(textReq)
		} else {
			if opt.General.URL != "" { // 使用-u指定的url优先级比-r高
				req.URL = opt.General.URL
			}
			fuzz.Send.Request = *req
		}
	}
	if req == nil {
		if opt.General.ReqFile != "" {
			if os.IsNotExist(readErr) {
				fmt.Printf("request fileOutput %s not found, ignored\n", opt.General.ReqFile)
			} else {
				fmt.Printf("error when reading request file %s - %v, skipping\n", opt.General.ReqFile, readErr)
			}
		}
		fuzz.Send.Request.HttpSpec.ForceHttps = opt.HTTP.HTTPS

		if opt.HTTP.HTTP2 == true {
			fuzz.Send.Request.HttpSpec.Version = "2"
		} else {
			fuzz.Send.Request.HttpSpec.Version = "1.1"
		}

		fuzz.Send.Request.HttpSpec.Headers = make([]string, 0)
		for _, h := range opt.HTTP.Headers {
			fuzz.Send.Request.HttpSpec.Headers = append(fuzz.Send.Request.HttpSpec.Headers, h)
		}

		if len(opt.HTTP.Cookies) > 0 {
			cookies := strings.Builder{}
			cookies.WriteString("Cookies: ")
			for i, cookie := range opt.HTTP.Cookies {
				cookies.WriteString(cookie)
				if i != len(opt.HTTP.Cookies)-1 {
					cookies.WriteString("; ")
				}
			}
			fuzz.Send.Request.HttpSpec.Headers = append(fuzz.Send.Request.HttpSpec.Headers, cookies.String())
		}

		fuzz.Send.Proxies = opt.HTTP.Proxies

		fuzz.Send.HttpFollowRedirects = opt.HTTP.FollowRedirect

		fuzz.Send.Request.HttpSpec.Method = opt.HTTP.Method
	}
	// opt.Filter
	fuzz.React.Filter.Lines = parseCommaSeparatedStr(opt.Filter.FilterLines)
	fuzz.React.Filter.Size = parseCommaSeparatedStr(opt.Filter.FilterSize)
	fuzz.React.Filter.Code = parseCommaSeparatedStr(opt.Filter.FilterCode)
	fuzz.React.Filter.Words = parseCommaSeparatedStr(opt.Filter.FilterWords)
	fuzz.React.Filter.Regex = opt.Filter.FilterRegex
	timeBounds := strings.Split(opt.Filter.FilterTime, "-")
	var upbound, downbound int
	if len(timeBounds) > 1 {
		upbound, _ = strconv.Atoi(timeBounds[1])
		downbound, _ = strconv.Atoi(timeBounds[0])
	} else {
		downbound = 0
		upbound, _ = strconv.Atoi(timeBounds[0])
	}
	fuzz.React.Filter.Time.UpBound = time.Duration(upbound) * time.Millisecond
	fuzz.React.Filter.Time.DownBound = time.Duration(downbound) * time.Millisecond
	fuzz.React.Filter.Mode = opt.Filter.FilterMode
	// opt.Match
	fuzz.React.Matcher.Lines = parseCommaSeparatedStr(opt.Matcher.MatcherLines)
	fuzz.React.Matcher.Size = parseCommaSeparatedStr(opt.Matcher.MatcherSize)
	fuzz.React.Matcher.Code = parseCommaSeparatedStr(opt.Matcher.MatcherCode)
	fuzz.React.Matcher.Words = parseCommaSeparatedStr(opt.Matcher.MatcherWords)
	fuzz.React.Matcher.Regex = opt.Matcher.MatcherRegex
	timeBounds = strings.Split(opt.Matcher.MatcherTime, "-")
	if len(timeBounds) > 1 {
		upbound, _ = strconv.Atoi(timeBounds[1])
		downbound, _ = strconv.Atoi(timeBounds[0])
	} else {
		downbound = 0
		upbound, _ = strconv.Atoi(timeBounds[0])
	}
	fuzz.React.Matcher.Time.UpBound = time.Duration(upbound) * time.Millisecond
	fuzz.React.Matcher.Time.DownBound = time.Duration(downbound) * time.Millisecond
	fuzz.React.Matcher.Mode = opt.Matcher.MatcherMode
	// opt.Output
	fuzz.React.OutSettings.Verbosity = opt.Output.Verbosity
	fuzz.React.OutSettings.OutputFormat = opt.Output.Fmt
	fuzz.React.OutSettings.IgnoreError = opt.Output.IgnoreError
	fuzz.React.OutSettings.OutputFile = opt.Output.File
	fuzz.React.OutSettings.NativeStdout = opt.Output.NativeStdout
	// opt.ErrorHandling
	fuzz.Send.Retry = opt.ErrorHandling.Retry
	fuzz.Send.RetryCode = opt.ErrorHandling.RetryOnStatus
	fuzz.Send.RetryRegex = opt.ErrorHandling.RetryRegex
	// opt.PayloadSetting
	fuzz.Preprocess.Mode = opt.Payload.Mode
	fuzz.Preprocess.PlTemp = make(map[string]fuzzTypes.PayloadTemp)
	appendPayloadTmp(fuzz.Preprocess.PlTemp, opt.Payload.Generators, 0, "plugin")
	appendPayloadTmp(fuzz.Preprocess.PlTemp, opt.Payload.Wordlists, 0, "wordlist")
	appendPayloadTmp(fuzz.Preprocess.PlTemp, opt.Payload.Processors, 1, "")
	if opt.Payload.Mode == "sniper" && len(fuzz.Preprocess.PlTemp) > 1 {
		panic("sniper mode only supports single fuzz keyword")
	}
	// opt.Plugin
	sb := strings.Builder{}
	for i, preprocessors := range opt.Plugin.Preprocessors {
		sb.WriteString(preprocessors)
		if i != len(opt.Plugin.Preprocessors)-1 {
			sb.WriteString(",")
		}
	}
	fuzz.Preprocess.Preprocessors = sb.String()
	fuzz.React.Reactor = opt.Plugin.Reactors
	// opt.RecursionControl
	if opt.RecursionControl.Recursion {
		if len(fuzz.Preprocess.PlTemp) > 1 {
			panic("recursion mode only supports single fuzz keyword")
		}
		fuzz.React.RecursionControl.MaxRecursionDepth = opt.RecursionControl.RecursionDepth
		fuzz.React.RecursionControl.StatCodes = make([]int, 0)
		for _, stat := range strings.Split(opt.RecursionControl.RecursionStatus, ",") {
			statCode, err := strconv.Atoi(stat)
			if err == nil {
				fuzz.React.RecursionControl.StatCodes = append(fuzz.React.RecursionControl.StatCodes, statCode)
			}
		}
		fuzz.React.RecursionControl.Regex = opt.RecursionControl.RecursionRegex
		fuzz.React.RecursionControl.Splitter = opt.RecursionControl.RecursionSplitter
		for k, _ := range fuzz.Preprocess.PlTemp {
			// 递归关键字设置为从关键字列表中取的第一个键
			fuzz.React.RecursionControl.Keyword = k
			break
		}
	}
	return fuzz
}

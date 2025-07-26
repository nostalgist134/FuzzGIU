package main

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/options"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultKeyword = "MILAOGIU"

func appendPayloadTmp(tempMap map[string]fuzzTypes.PayloadTemp, pluginStrings []string, appendType int,
	genType string) {
	/*
		-w C:/aaa.txt,Q:/az/www.txt::FUZZ1 -> "FUZZ1":{"C:/aaa.txt,Q:/az/www.txt|wordlist", processor, pllist}
		-pl-gen giu1(1,2,3),zzwa(1,"6666412",3)::FUZZ2 -> "FUZZ2":{"giu1(1,2,3),zzwa(1,\"6666412\",3)|plugin", processor, pllist}
		-pl-processor proc1(1,"hello"),proc2("1234214")::FUZZ2 -> "FUZZ2":{giu1(1,2,3),zzwa(1,"6666412",3)|plugin, "proc1(1,\"hello\"),proc2(\"1234214\")", pllist}
	*/
	const (
		appendGen  = 0
		keywordSep = "::"
	)
	for _, tmp := range pluginStrings {
		// 在命令行参数中，要使用的文件/插件与fuzz关键字使用"::"关联，
		// 比如 -w C:\aaa.txt::FUZZ1, -pl-proc base64,suffix("123")::FUZZ2
		indSep := strings.LastIndex(tmp, keywordSep)
		keyword := ""
		if indSep+len(keywordSep) >= len(tmp) || indSep == -1 { // 未指定keyword，使用默认keyword
			indSep = len(tmp)
			keyword = defaultKeyword
		} else {
			keyword = tmp[indSep+len(keywordSep):]
		}
		pluginExpr := tmp[:indSep]
		p := plugin.ParsePluginsStr(pluginExpr)
		var originalGen = fuzzTypes.PlGen{}
		var originalProc []fuzzTypes.Plugin
		_, keyExist := tempMap[keyword]
		// 添加新的payload生成器
		if appendType == appendGen {
			// 判断键是否已经存在
			if keyExist {
				originalGen = tempMap[keyword].Generators
				originalGenType := tempMap[keyword].Generators.Type
				// 如果原先的生成器类型与现有的不符则不修改，直接退出
				if originalGenType != genType {
					return
				}
				originalProc = tempMap[keyword].Processors
			}
			// 添加新项
			tempMap[keyword] = fuzzTypes.PayloadTemp{
				Generators: fuzzTypes.PlGen{
					Type: originalGen.Type,
					Gen:  append(originalGen.Gen, p...),
				},
				Processors: originalProc,
			}
		} else {
			if keyExist {
				originalGen = tempMap[keyword].Generators
				originalProc = tempMap[keyword].Processors
				tempMap[keyword] = fuzzTypes.PayloadTemp{
					Generators: originalGen,
					Processors: append(originalProc, p...),
				}
			} else {
				return
			}
		}
	}
}

// str2Ranges 将string类型转化为range切片
func str2Ranges(s string) []fuzzTypes.Range {
	if s == "" {
		return nil
	}
	var errRange = fuzzTypes.Range{Upper: -1, Lower: 0}
	ranges := make([]fuzzTypes.Range, 0)
	sRanges := strings.Split(s, ",")
	for _, r := range sRanges {
		// 单个int值
		if strings.Index(r, "-") == -1 {
			v, err := strconv.Atoi(r)
			if err != nil {
				ranges = append(ranges, errRange)
				continue
			}
			ranges = append(ranges, fuzzTypes.Range{Upper: v, Lower: v})
			continue
		}
		// 多个int值
		bounds := strings.Split(r, "-")
		lower, err := strconv.Atoi(bounds[0])
		if err != nil {
			ranges = append(ranges, errRange)
			continue
		}
		upper, err := strconv.Atoi(bounds[1])
		if err != nil {
			ranges = append(ranges, errRange)
		}
		ranges = append(ranges, fuzzTypes.Range{Upper: upper, Lower: lower})
	}
	return ranges
}

func str2TimeBounds(s string) struct {
	Lower time.Duration `json:"lower,omitempty"`
	Upper time.Duration `json:"upper,omitempty"`
} {
	timeBounds := strings.Split(s, "-")
	var upper, lower int
	if len(timeBounds) > 1 {
		upper, _ = strconv.Atoi(timeBounds[1])
		lower, _ = strconv.Atoi(timeBounds[0])
	} else {
		lower = 0
		upper, _ = strconv.Atoi(timeBounds[0])
	}
	return struct {
		Lower time.Duration `json:"lower,omitempty"`
		Upper time.Duration `json:"upper,omitempty"`
	}{
		Upper: time.Duration(upper) * time.Millisecond,
		Lower: time.Duration(lower) * time.Millisecond,
	}
}

func setMatch(fuzzMatch *fuzzTypes.Match, optMatch *options.Match) {
	fuzzMatch.Lines = str2Ranges(optMatch.Lines)
	fuzzMatch.Size = str2Ranges(optMatch.Size)
	fuzzMatch.Code = str2Ranges(optMatch.Code)
	fuzzMatch.Words = str2Ranges(optMatch.Words)
	fuzzMatch.Regex = optMatch.Regex
	fuzzMatch.Time = str2TimeBounds(optMatch.Time)
	fuzzMatch.Mode = optMatch.Mode
}

// parseRequestFile 解析请求文件
func parseRequestFile(fileName string) (req *fuzzTypes.Req, raw []byte, err error) {
	// 尝试解析为http请求
	raw, err = rawData(fileName)
	if err != nil {
		return
	}
	req, err = parseHttpRequest(fileName)
	if err == nil {
		return
	}
	// 尝试解析为json
	req, err = jsonRequest(fileName)
	if err == nil {
		return
	}
	return
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
	var raw []byte
	var err error
	// 指定从文件中读取请求结构（req结构的json或者http请求）
	if opt.General.ReqFile != "" {
		req, raw, err = parseRequestFile(opt.General.ReqFile)
		if req != nil {
			fuzz.Send.Request = *req
			// -u指定的url优先级更高
			fuzz.Send.Request.URL = opt.General.URL
		} else { // 如果不是json或http，则将其视作data
			fuzz.Send.Request.Data = string(raw)
		}
	}
	if err != nil || opt.General.ReqFile == "" {
		if opt.General.ReqFile != "" {
			if os.IsNotExist(err) {
				fmt.Printf("request file %s not found, ignored\n", opt.General.ReqFile)
			} else {
				fmt.Printf("error when parsing request file %s: %v, skipping\n", opt.General.ReqFile, err)
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
	setMatch(&fuzz.React.Filter, opt.Filter)
	// opt.Match
	setMatch(&fuzz.React.Matcher, opt.Matcher)
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
	fuzz.Preprocess.Preprocessors = plugin.ParsePluginsStr(sb.String())
	fuzz.React.Reactor = opt.Plugin.Reactors
	// opt.RecursionControl
	if opt.RecursionControl.Recursion {
		if len(fuzz.Preprocess.PlTemp) > 1 {
			panic("recursion mode only supports single fuzz keyword")
		}
		fuzz.React.RecursionControl.MaxRecursionDepth = opt.RecursionControl.RecursionDepth
		fuzz.React.RecursionControl.StatCodes = str2Ranges(opt.RecursionControl.RecursionStatus)
		fuzz.React.RecursionControl.Regex = opt.RecursionControl.RecursionRegex
		fuzz.React.RecursionControl.Splitter = opt.RecursionControl.RecursionSplitter
		for k, _ := range fuzz.Preprocess.PlTemp {
			// 递归关键字设置为从关键字列表中取的第一个键（递归模式只支持一个关键字，所以怎么取都无所谓了）
			fuzz.React.RecursionControl.Keyword = k
			break
		}
	}
	return fuzz
}

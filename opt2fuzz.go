package main

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageSend"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/input"
	"github.com/nostalgist134/FuzzGIU/components/options"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"os"
	"strconv"
	"strings"
	"time"
)

var globKeywords = make([]string, 0)

const defaultKeyword = "MILAOGIU"

func keywordOverlap(keyword string) (string, bool) {
	for _, k := range globKeywords {
		if strings.Index(k, keyword) != -1 || strings.Index(keyword, k) != -1 {
			return k, true
		}
	}
	return "", false
}

func hasPathTraverse(plugins []fuzzTypes.Plugin) bool {
	for _, p := range plugins {
		// 统一分隔符
		pName := strings.Replace(p.Name, "\\", "/", -1)
		if strings.Contains(pName, "../") || strings.Contains(pName, "/..") {
			return true
		}
	}
	return false
}

func quitIfPathTraverse(p []fuzzTypes.Plugin) {
	if hasPathTraverse(p) {
		fmt.Fprintln(os.Stderr, "path traverse huh? so clever, but not clever enough")
		os.Exit(1)
	}
}

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
		quitIfPathTraverse(p)
		var oldPlGen = fuzzTypes.PlGen{}
		var oldProc []fuzzTypes.Plugin
		_, keyExist := tempMap[keyword]
		if !keyExist {
			k, isOverlap := keywordOverlap(keyword)
			if isOverlap {
				fmt.Fprintf(os.Stderr, "one keyword you added is one another's substring (%s and %s),\n"+
					"which will lead to template parse error in the future, now exitting...\n", k, keyword)
				os.Exit(1)
			}
			globKeywords = append(globKeywords, keyword)
		}
		// 添加新的payload生成器
		if appendType == appendGen {
			// 判断键是否已经存在
			if keyExist {
				oldPlGen = tempMap[keyword].Generators
				// 如果原先的生成器类型与现有的不符则不修改，直接退出
				if tempMap[keyword].Generators.Type != genType {
					return
				}
				oldProc = tempMap[keyword].Processors
			}
			// 添加新项
			tempMap[keyword] = fuzzTypes.PayloadTemp{
				Generators: fuzzTypes.PlGen{
					Type: genType,
					Gen:  append(oldPlGen.Gen, p...),
				},
				Processors: oldProc,
			}
		} else {
			if keyExist {
				oldPlGen = tempMap[keyword].Generators
				oldProc = tempMap[keyword].Processors
				tempMap[keyword] = fuzzTypes.PayloadTemp{
					Generators: oldPlGen,
					Processors: append(oldProc, p...),
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

func getDelayGranularity(gran string) time.Duration {
	switch gran {
	case "ns":
		return time.Nanosecond
	case "us":
		return time.Microsecond
	case "s":
		return time.Second
	default:
		return time.Millisecond
	}
}

// opt2fuzz 将opt结构转化为fuzz结构
func opt2fuzz(opt *options.Opt) *fuzzTypes.Fuzz {
	fuzz := new(fuzzTypes.Fuzz)
	// opt.General
	fuzz.Preprocess.ReqTemplate.URL = opt.General.URL
	fuzz.Preprocess.ReqTemplate.Data = opt.General.Data
	fuzz.Send.Timeout = opt.General.Timeout
	fuzz.Misc.PoolSize = opt.General.RoutinePoolSize
	fuzz.Misc.Delay = opt.General.Delay
	fuzz.Misc.DelayGranularity = getDelayGranularity(opt.General.DelayGranularity)
	if input.Enabled = opt.General.Input; input.Enabled {
		err := input.InitInput(opt.General.InputAddr)
		if err != nil {
			output.PendLog(fmt.Sprintf("failed to init input: %v", err))
			input.Enabled = false
		}
	}
	// opt.Request
	var req *fuzzTypes.Req
	var raw []byte
	var err error
	stageSend.HTTPRandomAgent = opt.Request.RandomAgent
	// 指定从文件中读取请求结构（req结构的json或者http请求）
	if opt.General.ReqFile != "" {
		req, raw, err = parseRequestFile(opt.General.ReqFile)
		if req != nil {
			fuzz.Preprocess.ReqTemplate = *req
			// -u指定的url优先级更高
			fuzz.Preprocess.ReqTemplate.URL = opt.General.URL
		} else { // 如果不是json或http，则将其视作data
			fuzz.Preprocess.ReqTemplate.Data = string(raw)
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
		fuzz.Preprocess.ReqTemplate.HttpSpec.ForceHttps = opt.Request.HTTPS

		if opt.Request.HTTP2 == true {
			fuzz.Preprocess.ReqTemplate.HttpSpec.Version = "2"
		} else {
			fuzz.Preprocess.ReqTemplate.HttpSpec.Version = "1.1"
		}

		fuzz.Preprocess.ReqTemplate.HttpSpec.Headers = make([]string, 0)
		for _, h := range opt.Request.Headers {
			fuzz.Preprocess.ReqTemplate.HttpSpec.Headers = append(fuzz.Preprocess.ReqTemplate.HttpSpec.Headers, h)
		}

		if len(opt.Request.Cookies) > 0 {
			cookies := strings.Builder{}
			cookies.WriteString("Cookies: ")
			for i, cookie := range opt.Request.Cookies {
				cookies.WriteString(cookie)
				if i != len(opt.Request.Cookies)-1 {
					cookies.WriteString("; ")
				}
			}
			fuzz.Preprocess.ReqTemplate.HttpSpec.Headers = append(fuzz.Preprocess.ReqTemplate.HttpSpec.Headers,
				cookies.String())
		}

		fuzz.Send.Proxies = opt.Request.Proxies

		fuzz.Send.HttpFollowRedirects = opt.Request.FollowRedirect

		fuzz.Preprocess.ReqTemplate.HttpSpec.Method = opt.Request.Method
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
	quitIfPathTraverse(fuzz.Preprocess.Preprocessors)
	reactPlugin := plugin.ParsePluginsStr(opt.Plugin.Reactor)
	if len(reactPlugin) == 0 {
		fuzz.React.Reactor = fuzzTypes.Plugin{}
	} else {
		fuzz.React.Reactor = reactPlugin[0]
	}
	quitIfPathTraverse([]fuzzTypes.Plugin{fuzz.React.Reactor})
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

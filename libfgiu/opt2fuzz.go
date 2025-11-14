package libfgiu

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/opt"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

var globKeywords = make([]string, 0)

const defaultKeyword = "MILAOGIU"

func keywordOverlap(keyword string) (string, bool) {
	for _, k := range globKeywords {
		if strings.Contains(k, keyword) || strings.Contains(keyword, k) {
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
		p, _ := plugin.ParsePluginsStr(pluginExpr)
		quitIfPathTraverse(p)
		var oldPlGen = fuzzTypes.PlGen{}
		var oldProc []fuzzTypes.Plugin
		_, keyExist := tempMap[keyword]
		if !keyExist {
			k, isOverlap := keywordOverlap(keyword)
			if isOverlap {
				fmt.Fprintf(os.Stderr, "one keyword you added is one another's substring (%s and %s),\n"+
					"which will lead to template parse error in the future, exitting now\n", k, keyword)
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
			continue
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

func setMatch(fuzzMatch *fuzzTypes.Match, optMatch *opt.Match) {
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

// Opt2fuzz 将opt结构转化为fuzz结构
func Opt2fuzz(o *opt.Opt) (fuzz1 *fuzzTypes.Fuzz, pendingLogs []string) {
	fuzz1 = new(fuzzTypes.Fuzz)
	var err error

	/*--- o.Request ---*/
	var req *fuzzTypes.Req
	var raw []byte

	// 指定从文件中读取请求结构（req结构的json或者http请求）
	if o.Request.ReqFile != "" {
		req, raw, err = parseRequestFile(o.Request.ReqFile)
		if req != nil {
			fuzz1.Preprocess.ReqTemplate = *req
		} else { // 如果不是json或http，则将其视作data
			fuzz1.Preprocess.ReqTemplate.Data = raw
		}
	}

	if o.Request.ReqFile != "" {
		if os.IsNotExist(err) {
			pendingLogs = append(pendingLogs,
				fmt.Sprintf("request file %s not found, ignored\n", o.Request.ReqFile))
		} else if err != nil {
			pendingLogs = append(pendingLogs,
				fmt.Sprintf("error when parsing request file %s: %v, skipping\n", o.Request.ReqFile, err))
		}
	}

	// 无论文件是否读取成功，都读取命令行参数
	if o.Request.URL != "" { // -u指定的url优先级更高
		fuzz1.Preprocess.ReqTemplate.URL = o.Request.URL
	}

	if o.Request.Data != "" {
		fuzz1.Preprocess.ReqTemplate.Data = []byte(o.Request.Data)
	}

	fuzz1.Preprocess.ReqTemplate.HttpSpec.ForceHttps = o.Request.HTTPS

	if o.Request.HTTP2 {
		fuzz1.Preprocess.ReqTemplate.HttpSpec.Proto = "HTTP/2"
	}

	if len(fuzz1.Preprocess.ReqTemplate.HttpSpec.Headers) == 0 {
		fuzz1.Preprocess.ReqTemplate.HttpSpec.Headers = make([]string, 0)
	}
	for _, h := range o.Request.Headers {
		fuzz1.Preprocess.ReqTemplate.HttpSpec.Headers = append(fuzz1.Preprocess.ReqTemplate.HttpSpec.Headers, h)
	}

	fuzz1.Preprocess.ReqTemplate.HttpSpec.RandomAgent = o.Request.RandomAgent

	if len(o.Request.Cookies) > 0 {
		cookies := strings.Builder{}
		cookies.WriteString("Cookies: ")
		for i, cookie := range o.Request.Cookies {
			cookies.WriteString(cookie)
			if i != len(o.Request.Cookies)-1 {
				cookies.WriteString("; ")
			}
		}
		fuzz1.Preprocess.ReqTemplate.HttpSpec.Headers = append(fuzz1.Preprocess.ReqTemplate.HttpSpec.Headers,
			cookies.String())
	}

	fuzz1.Request.Proxies = o.Request.Proxies

	fuzz1.Request.HttpFollowRedirects = o.Request.FollowRedirect

	// 若-r选项指定了http方法，且-X选项没出现过，才使用-r选项的
	if o.Request.Method != "" {
		fuzz1.Preprocess.ReqTemplate.HttpSpec.Method = o.Request.Method
	} else if fuzz1.Preprocess.ReqTemplate.HttpSpec.Method == "" { // 若-r、-X选项均没出现，使用默认的GET方法
		fuzz1.Preprocess.ReqTemplate.HttpSpec.Method = "GET"
	}

	/*--- o.Filter ---*/
	setMatch(&fuzz1.React.Filter, o.Filter)

	/*--- opts.Match ---*/
	setMatch(&fuzz1.React.Matcher, o.Matcher)

	/*--- opts.Output ---*/
	fuzz1.Control.OutSetting.Verbosity = o.Output.Verbosity
	fuzz1.Control.OutSetting.OutputFormat = o.Output.Fmt
	fuzz1.React.IgnoreError = o.Output.IgnoreError
	fuzz1.Control.OutSetting.OutputFile = o.Output.File
	if o.Output.NativeStdout {
		fuzz1.Control.OutSetting.ToWhere = outputFlag.OutToStdout
	} else if o.Output.TviewOutput {
		fuzz1.Control.OutSetting.ToWhere = outputFlag.OutToTview
	}

	/*--- o.ErrorHandling ---*/
	fuzz1.Request.Retry = o.ErrorHandling.Retry
	fuzz1.Request.RetryCode = o.ErrorHandling.RetryOnStatus
	fuzz1.Request.RetryRegex = o.ErrorHandling.RetryRegex

	/*--- opts.PayloadSetting ---*/
	fuzz1.Preprocess.PlTemp = make(map[string]fuzzTypes.PayloadTemp)
	appendPayloadTmp(fuzz1.Preprocess.PlTemp, o.Payload.Generators, 0, "plugin")
	appendPayloadTmp(fuzz1.Preprocess.PlTemp, o.Payload.Wordlists, 0, "wordlist")
	appendPayloadTmp(fuzz1.Preprocess.PlTemp, o.Payload.Processors, 1, "")

	/*--- o.General ---*/
	fuzz1.Request.Timeout = o.General.Timeout
	fuzz1.Control.PoolSize = o.General.RoutinePoolSize
	fuzz1.Control.Delay, err = time.ParseDuration(o.General.Delay)
	if o.General.Iter != "" {
		iterator, _ := plugin.ParsePluginsStr(o.General.Iter)
		if len(iterator) > 1 {
			log.Fatal("only single iterator is permitted")
		}
		if o.General.Iter == "sniper" && len(fuzz1.Preprocess.PlTemp) > 1 {
			log.Fatal("sniper mode only supports single fuzz keyword")
		}
		fuzz1.Control.IterCtrl.Iterator = iterator[0]
	} else {
		fuzz1.Control.IterCtrl.Iterator = fuzzTypes.Plugin{Name: "clusterbomb"}
	}

	/*--- o.Plugin ---*/
	sb := strings.Builder{}
	for i, preprocessors := range o.Plugin.Preprocessors {
		sb.WriteString(preprocessors)
		if i != len(o.Plugin.Preprocessors)-1 {
			sb.WriteString(",")
		}
	}
	fuzz1.Preprocess.Preprocessors, _ = plugin.ParsePluginsStr(sb.String())
	quitIfPathTraverse(fuzz1.Preprocess.Preprocessors)

	if o.Plugin.Reactor != "" {
		reactPlugin, _ := plugin.ParsePluginsStr(o.Plugin.Reactor)
		if len(reactPlugin) == 0 {
			fuzz1.React.Reactor = fuzzTypes.Plugin{}
		} else if len(reactPlugin) > 1 {
			log.Fatal("only single reactor plugin is permitted")
		}
		fuzz1.React.Reactor = reactPlugin[0]
		quitIfPathTraverse([]fuzzTypes.Plugin{fuzz1.React.Reactor})
	}

	/*--- o.RecursionControl ---*/
	if o.RecursionControl.Recursion {
		if len(fuzz1.Preprocess.PlTemp) > 1 {
			log.Fatal("recursion mode only supports single fuzz keyword")
		}
		fuzz1.React.RecursionControl.MaxRecursionDepth = o.RecursionControl.RecursionDepth
		fuzz1.React.RecursionControl.StatCodes = str2Ranges(o.RecursionControl.RecursionStatus)
		fuzz1.React.RecursionControl.Regex = o.RecursionControl.RecursionRegex
		fuzz1.React.RecursionControl.Splitter = o.RecursionControl.RecursionSplitter
		for k, _ := range fuzz1.Preprocess.PlTemp {
			// 递归关键字设置为从关键字列表中取的第一个键（递归模式只支持一个关键字，所以怎么取都无所谓了）
			fuzz1.React.RecursionControl.Keyword = k
			break
		}
	}
	return
}

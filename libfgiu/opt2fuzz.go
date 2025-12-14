package libfgiu

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/opt"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"os"
	"strconv"
	"strings"
	"time"
)

const keywordSep = "::"

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

// cutGenProcArg 将命令行参数解析为生成器/处理器字符串与fuzz关键字
func cutGenProcArg(arg string) (plGenProc string, keyword string) {
	var found bool
	plGenProc, keyword, found = strings.Cut(arg, keywordSep)
	if !found {
		keyword = fuzzTypes.DefaultFuzzKeyword
	}
	return
}

func assignToPluginSlice(assignType string, slice *[]fuzzTypes.Plugin, assign string) error {
	switch assignType {
	case "wordlist":
		if len(*slice) == 0 {
			*slice = make([]fuzzTypes.Plugin, 1)
		}
		if len((*slice)[0].Name) != 0 {
			(*slice)[0].Name += ","
		}
		(*slice)[0].Name += strings.TrimSuffix(assign, ",")
	case "plugin":
		p, err := plugin.ParsePluginsStr(assign)
		if err != nil {
			return fmt.Errorf("failed to assign gen by plugin expression parsing error: %v", err)
		}
		*slice = append(*slice, p...)
	default:
		return fmt.Errorf("unknown gen type '%s'", assignType)
	}
	return nil
}

func appendPlGen(f *fuzzTypes.Fuzz, args []string, genType string) error {
	if f.Preprocess.PlMeta == nil {
		f.Preprocess.PlMeta = make(map[string]*fuzzTypes.PayloadMeta)
	}

	for _, arg := range args {
		gen, keyword := cutGenProcArg(arg) // 分离关键字部分与表达式部分
		switch genType {
		case "wordlists":
			f.AddKeywordWordlists(keyword, []string{gen})
		case "plugin":
			plugins, err := plugin.ParsePluginsStr(arg)
			if err != nil {
				return fmt.Errorf("failed to parse payload generator plugin expression: %w", err)
			}
			f.AddKeywordPlGenPlugins(keyword, plugins)
		}
	}
	return nil
}

func appendPlProc(f *fuzzTypes.Fuzz, args []string) error {
	for _, arg := range args {
		proc, keyword := cutGenProcArg(arg)
		m, ok := f.Preprocess.PlMeta[keyword]
		if !ok {
			return fmt.Errorf("try to assign payload processor to a inexistent keyword '%s'", keyword)
		}
		err := assignToPluginSlice("plugin", &m.Processors, proc)
		if err != nil {
			return err
		}
	}
	return nil
}

// str2Ranges 将string类型转化为range切片
func str2Ranges(s string) fuzzTypes.Ranges {
	if s == "" {
		return fuzzTypes.Ranges{}
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

func str2TimeBounds(s string) fuzzTypes.TimeBound {
	timeBounds := strings.Split(s, "-")
	var upper, lower int
	if len(timeBounds) > 1 {
		upper, _ = strconv.Atoi(timeBounds[1])
		lower, _ = strconv.Atoi(timeBounds[0])
	} else {
		lower = 0
		upper, _ = strconv.Atoi(timeBounds[0])
	}
	return fuzzTypes.TimeBound{
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
	req, err = toHttpRequest(fileName)
	if err == nil {
		return
	}
	// 尝试解析为json
	req, err = toJsonRequest(fileName)
	if err == nil {
		return
	}
	return
}

// Opt2fuzz 将opt结构转化为fuzz结构
func Opt2fuzz(o *opt.Opt) (fuzz1 *fuzzTypes.Fuzz, err error) {
	fuzz1 = new(fuzzTypes.Fuzz)

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
			err = fmt.Errorf("request file %s not found", o.Request.ReqFile)
		} else if err != nil {
			err = fmt.Errorf("failed to parse request file %s: %v", o.Request.ReqFile, err)
		}
	}

	// 无论文件是否读取成功，都读取命令行参数
	if o.Request.URL != "" { // -u指定的url优先级更高
		fuzz1.Preprocess.ReqTemplate.URL = o.Request.URL
	}

	if o.Request.Data != "" {
		fuzz1.Preprocess.ReqTemplate.Data = []byte(o.Request.Data)
	}

	if o.Request.DataFile != "" {
		df, err2 := os.ReadFile(o.Request.DataFile)
		if err2 == nil {
			fuzz1.Preprocess.ReqTemplate.Data = df
		} else {
			err = errors.Join(err, err2)
		}
	}

	fuzz1.Preprocess.ReqTemplate.HttpSpec.ForceHttps = o.Request.HTTPS

	if o.Request.HTTP2 {
		fuzz1.Preprocess.ReqTemplate.HttpSpec.Proto = "HTTP/2"
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
	if o.Output.File != "" {
		fuzz1.Control.OutSetting.ToWhere |= outputFlag.OutToFile
	}
	fuzz1.Control.OutSetting.HttpURL = o.Output.HttpUrl
	if o.Output.HttpUrl != "" {
		fuzz1.Control.OutSetting.ToWhere |= outputFlag.OutToHttp
	}
	if o.Output.NativeStdout {
		fuzz1.Control.OutSetting.ToWhere |= outputFlag.OutToStdout
	} else if o.Output.TviewOutput {
		fuzz1.Control.OutSetting.ToWhere |= outputFlag.OutToTview
	}

	/*--- o.Retry ---*/
	fuzz1.Request.Retry = o.Retry.Retry
	fuzz1.Request.RetryCodes = str2Ranges(o.Retry.RetryOnStatus)
	fuzz1.Request.RetryRegex = o.Retry.RetryRegex

	/*--- opts.PayloadSetting ---*/
	err = errors.Join(err, appendPlGen(fuzz1, o.Payload.Generators, "plugin"))
	err = errors.Join(appendPlGen(fuzz1, o.Payload.Wordlists, "wordlists"))
	err = errors.Join(appendPlProc(fuzz1, o.Payload.Processors))
	fuzz1.Preprocess.PlDeduplicate = o.Payload.Deduplicate

	/*--- o.General ---*/
	fuzz1.Request.Timeout = o.General.Timeout
	fuzz1.Control.PoolSize = o.General.RoutinePoolSize
	var err1 error
	fuzz1.Control.Delay, err1 = time.ParseDuration(o.General.Delay)
	err = errors.Join(err, err1)
	if o.General.Iter != "" {
		iterator, _ := plugin.ParsePluginsStr(o.General.Iter)
		if len(iterator) > 1 {
			err = errors.Join(err, errors.New("only single iterator is permitted"))
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
			sb.WriteByte(',')
		}
	}
	fuzz1.Preprocess.Preprocessors, _ = plugin.ParsePluginsStr(sb.String())
	quitIfPathTraverse(fuzz1.Preprocess.Preprocessors)

	sb.Reset()
	for i, preprocPrior := range o.Plugin.PreprocPriorGen {
		sb.WriteString(preprocPrior)
		if i != len(o.Plugin.Preprocessors)-1 {
			sb.WriteByte(',')
		}
	}
	fuzz1.Preprocess.PreprocPriorGen, _ = plugin.ParsePluginsStr(sb.String())
	quitIfPathTraverse(fuzz1.Preprocess.Preprocessors)

	if o.Plugin.Reactor != "" {
		reactPlugin, _ := plugin.ParsePluginsStr(o.Plugin.Reactor)
		if len(reactPlugin) == 0 {
			fuzz1.React.Reactor = fuzzTypes.Plugin{}
		} else if len(reactPlugin) > 1 {
			err = errors.Join(err, errors.New("only single reactor plugin is permitted"))
		}
		fuzz1.React.Reactor = reactPlugin[0]
		quitIfPathTraverse([]fuzzTypes.Plugin{fuzz1.React.Reactor})
	}

	/*--- o.RecursionControl ---*/
	if o.RecursionControl.Recursion {
		fuzz1.React.RecursionControl.MaxRecursionDepth = o.RecursionControl.RecursionDepth
		fuzz1.React.RecursionControl.StatCodes = str2Ranges(o.RecursionControl.RecursionStatus)
		fuzz1.React.RecursionControl.Regex = o.RecursionControl.RecursionRegex
		fuzz1.React.RecursionControl.Splitter = o.RecursionControl.RecursionSplitter
		for k := range fuzz1.Preprocess.PlMeta {
			// 递归关键字设置为从关键字列表中取的第一个键（递归模式只支持一个关键字，所以怎么取都无所谓了）
			fuzz1.React.RecursionControl.Keyword = k
			break
		}
	}
	return
}

package options

import "flag"

func (sliceStr *sliceStr) String() string {
	return ""
}

func (sliceStr *sliceStr) Set(value string) error {
	*sliceStr = append(*sliceStr, value)
	return nil
}

// ParseOptCmdline 解析命令行参数
func ParseOptCmdline() *Opt {
	general := &General{}
	output := &Output{}
	matcher := &Matcher{}
	filter := &Filter{}
	http := &Http{}
	payload := &PayloadSetting{}
	recursionControl := &RecursionControl{}
	errHandling := &ErrorHandling{}
	pluginSettings := &Plugin{}

	// 一般性设置
	flag.StringVar(&general.URL, "u", "", "url to giu")
	flag.StringVar(&general.Data, "d", "", "request data")
	flag.StringVar(&general.ReqFile, "r", "", "request file")
	flag.IntVar(&general.RoutinePoolSize, "t", 64, "routine pool size")
	flag.IntVar(&general.Timeout, "timeout", 10, "timeout(second)")
	flag.IntVar(&general.Delay, "delay", 0, "delay between each request(millisecond)")
	// 响应匹配器
	flag.StringVar(&matcher.MatcherCode, "mc", "200,204,301,302,307,401,403,405,500",
		"Match status code from response")
	flag.StringVar(&matcher.MatcherSize, "ms", "", "Match response size")
	flag.StringVar(&matcher.MatcherMode, "mmode", "or", "Matcher set operator")
	flag.StringVar(&matcher.MatcherRegex, "mr", "", "Match regexp")
	flag.StringVar(&matcher.MatcherTime, "mt", "",
		"Match time(millisecond) to the first response byte")
	flag.StringVar(&matcher.MatcherWords, "mw", "", "Match amount of words in response")
	flag.StringVar(&matcher.MatcherLines, "ml", "", "Match amount of lines in response")
	// 响应过滤器
	flag.StringVar(&filter.FilterCode, "fc", "", "Filter status code from response")
	flag.StringVar(&filter.FilterSize, "fs", "", "Filter response size")
	flag.StringVar(&filter.FilterMode, "fmode", "and", "Filter set operator")
	flag.StringVar(&filter.FilterRegex, "fr", "", "Filter regexp")
	flag.StringVar(&filter.FilterTime, "ft", "",
		"Filter time(millisecond) to the first response byte")
	flag.StringVar(&filter.FilterWords, "fw", "", "Filter amount of words in response")
	flag.StringVar(&filter.FilterLines, "fl", "", "Filter amount of lines in response")
	// http设置
	flag.StringVar(&http.Method, "W", "GET", "Method")
	flag.Var(&http.Cookies, "b", "Cookies")
	flag.Var(&http.Headers, "H", "http headers to be used")
	flag.BoolVar(&http.HTTP2, "http2", false, "force http2")
	flag.BoolVar(&http.FollowRedirect, "F", true, "follow redirects")
	flag.BoolVar(&http.HTTPS, "s", false, "force https")
	flag.Var(&http.Proxies, "x", "proxies")
	// payload设置
	flag.StringVar(&payload.Mode, "mode", "clusterbomb", "mode for keywords used, basically "+
		"the same as those in burpsuite")
	flag.Var(&payload.Wordlists, "w", "wordlists to be used for payload")
	flag.Var(&payload.Generators, "pl-gen", "plugin payload generators")
	flag.Var(&payload.Processors, "pl-processor", "payload processors")
	// 输出设置
	flag.StringVar(&output.File, "o", "", "file to output")
	flag.StringVar(&output.Fmt, "fmt", "native", "output file format(native, xml or json. only "+
		"for file output)")
	flag.IntVar(&output.Verbosity, "v", 1, "verbosity level(native output format only)")
	flag.BoolVar(&output.IgnoreError, "ie", false, "ignore errors(will not output error message)")
	flag.BoolVar(&output.NativeStdout, "ns", false, "native stdout")
	// 递归控制
	flag.BoolVar(&recursionControl.Recursion, "R", false, "enable recursion mode(only "+
		"support single fuzz keyword)")
	flag.IntVar(&recursionControl.RecursionDepth, "rec-depth", 2, "recursion depth(when recursion "+
		"is enabled)")
	flag.StringVar(&recursionControl.RecursionStatus, "rec-code",
		"200", "Recursion status code(http protocol only)")
	flag.StringVar(&recursionControl.RecursionRegex, "rec-regex", "", "Recursion when matched regex")
	flag.StringVar(&recursionControl.RecursionSplitter, "rec-splitter", "/",
		"splitter to be used to split recursion positions")
	// 错误处理
	flag.IntVar(&errHandling.Retry, "retry", 0, "max retries")
	flag.StringVar(&errHandling.RetryOnStatus, "retry-code", "",
		"retry on status code(http protocol only)")
	flag.StringVar(&errHandling.RetryRegex, "retry-regex", "", "retry when regex matched")
	// 插件
	flag.Var(&pluginSettings.Preprocessors, "preproc", "preprocessor plugin to be used")
	flag.StringVar(&pluginSettings.Reactors, "react", "", "reactor plugin to be used")
	flag.Parse()
	flagIsSet := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		flagIsSet[f.Name] = true
	})
	// 当用户未指定mc参数，但是指定其它过滤器或者匹配器参数时，将mc参数设置为空，避免因匹配器优先级导致过滤器无法正常运行
	// 或者匹配了本无意匹配的条件
	if (flagIsSet["ms"] || flagIsSet["mr"] || flagIsSet["mt"] ||
		flagIsSet["mw"] || flagIsSet["ml"]) && !flagIsSet["mc"] {
		matcher.MatcherCode = ""
	}
	return &Opt{
		General:          general,
		Output:           output,
		Matcher:          matcher,
		Filter:           filter,
		RecursionControl: recursionControl,
		ErrorHandling:    errHandling,
		Plugin:           pluginSettings,
		HTTP:             http,
		Payload:          payload,
	}
}

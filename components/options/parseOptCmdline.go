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
	payload := &PayloadControl{}
	recursionControl := &RecursionControl{}
	errHandling := &ErrorHandling{}
	pluginSettings := &Plugin{}

	// 一般性设置
	flag.StringVar(&general.URL, "u", "", "url to giu")
	flag.StringVar(&general.Data, "d", "", "request data")
	flag.StringVar(&general.ReqFile, "r", "", "request file")
	flag.IntVar(&general.RoutinePoolSize, "t", 64, "routine pool size")
	flag.BoolVar(&general.IgnoreError, "ie", false, "ignore errors")
	flag.IntVar(&general.Timeout, "timeout", 10, "timeout in seconds")
	flag.IntVar(&general.Delay, "delay", 0, "delay between each request(milliseconds)")
	// 响应匹配器
	flag.StringVar(&matcher.MatcherCode, "mc", "200,204,301,302,307,401,403,405,500",
		"Match status code from response")
	flag.StringVar(&matcher.MatcherSize, "ms", "", "Match HTTP response size")
	flag.StringVar(&matcher.MatcherMode, "mmode", "or", "Matcher set operator")
	flag.StringVar(&matcher.MatcherRegex, "mr", "", "Match regexp")
	flag.StringVar(&matcher.MatcherTime, "mt", "",
		"Match how many milliseconds to the first response byte")
	flag.StringVar(&matcher.MatcherWords, "mw", "", "Match amount of words in response")
	flag.StringVar(&matcher.MatcherLines, "ml", "", "Match amount of lines in response")
	// 响应过滤器
	flag.StringVar(&filter.FilterCode, "fc", "", "Filter status code from response")
	flag.StringVar(&filter.FilterSize, "fs", "", "Filter HTTP response size")
	flag.StringVar(&filter.FilterMode, "fmode", "and", "Filter set operator")
	flag.StringVar(&filter.FilterRegex, "fr", "", "Filter regexp")
	flag.StringVar(&filter.FilterTime, "ft", "",
		"Filter seconds to the first response byte")
	flag.StringVar(&filter.FilterWords, "fw", "", "Filter amount of words in response")
	flag.StringVar(&filter.FilterLines, "fl", "", "Filter amount of lines in response")
	// http设置
	flag.StringVar(&http.Method, "X", "GET", "Method")
	flag.Var(&http.Cookies, "b", "Cookies")
	flag.Var(&http.Headers, "H", "http headers to be used")
	flag.BoolVar(&http.HTTP2, "http2", false, "force http2")
	flag.BoolVar(&http.FollowRedirect, "F", true, "follow redirects")
	flag.BoolVar(&http.HTTPS, "s", false, "force https")
	flag.Var(&http.Proxies, "x", "proxies")
	// payload控制
	flag.StringVar(&payload.Mode, "mode", "clusterbomb", "mode when multiple keywords are used")
	flag.Var(&payload.Wordlists, "w", "wordlists")
	flag.Var(&payload.Generators, "pl-gen", "plugin payload generators")
	flag.Var(&payload.Processors, "pl-processor", "plugin payload processors")
	// 输出设置
	flag.IntVar(&output.Verbosity, "v", 1, "Verbosity level")
	flag.StringVar(&output.ToFile, "o", "", "File to output")
	// 递归控制
	flag.BoolVar(&recursionControl.Recursion, "R", false, "enable recursion mode")
	flag.IntVar(&recursionControl.RecursionDepth, "rec-depth", 2, "Recursion depth")
	flag.StringVar(&recursionControl.RecursionStatus, "rec-status-code",
		"200", "Recursion status code(http protocol only)")
	flag.StringVar(&recursionControl.RecursionRegex, "rec-regex", "", "Recursion when matched regex")
	flag.StringVar(&recursionControl.RecursionSplitter, "rec-splitter", "/",
		"splitter to be used to split recursion positions")
	// 错误处理
	flag.IntVar(&errHandling.Retry, "retry", 0,
		"max retry time, when some conditions satisfied or send request error")
	flag.StringVar(&errHandling.RetryOnStatus, "retry-status-code", "",
		"retry on status code(http protocol only)")
	flag.StringVar(&errHandling.RetryRegex, "retry-regex", "", "retry when matched regex")
	// 插件
	flag.Var(&pluginSettings.Preprocessors, "preproc", "preprocessor plugin to be used")
	flag.StringVar(&pluginSettings.Reactors, "reactor", "", "reactor plugin to be used")
	flag.Parse()
	flagIsSet := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		flagIsSet[f.Name] = true
	})
	// patch: 当用户未指定mc参数，但是指定其它过滤器或者匹配器参数时，将mc参数设置为空，避免因匹配器优先级导致过滤器无法正常运行
	// 或者匹配了本无意匹配的条件
	if flagIsSet["fc"] || flagIsSet["fs"] || flagIsSet["fr"] ||
		flagIsSet["ft"] || flagIsSet["fw"] || flagIsSet["fl"] ||
		((flagIsSet["ms"] || flagIsSet["mr"] || flagIsSet["mt"] ||
			flagIsSet["mw"] || flagIsSet["ml"]) && !flagIsSet["mc"]) {
		if !flagIsSet["mc"] {
			matcher.MatcherCode = ""
		}
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

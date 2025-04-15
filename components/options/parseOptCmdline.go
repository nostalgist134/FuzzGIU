package options

import "flag"

/*
	todo: 添加自定义插件、字典的解析功能具体如下
		-1. -w 指定字典，如果出现多个w参数则叠加，单个w参数指定多个字典用“,”隔开，字典支持绑定关键字 C:\1.txt,C:\2.txt:FUZZ1
		-2. -pl-gen - 使用的payload generator，多个generator由“[]”括起并由“,”隔开，generator可以有参数
			格式为: [gen1(1,2,3),gen2,gen3,...]:KEYWORD1,gen1:KEYWORD2,...
		-3. -pl-processor - payload processor，多个由“,”隔开，可以有参数，格式为: [proc1(1,2,3),proc2,...]:KEYWORD1,proc2:KEYWORD2,...
		-4. -reactor - 自定义的reactor，可以有参数
		-5. -preprocessor - 自定义的preprocessor，可以有参数
*/

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
	flag.IntVar(&general.Timeout, "timeout", 10, "timeout in seconds")
	// 响应匹配器
	flag.StringVar(&matcher.MatcherCode, "mc", "200,204,301,302,307,401,403,405,500", "Match status code from response")
	flag.StringVar(&matcher.MatcherSize, "ms", "", "Match HTTP response size")
	flag.StringVar(&matcher.MatcherMode, "mmode", "and", "Matcher set operator")
	flag.StringVar(&matcher.MatcherRegex, "mr", "", "Match regexp")
	flag.StringVar(&matcher.MatcherTime, "mt", "", "Match how many milliseconds to the first response byte")
	flag.StringVar(&matcher.MatcherWords, "mw", "", "Match amount of words in response")
	flag.StringVar(&matcher.MatcherLines, "ml", "", "Match amount of lines in response")
	// 响应过滤器
	flag.StringVar(&filter.FilterCode, "fc", "", "Filter status code from response")
	flag.StringVar(&filter.FilterSize, "fs", "", "Filter HTTP response size")
	flag.StringVar(&filter.FilterMode, "fmode", "and", "Filter set operator")
	flag.StringVar(&filter.FilterRegex, "fr", "", "Filter regexp")
	flag.StringVar(&filter.FilterTime, "ft", "", "Filter how many milliseconds to the first response byte")
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
	flag.StringVar(&recursionControl.RecursionStatus, "rec-status-code", "200", "Recursion status code(http protocol only)")
	flag.StringVar(&recursionControl.RecursionRegex, "rec-regex", "", "Recursion when matched regex")
	flag.StringVar(&recursionControl.RecursionSplitter, "rec-splitter", "/", "splitter to be used to split recursion positions")
	// 错误处理
	flag.IntVar(&errHandling.Retry, "retry", 0, "max retry time, when some conditions satisfied or send request error")
	flag.StringVar(&errHandling.RetryOnStatus, "retry-status-code", "", "retry on status code(http protocol only)")
	flag.StringVar(&errHandling.RetryRegex, "retry-regex", "", "retry when matched regex")
	// 插件
	flag.Var(&pluginSettings.Preprocessors, "preproc", "preprocessor plugin to be used")
	flag.StringVar(&pluginSettings.Reactors, "reactor", "", "reactor plugin to be used")
	flag.Parse()
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

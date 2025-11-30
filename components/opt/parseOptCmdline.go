package opt

import "flag"

func (s *sliceStr) String() string {
	return ""
}

func (s *sliceStr) Set(value string) error {
	*s = append(*s, value)
	return nil
}

// ParseOptCmdline 解析命令行参数
func ParseOptCmdline() *Opt {
	general := &General{}
	output := &Output{}
	matcher := &Match{}
	filter := &Match{}
	request := &Request{}
	payload := &PayloadSetting{}
	recursionControl := &RecursionControl{}
	errHandling := &Retry{}
	pluginSettings := &Plugin{}
	apiConfig := &ApiConfig{}

	flag.Usage = usage

	// 一般性设置
	flag.IntVar(&general.RoutinePoolSize, "t", 64, "routine pool size")
	flag.IntVar(&general.Timeout, "timeout", 10, "timeout(second)")
	flag.StringVar(&general.Delay, "delay", "0s", "delay between each job submission")
	flag.StringVar(&general.Iter, "iter", "clusterbomb", "embedded/plugin iterator to be "+
		"used(embedded: clusterbomb, pitchfork, pitchfork-cycle)")

	// 响应匹配器
	flag.StringVar(&matcher.Code, "mc", "200,204,301,302,307,401,403,405,500",
		"match status code from response")
	flag.StringVar(&matcher.Size, "ms", "", "match response size")
	flag.StringVar(&matcher.Mode, "mmode", "or", "matcher set operator")
	flag.StringVar(&matcher.Regex, "mr", "", "match regexp")
	flag.StringVar(&matcher.Time, "mt", "",
		"match time(millisecond) to the first response byte")
	flag.StringVar(&matcher.Words, "mw", "", "match amount of words in response")
	flag.StringVar(&matcher.Lines, "ml", "", "match amount of lines in response")

	// 响应过滤器
	flag.StringVar(&filter.Code, "fc", "", "filter status code from response")
	flag.StringVar(&filter.Size, "fs", "", "filter response size")
	flag.StringVar(&filter.Mode, "fmode", "or", "filter set operator")
	flag.StringVar(&filter.Regex, "fr", "", "filter regexp")
	flag.StringVar(&filter.Time, "ft", "", "filter millisecond to the first response byte")
	flag.StringVar(&filter.Words, "fw", "", "filter amount of words in response")
	flag.StringVar(&filter.Lines, "fl", "", "filter amount of lines in response")

	// 请求设置
	flag.StringVar(&request.URL, "u", "", "url to fuzz")
	flag.StringVar(&request.Data, "d", "", "request data")
	flag.StringVar(&request.DataFile, "df", "", "request data from file")
	flag.StringVar(&request.ReqFile, "r", "", "read request(json or http, if parsable) or "+
		"data(unparsable) from file")
	flag.StringVar(&request.Method, "X", "GET", "http method")
	flag.BoolVar(&request.HTTP2, "http2", false, "force http2")
	flag.BoolVar(&request.FollowRedirect, "F", false, "follow redirects")
	flag.BoolVar(&request.HTTPS, "s", false, "force https")
	flag.BoolVar(&request.RandomAgent, "ra", false, "http random agent")
	flag.Var(&request.Cookies, "b", "http cookies")
	flag.Var(&request.Headers, "H", "http headers to be used")
	flag.Var(&request.Proxies, "x", "proxies")

	// payload设置
	flag.Var(&payload.Wordlists, "w", "wordlists to be used for payload")
	flag.Var(&payload.Generators, "pl-gen", "plugin payload generators")
	flag.Var(&payload.Processors, "pl-proc", "payload processors")

	// 输出设置
	flag.StringVar(&output.File, "out-file", "", "file to output")
	flag.StringVar(&output.HttpUrl, "out-url", "", "post output to a http url")
	flag.StringVar(&output.Fmt, "fmt", "native", "output format(native, xml or json)")
	flag.IntVar(&output.Verbosity, "v", 1, "verbosity level(native output format only)")
	flag.BoolVar(&output.IgnoreError, "ie", false, "ignore errors(will not output error message)")
	flag.BoolVar(&output.NativeStdout, "ns", false, "native stdout")
	flag.BoolVar(&output.TviewOutput, "tview", true, "tview output")

	// 递归控制
	flag.BoolVar(&recursionControl.Recursion, "R", false, "enable recursion mode(only "+
		"support single fuzz keyword)")
	flag.IntVar(&recursionControl.RecursionDepth, "rec-depth", 2, "recursion depth(when recursion "+
		"is enabled)")
	flag.StringVar(&recursionControl.RecursionStatus, "rec-code", "",
		"recursion status code(http protocol only)")
	flag.StringVar(&recursionControl.RecursionRegex, "rec-regex", "", "recursion when matched regex")
	flag.StringVar(&recursionControl.RecursionSplitter, "rec-splitter", "/",
		"splitter to be used to split recursion positions")

	// 错误处理
	flag.IntVar(&errHandling.Retry, "retry", 0, "max retries")
	flag.StringVar(&errHandling.RetryOnStatus, "retry-code", "",
		"retry on status code(http protocol only)")
	flag.StringVar(&errHandling.RetryRegex, "retry-regex", "", "retry when regex matched")

	// 插件
	flag.Var(&pluginSettings.Preprocessors, "preproc", "preprocessor plugins to be used")
	flag.Var(&pluginSettings.PreprocPriorGen, "preproc-prior-gen", "preprocessor plugins to be used"+
		"before generating payloads")
	flag.StringVar(&pluginSettings.Reactor, "react", "", "reactor plugin to be used")

	// http api设置
	flag.BoolVar(&apiConfig.HttpApi, "http-api", false, "enable http api mode")
	flag.BoolVar(&apiConfig.ApiTLS, "api-tls", false, "run http api server on https")
	flag.StringVar(&apiConfig.ApiAddr, "api-addr", "0.0.0.0:14514", "http api server listen address")
	flag.StringVar(&apiConfig.TLSKeyFile, "tls-cert-key", "", "tls cert key file to be used")
	flag.StringVar(&apiConfig.TLSCertFile, "tls-cert-file", "", "tls cert file to be used")

	flag.Parse()
	flagIsSet := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		flagIsSet[f.Name] = true
	})

	// 当用户未指定mc参数，但是指定其它匹配器参数时，将mc参数设置为空，避免匹配本无意匹配的条件
	if (flagIsSet["ms"] || flagIsSet["mr"] || flagIsSet["mt"] || flagIsSet["mw"] || flagIsSet["ml"]) &&
		!flagIsSet["mc"] {
		matcher.Code = ""
	}

	// 将http方法选项设为空，避免覆盖-r指定文件中的http方法
	if !flagIsSet["X"] && flagIsSet["r"] {
		request.Method = ""
	}

	return &Opt{
		General:          general,
		Output:           output,
		Matcher:          matcher,
		Filter:           filter,
		RecursionControl: recursionControl,
		Retry:            errHandling,
		Plugin:           pluginSettings,
		Request:          request,
		Payload:          payload,
		ApiConfig:        apiConfig,
	}
}

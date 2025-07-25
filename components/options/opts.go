package options

type (
	sliceStr []string
	Filter   struct {
		FilterCode  string `json:"filter_code,omitempty"`
		FilterSize  string `json:"filter_size,omitempty"`
		FilterTime  string `json:"filter_time,omitempty"`
		FilterMode  string `json:"filter_mode,omitempty"`
		FilterRegex string `json:"filter_regex,omitempty"`
		FilterLines string `json:"filter_lines,omitempty"`
		FilterWords string `json:"filter_words,omitempty"`
	}
	Matcher struct {
		MatcherCode  string `json:"matcher_code,omitempty"`
		MatcherSize  string `json:"matcher_size,omitempty"`
		MatcherTime  string `json:"matcher_time,omitempty"`
		MatcherMode  string `json:"matcher_mode,omitempty"`
		MatcherRegex string `json:"matcher_regex,omitempty"`
		MatcherLines string `json:"matcher_lines,omitempty"`
		MatcherWords string `json:"matcher_words,omitempty"`
	}
	Http struct {
		Headers        sliceStr `json:"header,omitempty"`
		Method         string   `json:"method,omitempty"`
		Cookies        sliceStr `json:"cookie,omitempty"`
		Proxies        sliceStr `json:"proxy,omitempty"`
		FollowRedirect bool     `json:"follow_redirect,omitempty"`
		HTTP2          bool     `json:"http2,omitempty"`
		HTTPS          bool     `json:"https,omitempty"`
	}
	PayloadSetting struct {
		Wordlists  sliceStr `json:"wordlists,omitempty"`
		Generators sliceStr `json:"generator,omitempty"`
		Processors sliceStr `json:"processor,omitempty"`
		Mode       string   `json:"mode,omitempty"`
	}
	Output struct {
		Verbosity    int    `json:"verbosity,omitempty"`
		File         string `json:"file,omitempty"`
		Fmt          string `json:"fmt,omitempty"`
		IgnoreError  bool   `json:"ignore_error,omitempty"`
		NativeStdout bool   `json:"native_stdout,omitempty"`
	}
	General struct {
		URL             string `json:"url,omitempty"`
		Data            string `json:"data,omitempty"`
		ReqFile         string `json:"req_file,omitempty"`
		RoutinePoolSize int    `json:"routine_pool_size,omitempty"`
		Timeout         int    `json:"timeout,omitempty"`
		Delay           int    `json:"delay,omitempty"`
	}
	RecursionControl struct {
		Recursion         bool   `json:"recursion,omitempty"`
		RecursionDepth    int    `json:"recursion_depth,omitempty"`
		RecursionStatus   string `json:"recursion_status,omitempty"`
		RecursionRegex    string `json:"recursion_regex,omitempty"`
		RecursionSplitter string `json:"recursion_splitter,omitempty"`
	}
	ErrorHandling struct {
		Timeout       int    `json:"timeout,omitempty"`
		Retry         int    `json:"retry,omitempty"`
		RetryRegex    string `json:"retry_regex,omitempty"`
		RetryOnStatus string `json:"retry_on_status,omitempty"`
	}
	Plugin struct {
		Preprocessors sliceStr `json:"pre_processor,omitempty"`
		Reactors      string   `json:"reactor,omitempty"`
	}
	Opt struct {
		Payload          *PayloadSetting
		Plugin           *Plugin
		Filter           *Filter
		Matcher          *Matcher
		HTTP             *Http
		Output           *Output
		General          *General
		ErrorHandling    *ErrorHandling
		RecursionControl *RecursionControl
	}
)

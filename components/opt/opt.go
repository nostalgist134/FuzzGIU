package opt

type (
	sliceStr []string
	Match    struct {
		Code  string `json:"code,omitempty"`
		Size  string `json:"size,omitempty"`
		Time  string `json:"time,omitempty"`
		Mode  string `json:"mode,omitempty"`
		Regex string `json:"regex,omitempty"`
		Lines string `json:"lines,omitempty"`
		Words string `json:"words,omitempty"`
	}
	Request struct {
		URL            string   `json:"url,omitempty"`
		Data           string   `json:"data,omitempty"`
		DataFile       string   `json:"data_file,omitempty"`
		ReqFile        string   `json:"req_file,omitempty"`
		Headers        sliceStr `json:"header,omitempty"`
		Method         string   `json:"method,omitempty"`
		Cookies        sliceStr `json:"cookie,omitempty"`
		Proxies        sliceStr `json:"proxy,omitempty"`
		FollowRedirect bool     `json:"follow_redirect,omitempty"`
		HTTP2          bool     `json:"http2,omitempty"`
		HTTPS          bool     `json:"https,omitempty"`
		RandomAgent    bool     `json:"random_agent"`
	}
	PayloadSetting struct {
		Wordlists  sliceStr `json:"wordlists,omitempty"`
		Generators sliceStr `json:"generator,omitempty"`
		Processors sliceStr `json:"processor,omitempty"`
	}
	Output struct {
		Verbosity    int    `json:"verbosity,omitempty"`
		File         string `json:"file,omitempty"`
		HttpUrl      string `json:"http_url,omitempty"`
		Fmt          string `json:"fmt,omitempty"`
		IgnoreError  bool   `json:"ignore_error,omitempty"`
		NativeStdout bool   `json:"native_stdout,omitempty"`
		TviewOutput  bool   `json:"tview_output,omitempty"`
	}
	General struct {
		RoutinePoolSize int    `json:"routine_pool_size,omitempty"`
		JobConcurrency  int    `json:"job_concurrency,omitempty"`
		Timeout         int    `json:"timeout,omitempty"`
		Delay           string `json:"delay,omitempty"`
		Iter            string `json:"iter,omitempty"`
	}
	RecursionControl struct {
		Recursion         bool   `json:"recursion,omitempty"`
		RecursionDepth    int    `json:"recursion_depth,omitempty"`
		RecursionStatus   string `json:"recursion_status,omitempty"`
		RecursionRegex    string `json:"recursion_regex,omitempty"`
		RecursionSplitter string `json:"recursion_splitter,omitempty"`
	}
	Retry struct {
		Timeout       int    `json:"timeout,omitempty"`
		Retry         int    `json:"retry,omitempty"`
		RetryRegex    string `json:"retry_regex,omitempty"`
		RetryOnStatus string `json:"retry_on_status,omitempty"`
	}
	Plugin struct {
		Preprocessors   sliceStr `json:"preprocessor,omitempty"`
		PreprocPriorGen sliceStr `json:"preproc_prior_gen,omitempty"`
		Reactor         string   `json:"reactor,omitempty"`
	}
	ApiConfig struct {
		HttpApi     bool   `json:"http_api,omitempty"`
		ApiTLS      bool   `json:"api_tls,omitempty"`
		ApiAddr     string `json:"api_addr,omitempty"`
		TLSCertFile string `json:"tls_cert_file,omitempty"`
		TLSKeyFile  string `json:"tls_key_file,omitempty"`
	}
	Opt struct {
		Payload          *PayloadSetting
		Plugin           *Plugin
		Filter           *Match
		Matcher          *Match
		Request          *Request
		Output           *Output
		General          *General
		Retry            *Retry
		RecursionControl *RecursionControl
		ApiConfig        *ApiConfig
	}
)

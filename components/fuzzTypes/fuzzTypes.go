package fuzzTypes

import (
	"net/http"
	"time"
)

type (
	// Plugin 标记插件名和参数
	Plugin struct {
		Name string
		Args []any
	}
	// Range 表示一个范围中的全部整数，上下界都是闭合的
	Range struct {
		Upper int `json:"upper,omitempty"`
		Lower int `json:"lower,omitempty"`
	}
	HTTPSpec struct {
		Method     string   `json:"method,omitempty" xml:"method,omitempty"`
		Headers    []string `json:"headers,omitempty" xml:"header>headers,omitempty"`
		Version    string   `json:"version,omitempty" xml:"version,omitempty"`
		ForceHttps bool     `json:"force_https,omitempty" xml:"force_https,omitempty"`
	}
	// Req 请求对象
	Req struct {
		URL      string   `json:"url,omitempty" xml:"url,omitempty"`
		HttpSpec HTTPSpec `json:"http_spec,omitempty" xml:"http_spec,omitempty"`
		Data     string   `json:"data,omitempty" xml:"data,omitempty"`
	}
	// Resp 响应对象
	Resp struct {
		HttpResponse      *http.Response `json:"-" xml:"-"`
		ResponseTime      time.Duration  `json:"response_time,omitempty" xml:"response_time,omitempty"`
		Size              int            `json:"size,omitempty" xml:"size,omitempty"`
		Words             int            `json:"words,omitempty" xml:"words,omitempty"`
		Lines             int            `json:"lines,omitempty" xml:"lines,omitempty"`
		HttpRedirectChain string         `json:"http_redirect_chain,omitempty" xml:"http_redirect_chain,omitempty"`
		RawResponse       []byte         `json:"raw_response,omitempty" xml:"raw_response,omitempty"`
		ErrMsg            string         `json:"err_msg,omitempty" xml:"err_msg,omitempty"`
	}
	// Reaction 响应
	Reaction struct {
		Flag   uint32 `json:"flag,omitempty"` // 响应标志
		Output struct {
			Msg       string `json:"msg,omitempty"`       // 输出信息
			Overwrite bool   `json:"overwrite,omitempty"` // 输出信息是否覆盖默认输出信息
		} `json:"output,omitempty"`
		NewJob *Fuzz `json:"new_job,omitempty"` // 如果要添加新任务，新任务结构体指针
		NewReq *Req  `json:"new_req,omitempty"` // 如果要添加新请求，新请求结构体指针
	}
	PlGen struct {
		Type string   `json:"type"`
		Gen  []Plugin `json:"gen"`
	}
	// PayloadTemp 与单个关键字相关联的payload相关设置
	PayloadTemp struct {
		Generators PlGen    `json:"generators,omitempty"`
		Processors []Plugin `json:"processors,omitempty"`
		PlList     []string `json:"pl_list,omitempty"`
	}
	// SendMeta 包括了请求本身以及与请求相关的设置（超时、代理等）的结构
	SendMeta struct {
		Request             *Req   `json:"request,omitempty"`               // 发送的请求
		Proxy               string `json:"proxy,omitempty"`                 // 使用的代理
		HttpFollowRedirects bool   `json:"http_follow_redirects,omitempty"` // 是否重定向
		Retry               int    `json:"retry,omitempty"`                 // 错误重试次数
		RetryCode           string `json:"retry_code,omitempty"`            // 返回特定状态码时重试
		RetryRegex          string `json:"retry_regex,omitempty"`           // 返回匹配正则时重试
		Timeout             int    `json:"timeout,omitempty"`
	}
	// OutputSettings 输出相关设置
	OutputSettings struct {
		Verbosity    int    `json:"verbosity,omitempty"`     // 输出详细程度
		IgnoreError  bool   `json:"ignore_error,omitempty"`  // 是否忽略错误
		OutputFormat string `json:"output_format,omitempty"` // 文件输出格式
		OutputFile   string `json:"output_file,omitempty"`   // 输出文件名
		NativeStdout bool   `json:"native_stdout,omitempty"` // 输出到原生标准输出流
	}
	Match struct {
		Code  []Range `json:"code,omitempty"`
		Lines []Range `json:"lines,omitempty"`
		Words []Range `json:"words,omitempty"`
		Size  []Range `json:"size,omitempty"`
		Regex string  `json:"regex,omitempty"`
		Mode  string  `json:"mode,omitempty"`
		Time  struct {
			Lower time.Duration `json:"lower,omitempty"`
			Upper time.Duration `json:"upper,omitempty"`
		} `json:"time,omitempty"`
	}
	// Fuzz 测试任务结构，包含执行单个fuzz任务所需的所有信息
	Fuzz struct {
		// 预处理阶段的设置
		Preprocess struct {
			// PlTemp map[string]PayloadTemp，键为fuzz关键字，值为使用的generator和processor，generator有两种
			//
			//	1.wordlist
			//	2.plugin
			//
			// 值的格式为 [generatorFiles]|generatorType，例如 C:\dic.txt|wordlist，不同的generatorFiles用“,”隔开
			// 无论是plugin还是wordlist类型，如果指定了多个generatorFiles，那么生成的payloads会叠加
			// plugin类型的generator指定file时能加自定义的参数，直接在文件名后加上(参数列表)，比如 test(1,2,3,4),test2,...|plugin
			// processor为由逗号隔开的多个processor名的列表，也可以有参数，如果指定了多个processor那么会按照在列表中的顺序调用
			PlTemp        map[string]PayloadTemp `json:"pl_temp,omitempty"`
			Preprocessors []Plugin               `json:"preprocessors,omitempty"` // 使用的自定义预处理器
			Mode          string                 `json:"mode,omitempty"`          // 出现多个payload关键字时处理的模式
			ReqTemplate   Req                    `json:"request_tmpl,omitempty"`  // 含有fuzz关键字的请求模板
		} `json:"preprocess,omitempty"`
		// 发包阶段的设置
		Send struct {
			Proxies             []string `json:"proxies,omitempty"`               // 使用的代理
			HttpFollowRedirects bool     `json:"http_follow_redirects,omitempty"` // 是否重定向
			Retry               int      `json:"retry,omitempty"`                 // 错误重试次数
			RetryCode           string   `json:"retry_code,omitempty"`            // 返回特定状态码时重试
			RetryRegex          string   `json:"retry_regex,omitempty"`           // 返回匹配正则时重试
			Timeout             int      `json:"timeout,omitempty"`               // 超时时间
		} `json:"send,omitempty"`
		// 响应阶段的设置
		React struct {
			Reactor          Plugin         `json:"reactors,omitempty"`
			OutSettings      OutputSettings `json:"output_settings,omitempty"` // 输出设置
			Filter           Match          `json:"filter,omitempty"`          // 过滤
			Matcher          Match          `json:"matcher,omitempty"`         // 匹配
			RecursionControl struct {
				RecursionDepth    int     `json:"recursion_depth,omitempty"`     // 当前递归深度
				MaxRecursionDepth int     `json:"max_recursion_depth,omitempty"` // 最大递归深度
				Keyword           string  `json:"keyword,omitempty"`
				StatCodes         []Range `json:"stat_codes,omitempty"`
				Regex             string  `json:"regex,omitempty"`
				Splitter          string  `json:"splitter,omitempty"`
			} `json:"recursion_control,omitempty"`
		} `json:"react,omitempty"`
		// 杂项设置
		Misc struct {
			PoolSize         int           `json:"pool_size,omitempty"`         // 使用的协程池大小
			Delay            int           `json:"delay,omitempty"`             // 主循环中每次等待的时间
			DelayGranularity time.Duration `json:"delay_granularity,omitempty"` // 等待时间的粒度
		} `json:"misc,omitempty"`
	}
)

// Reaction使用的flag
const (
	ReactOutput   = 0x1
	ReactAddJob   = 0x2
	ReactStopJob  = 0x4
	ReactExit     = 0x8
	ReactFiltered = 0x10
	ReactMatch    = 0x20
	ReactAddReq   = 0x40
)

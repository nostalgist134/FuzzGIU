package fuzzTypes

import (
	"net/http"
	"time"
)

type (
	// Req 发送的请求
	Req struct {
		URL      string `json:"url"`
		HttpSpec struct {
			Method     string   `json:"method"`
			Headers    []string `json:"headers"`
			Version    string   `json:"version"`
			ForceHttps bool     `json:"force_https"`
		} `json:"httpSpec,omitempty"`
		Data string `json:"data"`
	}
	// Resp response
	Resp struct {
		HttpResponse      *http.Response `json:"-"`
		ResponseTime      time.Duration  `json:"response_time"`
		Size              int            `json:"size"`
		Words             int            `json:"words"`
		Lines             int            `json:"lines"`
		HttpRedirectChain string         `json:"http_redirect_chain"`
		RawResponse       []byte         `json:"raw_response"`
		RespError         error          `json:"-"` // error位标记在发包过程是否有出错
	}
	// Reaction 响应
	Reaction struct {
		Flag   uint32 `json:"flag"` // 响应行为的标志位
		Output struct {
			Msg       string `json:"msg"`       // 输出信息
			Overwrite bool   `json:"overwrite"` // 输出信息是否覆盖默认输出信息
		} `json:"output"`
		NewJob *Fuzz `json:"new_job,omitempty"` // 如果要添加新任务，新任务的结构体
	}
	PayloadTemp struct {
		Generators string   `json:"generators"`
		Processors string   `json:"processors"`
		PlList     []string `json:"pl_list"`
	}
	SendMeta struct {
		Request             *Req   `json:"request"`               // 发送的请求
		Proxy               string `json:"proxy"`                 // 使用的代理
		HttpFollowRedirects bool   `json:"http_follow_redirects"` // 是否重定向
		Retry               int    `json:"retry"`                 // 错误重试次数
		RetryCode           string `json:"retry_code"`            // 返回特定状态码时重试
		RetryRegex          string `json:"retry_regex"`           // 返回匹配正则时重试
		Timeout             int    `json:"timeout"`
	}
	// Fuzz 测试结构
	Fuzz struct {
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
			PlTemp        map[string]PayloadTemp `json:"pl_temp"`
			Preprocessors string                 `json:"preprocessors"` // 使用的自定义预处理器
			Mode          string                 `json:"mode"`          // 出现多个payload关键字时处理的模式
		} `json:"preprocess"` // 预处理阶段的设置
		Send struct {
			Request             Req      `json:"request"`               // 发送的请求
			Proxies             []string `json:"proxies"`               // 使用的代理
			HttpFollowRedirects bool     `json:"http_follow_redirects"` // 是否重定向
			Retry               int      `json:"retry"`                 // 错误重试次数
			RetryCode           string   `json:"retry_code"`            // 返回特定状态码时重试
			RetryRegex          string   `json:"retry_regex"`           // 返回匹配正则时重试
			Timeout             int      `json:"timeout"`
		} `json:"send"` // 发包阶段的设置
		React struct {
			Reactor     string `json:"reactors"`  // 使用的自定义响应器
			Verbosity   int    `json:"verbosity"` // 输出详细程度
			IgnoreError bool   `json:"ignore_error"`
			Filter      struct {
				Code  []int  `json:"code"`
				Lines []int  `json:"lines"`
				Words []int  `json:"words"`
				Size  []int  `json:"size"`
				Regex string `json:"regex"`
				Mode  string `json:"mode"`
				Time  struct {
					DownBound time.Duration `json:"down_bound"`
					UpBound   time.Duration `json:"up_bound"`
				} `json:"time"`
			} `json:"filter"` // 过滤
			Matcher struct {
				Code  []int  `json:"code"`
				Lines []int  `json:"lines"`
				Words []int  `json:"words"`
				Size  []int  `json:"size"`
				Regex string `json:"regex"`
				Mode  string `json:"mode"`
				Time  struct {
					DownBound time.Duration `json:"down_bound"`
					UpBound   time.Duration `json:"up_bound"`
				} `json:"time"`
			} `json:"matcher"` // 匹配
			RecursionControl struct {
				RecursionDepth    int    `json:"recursion_depth"`     // 当前递归深度
				MaxRecursionDepth int    `json:"max_recursion_depth"` // 最大递归深度
				Keyword           string `json:"keyword"`
				StatCodes         []int  `json:"stat_codes"`
				Regex             string `json:"regex"`
				Splitter          string `json:"splitter"`
			} `json:"recursion_control"`
		} `json:"react"` // 响应阶段的设置
		Misc struct {
			PoolSize int `json:"pool_size"` // 使用的协程池大小
			Delay    int `json:"delay"`
		} `json:"misc"` // 杂项设置
	}
)

// Reaction使用的flag
const (
	ReactFlagOutput   = 0x1
	ReactFlagAddJob   = 0x2
	ReactFlagStopJob  = 0x4
	ReactFlagExit     = 0x8
	ReactFlagFiltered = 0x10
	ReactFlagMatch    = 0x20
	ReactError        = 0x40
)

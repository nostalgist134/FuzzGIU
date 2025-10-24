package fuzzTypes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type (
	// Plugin 标记插件名和参数
	Plugin struct {
		Name string `json:"name,omitempty"`
		Args []any  `json:"args,omitempty"`
	}

	// Range 表示一个范围中的全部整数，上下界都是闭合的
	Range struct {
		Upper int `json:"upper,omitempty"`
		Lower int `json:"lower,omitempty"`
	}

	// Field 描述一个请求字段
	Field struct {
		Name  string `json:"name"`  // 字段名
		Value string `json:"value"` // 字段值
	}

	HTTPSpec struct {
		Method      string   `json:"method,omitempty" xml:"method,omitempty"`
		Headers     []string `json:"headers,omitempty" xml:"header>headers,omitempty"`
		Version     string   `json:"version,omitempty" xml:"version,omitempty"`
		ForceHttps  bool     `json:"force_https,omitempty" xml:"force_https,omitempty"`
		RandomAgent bool     `json:"http_random_agent,omitempty"`
	}

	// Req 请求对象
	Req struct {
		URL      string   `json:"url,omitempty" xml:"url,omitempty"`             // 请求url
		HttpSpec HTTPSpec `json:"http_spec,omitempty" xml:"http_spec,omitempty"` // http相关的设置与字段
		Fields   []Field  `json:"fields,omitempty" xml:"fields,omitempty"`       // 请求中的额外字段
		Data     []byte   `json:"data,omitempty" xml:"data,omitempty"`           // 数据载体
	}

	// Resp 响应对象
	Resp struct {
		HttpResponse      *http.Response `json:"-" xml:"-"` // http响应包（但是tag标记为空，因为不能反序列化）
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

	// RequestCtx 包括了请求本身以及与请求相关的设置（超时、代理等）的结构
	RequestCtx struct {
		Request             *Req   `json:"request,omitempty"`               // 发送的请求
		Proxy               string `json:"proxy,omitempty"`                 // 使用的代理
		Retry               int    `json:"retry,omitempty"`                 // 错误重试次数
		HttpFollowRedirects bool   `json:"http_follow_redirects,omitempty"` // http重定向
		RetryCode           string `json:"retry_code,omitempty"`            // 返回特定状态码时重试
		RetryRegex          string `json:"retry_regex,omitempty"`           // 返回匹配正则时重试
		Timeout             int    `json:"timeout,omitempty"`               // 超时
	}

	// OutputSetting 输出相关设置
	OutputSetting struct {
		Verbosity    int    `json:"verbosity,omitempty"`     // 输出详细程度
		OutputFile   string `json:"output_file,omitempty"`   // 输出文件名
		OutputFormat string `json:"output_format,omitempty"` // 文件输出格式
		HttpURL      string `json:"http_url,omitempty"`      // 将结果POST到http url上
		ChanSize     int    `json:"chan_size,omitempty"`     // 使用管道输出时，管道的大小
		ToWhere      int32  `json:"to_where,omitempty"`      // 输出到什么地方（文件、屏幕、管道）
	}

	// Iteration 迭代设置
	Iteration struct {
		Start    int    `json:"start"`    // 迭代起始下标
		End      int    `json:"end"`      // 迭代终止下标
		Iterator Plugin `json:"iterator"` // 迭代器
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
			PlTemp        map[string]PayloadTemp `json:"pl_temp,omitempty"`
			Preprocessors []Plugin               `json:"preprocessors,omitempty"` // 使用的自定义预处理器
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
			Reactor          Plugin `json:"reactor,omitempty"`      // 响应处理插件
			Filter           Match  `json:"filter,omitempty"`       // 过滤
			Matcher          Match  `json:"matcher,omitempty"`      // 匹配
			IgnoreError      bool   `json:"ignore_error,omitempty"` // 是否忽略发送过程中出现的错误
			RecursionControl struct {
				RecursionDepth    int     `json:"recursion_depth,omitempty"`     // 当前递归深度
				MaxRecursionDepth int     `json:"max_recursion_depth,omitempty"` // 最大递归深度
				Keyword           string  `json:"keyword,omitempty"`
				StatCodes         []Range `json:"stat_codes,omitempty"`
				Regex             string  `json:"regex,omitempty"`
				Splitter          string  `json:"splitter,omitempty"`
			} `json:"recursion_control,omitempty"`
		} `json:"react,omitempty"`
		// 任务控制设置
		Control struct {
			PoolSize   int           `json:"pool_size,omitempty"`   // 使用的协程池大小
			Delay      time.Duration `json:"delay,omitempty"`       // 每次提交任务前的延迟
			OutSetting OutputSetting `json:"out_setting,omitempty"` // 输出设置
			IterCtrl   Iteration     `json:"iter_ctrl,omitempty"`   // 迭代控制
		} `json:"control,omitempty"`
	}
)

// Reaction使用的flag
const (
	ReactOutput = 1 << iota
	ReactAddJob
	ReactStopJob
	ReactFiltered
	ReactMatch
	ReactAddReq
	ReactMerge
)

/* -----------下面是给plugin类用的序列化/反序列化函数，由于plugin用了any切片，因此需要稍微特殊处理一下----------- */
/*
下面的函数是为了解决默认情况下处理any切片的序列化与反序列化时出现的类型丢失问题，其实最主要的就是int类型丢失问题。
go默认情况下会将json字节中所有的数字转化为float64，这也就代表如果args中含有一个int类型参数，序列化后再反序列化
回来时，这个int类型会变成float64，但是插件调用对于参数类型是敏感的（处理参数时使用类型断言），因此必须想办法将参
数的类型信息在序列化时注入，因此采用下面的自定义的序列化与反序列化函数。
原本我是准备使用json.Number实现数字的自动处理，但是现在还是放弃了，原因在于：如果一个any类型是float，但是它的小
数位为0，这种情况下序列化之后它的点号会被去掉，也就无法分辨它到底是float还是int了
插件系统调用约定中仅仅允许int、float64、string以及bool这4种类型的参数，因此其它的就不考虑了
*/
type marshalInterior struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

// MarshalJSON 自定义Plugin类序列化函数
func (p Plugin) MarshalJSON() ([]byte, error) {
	tmp := marshalInterior{Name: p.Name}

	// 将any类型全部转为“类型+字面值”的字符串表示
	if len(p.Args) != 0 {
		args := make([]string, len(p.Args))
		for i, arg := range p.Args {
			args[i] = fmt.Sprintf("%T %v", arg, arg)
		}
		tmp.Args = args
	}
	return json.Marshal(tmp)
}

func (p *Plugin) UnmarshalJSON(data []byte) error {
	tmp := marshalInterior{}
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}

	p.Name = tmp.Name
	p.Args = make([]any, len(tmp.Args))

	for i, expr := range tmp.Args {
		typ, val, ok := strings.Cut(expr, " ")
		if !ok {
			return fmt.Errorf("incorrect string expression at arg index %d", i)
		}
		switch typ {
		case "int":
			intVal, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return fmt.Errorf("incorrect %s value at arg index %d", typ, i)
			}
			p.Args[i] = int(intVal)
		case "bool":
			boolVal, _ := strconv.ParseBool(val)
			p.Args[i] = boolVal
		case "float64":
			floatVal, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return fmt.Errorf("incorrect %s value at arg index %d", typ, i)
			}
			p.Args[i] = floatVal
		case "string":
			p.Args[i] = val
		default:
			return fmt.Errorf("unsupported type %s at arg index %d", typ, i)
		}
	}
	return nil
}

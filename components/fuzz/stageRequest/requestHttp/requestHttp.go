package requestHttp

import (
	"bytes"
	"crypto/tls"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// http agents 放到别的文件去了

var cliPool = sync.Pool{
	New: func() any {
		return new(http.Client)
	},
}

// initHttpCli 初始化 Http 客户端，设置代理、超时、重定向等
func initHttpCli(proxy string, timeout int, redirect bool, redirectChain *string) (*http.Client, error) {
	// 从池中获取一个 http.Client
	cli := cliPool.Get().(*http.Client)

	tr := http.DefaultTransport.(*http.Transport).Clone()
	if tr.TLSClientConfig == nil {
		tr.TLSClientConfig = &tls.Config{}
	}
	tr.TLSClientConfig.InsecureSkipVerify = true
	tr.ForceAttemptHTTP2 = true
	// 设置代理
	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	} else {
		tr.Proxy = http.ProxyFromEnvironment
	}
	cli.Transport = tr
	if redirect {
		// 是否跟随重定向
		cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			// 简化重定向链拼接逻辑：via是已完成的请求，req是下一个请求
			if len(*redirectChain) == 0 && len(via) > 0 {
				*redirectChain = via[0].URL.String() // 初始请求URL（via[0]是第一个请求）
			}
			*redirectChain += "->" + req.URL.String()
			return nil
		}
	} else {
		cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 不跟随重定向
		}
	}

	// 设置超时
	cli.Timeout = time.Duration(timeout) * time.Second
	return cli, nil
}

// fuzzReq2HttpReq 根据fuzzTypes.req结构生成http.Request结构
func fuzzReq2HttpReq(fuzzReq *fuzzTypes.Req) (*http.Request, error) {
	var err error = nil
	var httpReq *http.Request
	URL := fuzzReq.URL
	httpReq, err = http.NewRequest(fuzzReq.HttpSpec.Method, URL, bytes.NewBuffer(fuzzReq.Data))
	if fuzzReq.HttpSpec.ForceHttps { // 强制使用https
		httpReq.URL.Scheme = "https"
	}
	if err != nil {
		return nil, err
	}
	// proto，即HTTP/1.1部分
	httpReq.Proto = fuzzReq.HttpSpec.Proto
	// http请求头部分
	for i := 0; i < len(fuzzReq.HttpSpec.Headers); i++ {
		indColon := strings.Index(fuzzReq.HttpSpec.Headers[i], ":")
		headerName := ""
		headerVal := ""
		// 如果没有冒号，则加入一个值为空的头（net/http不允许无冒号的单独头名）
		if indColon == len(fuzzReq.HttpSpec.Headers[i])-1 || indColon == -1 {
			headerName = fuzzReq.HttpSpec.Headers[i]
		} else {
			headerName = fuzzReq.HttpSpec.Headers[i][:indColon]
			headerVal = strings.TrimSpace(fuzzReq.HttpSpec.Headers[i][indColon+1:])
		}
		if headerName == "Host" {
			httpReq.Host = headerVal
		} else {
			httpReq.Header.Add(headerName, headerVal)
		}
	}
	// 设置UA头
	if httpReq.Header.Get("User-Agent") == "" {
		if fuzzReq.HttpSpec.RandomAgent {
			httpReq.Header.Set("User-Agent", agents[rand.Int()%len(agents)])
		} else {
			httpReq.Header.Set("User-Agent", "milaogiu browser(114.54)")
		}
	}
	return httpReq, err
}

func containRetryCode(code int, retryCodes []string) bool {
	codeStr := strconv.Itoa(code)
	for _, c := range retryCodes {
		if c == codeStr {
			return true
		}
	}
	return false
}

func countWords(data []byte) int {
	count := 0
	inWord := false
	for _, b := range data {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == '\f' || b == '\v' {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	line := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		line++
	}
	return line
}

// DoRequestHttp http发包函数
func DoRequestHttp(reqCtx *fuzzTypes.RequestCtx, timeout int, httpRedirect bool, retry int,
	retryCode, retryRegex, proxy string) (*fuzzTypes.Resp, error) {

	resp := new(fuzzTypes.Resp)
	request := reqCtx.Request

	// HTTP/1.x 使用FastHTTP
	if request.HttpSpec.Proto != "HTTP/2" {
		return doRequestFastHttp(reqCtx, timeout, httpRedirect, retry, retryCode, retryRegex, proxy)
	}

	httpReq, err := fuzzReq2HttpReq(request)
	if err != nil {
		resp.ErrMsg = err.Error()
		return resp, err
	}

	var redirectChain string
	cli, err := initHttpCli(proxy, timeout, httpRedirect, &redirectChain)
	if err != nil {
		resp.ErrMsg = err.Error()
		return resp, err
	}
	defer cliPool.Put(cli) // 确保客户端放回池中

	timeStart := time.Now()
	retryCodes := strings.Split(retryCode, ",")
	reqData := request.Data // 保存原始请求数据

	var httpResponse *http.Response
	var sendErr error

	// 主请求和重试逻辑
	for attempt := 0; attempt <= retry; attempt++ {
		if attempt > 0 {
			// 重试延迟
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
			// 重新创建body
			httpReq.Body = io.NopCloser(bytes.NewBuffer(reqData))
		}

		httpResponse, sendErr = cli.Do(httpReq)

		if sendErr != nil {
			resp.ErrMsg = sendErr.Error()
			if attempt < retry {
				continue // 继续重试
			}
			break
		}

		// 构建响应
		rawResp, bodyBytes, buildErr := buildRawHTTPResponse(httpResponse)
		if buildErr != nil {
			if httpResponse.Body != nil {
				httpResponse.Body.Close()
			}
			resp.ErrMsg = buildErr.Error()
			return resp, buildErr
		}

		resp.RawResponse = rawResp
		resp.HttpResponse = httpResponse
		resp.HttpRedirectChain = redirectChain

		if rawResp != nil {
			resp.Lines = countLines(bodyBytes)
			resp.Words = countWords(bodyBytes)
			resp.Size = len(bodyBytes)
		}

		// 检查是否需要重试
		needRetry := containRetryCode(httpResponse.StatusCode, retryCodes) ||
			common.RegexMatch(rawResp, retryRegex)

		if !needRetry || attempt >= retry {
			break
		}

		// 准备下一次重试
		if httpResponse.Body != nil {
			httpResponse.Body.Close()
		}
	}

	resp.ResponseTime = time.Since(timeStart)
	return resp, sendErr
}

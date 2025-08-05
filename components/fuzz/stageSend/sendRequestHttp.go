package stageSend

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

func buildRawHTTPResponse(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}
	var raw bytes.Buffer

	// 状态行，例如：HTTP/1.1 302 Found
	raw.WriteString(fmt.Sprintf("HTTP/%d.%d %d %s\r\n",
		resp.ProtoMajor, resp.ProtoMinor, resp.StatusCode, resp.Status))

	// 响应头
	for k, vals := range resp.Header {
		for _, v := range vals {
			raw.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	raw.WriteString("\r\n")

	// 响应体
	if resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		raw.Write(bodyBytes)

		// 重新填充 resp.Body 以便后续代码还能使用它
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return raw.Bytes(), nil
}

var cliPool = sync.Pool{
	New: func() any {
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			ExpectContinueTimeout: 1 * time.Second,
		}
		return &http.Client{
			Transport: transport,
		}
	},
}

// 初始化 Http 客户端，设置代理、超时、重定向等
func initHttpCli(proxy string, timeout int, redirect bool, httpVer string,
	redirectChain *string) (*http.Client, error) {
	// 从池中获取一个 http.Client
	cli := cliPool.Get().(*http.Client)

	// 设置 Transport
	var forceHttp2 bool
	ver, parseErr := strconv.ParseFloat(httpVer, 32)
	if parseErr != nil || ver < 2.0 {
		forceHttp2 = false
	} else {
		forceHttp2 = true
	}
	tr := cli.Transport.(*http.Transport)
	tr.ForceAttemptHTTP2 = forceHttp2

	// 设置代理
	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		tr.Proxy = http.ProxyURL(proxyUrl)
	} else {
		tr.Proxy = nil
	}
	cli.Transport = tr

	// 是否跟随重定向
	cli.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if !redirect {
			return http.ErrUseLastResponse
		}
		for i, r := range via {
			if i == 0 && len(*redirectChain) == 0 {
				*redirectChain += r.URL.String()
			} else {
				*redirectChain += "->" + r.URL.String()
			}
		}
		*redirectChain += "->" + req.URL.String()
		return nil
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
	if fuzzReq.HttpSpec.ForceHttps { // 强制使用https
		URL = strings.Replace(URL, "http://", "https://", 1)
	}
	httpReq, err = http.NewRequest(fuzzReq.HttpSpec.Method, URL, bytes.NewBuffer([]byte(fuzzReq.Data)))
	if err != nil {
		return nil, err
	}
	// proto，即HTTP/1.1部分
	httpReq.Proto = "HTTP/" + fuzzReq.HttpSpec.Version
	// http请求头部分
	for i := 0; i < len(fuzzReq.HttpSpec.Headers); i++ {
		indColon := strings.Index(fuzzReq.HttpSpec.Headers[i], ":")
		// 如果没有冒号，则加入一个值为空的头（net/http不允许无冒号的单独头名）
		if indColon == len(fuzzReq.HttpSpec.Headers[i])-1 || indColon == -1 {
			httpReq.Header.Add(fuzzReq.HttpSpec.Headers[i], "")
		} else {
			httpReq.Header.Add(fuzzReq.HttpSpec.Headers[i][0:indColon],
				strings.TrimSpace(fuzzReq.HttpSpec.Headers[i][indColon+1:]))
		}
	}
	// 设置UA头
	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", "milaogiu browser(114.54)")
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

// http发包函数
func sendRequestHttp(request *fuzzTypes.Req, timeout int, httpRedirect bool, retry int,
	retryCode, retryRegex, proxy string) (*fuzzTypes.Resp, error) {
	resp := new(fuzzTypes.Resp)
	resp.ErrMsg = ""

	req, err := fuzzReq2HttpReq(request)
	if err != nil {
		resp.ErrMsg = err.Error()
		return resp, err
	}
	redirectChain := ""
	cli, err := initHttpCli(proxy, timeout, httpRedirect, request.HttpSpec.Version, &redirectChain)
	if err != nil {
		resp.ErrMsg = err.Error()
		return resp, err
	}
	timeStart := time.Now()
	httpResponse, sendErr := cli.Do(req)
	var buildErr error = nil
	var rawResp []byte
	if sendErr == nil {
		resp.HttpResponse = httpResponse
		// 生成rawResponse
		rawResp, buildErr = buildRawHTTPResponse(httpResponse)
		if buildErr != nil {
			rawResp = nil
			return resp, buildErr
		}
	} else if retry == 0 {
		return resp, sendErr
	}

	retryCodes := strings.Split(retryCode, ",")
	// 重试：正则匹配、发送出错、返回码匹配时
	if (retryRegex != "" && common.RegexMatch(rawResp, retryRegex)) ||
		sendErr != nil || containRetryCode(httpResponse.StatusCode, retryCodes) {
		// 获取request.Data从而给下面的http请求使用
		reqData := []byte(request.Data)
		// 重新请求URL直到达到重试次数或者出错
		for ; retry > 0; retry-- {
			if httpResponse != nil {
				httpResponse.Body.Close()
			}
			// patchLog#2: 重新填充data部分，因为http.request每次发完请求后就会把body消耗掉（完全傻逼的设计）
			req.Body = io.NopCloser(bytes.NewBuffer(reqData))
			// 重试请求
			httpResponse, sendErr = cli.Do(req)
			resp.HttpResponse = httpResponse
			if sendErr != nil {
				resp.ErrMsg = sendErr.Error() // 每次发送请求后都设置respError位
			} else {
				resp.ErrMsg = ""
			}
			rawResp, buildErr = buildRawHTTPResponse(httpResponse)
			if buildErr != nil {
				rawResp = nil
				return resp, buildErr
			}
			// 不满足重试条件（正则不匹配、发送成功、返回码不匹配），不再重试，并将error设置为nil
			if retryRegex != "" && !common.RegexMatch(rawResp, retryRegex) && sendErr == nil &&
				!containRetryCode(httpResponse.StatusCode, retryCodes) {
				resp.ErrMsg = ""
				break
			}
		}
	}
	cliPool.Put(cli) // cli用完之后放回
	resp.ResponseTime = time.Since(timeStart)
	resp.RawResponse = rawResp
	resp.HttpRedirectChain = redirectChain // 记录重定向链
	if rawResp != nil {
		resp.Lines = bytes.Count(rawResp, []byte{'\n'})
		if rawResp[len(rawResp)-1] != '\n' {
			resp.Lines++
		}
		resp.Words = countWords(rawResp)
		resp.Size = len(rawResp)
	}
	// 关闭HttpResponse的body，因为rawResponse已经记录body了，这个成员之后不会再用
	if resp.HttpResponse != nil {
		resp.HttpResponse.Body.Close()
	}
	return resp, err
}

package requestHttp

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	regn "github.com/xsxo/regnhttp"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var regnClientsByPxy = sync.Map{}

func fuzzReq2Regn(req *fuzzTypes.Req) (*regn.RequestType, *url.URL, error) {
	regnReq := regn.Http2Request()
	regnReq.SetMethod(req.HttpSpec.Method)
	regnReq.SetURL(req.URL)
	hasUa := false
	for _, h := range req.HttpSpec.Headers {
		name, value, _ := strings.Cut(h, ":")
		regnReq.Header.Add(name, strings.TrimSpace(value))
		if strings.EqualFold(name, "user-agent") {
			hasUa = true
		}
	}
	if !hasUa && req.HttpSpec.RandomAgent {
		regnReq.Header.Set("User-Agent", getRandAgent())
	}
	regnReq.SetBody(req.Data)
	u, err := url.Parse(req.URL)
	return regnReq, u, err
}

// getRegnCli 按照代理分类的regn.client池
// regn.client一旦使用了代理，必须把连接全关掉才能重设，但是这样就不能复用链接了，因此使用这种方法
func getRegnCli(pxy string, timeout int) *regn.Client {
	actual, _ := regnClientsByPxy.LoadOrStore(pxy, &sync.Pool{New: func() any {
		c := new(regn.Client)
		if pxy != "" {
			c.Proxy(pxy)
		}
		return c
	}})
	cli := actual.(*sync.Pool).Get().(*regn.Client)
	cli.Timeout = time.Duration(timeout) * time.Second
	return cli
}

func putRegnCli(toPut *regn.Client, pxy string) {
	v, ok := regnClientsByPxy.Load(pxy)
	if !ok {
		return
	}
	if toPut.HttpVesrion() != 2 { // 仅当client不为http2时才升级回去，避免任何时候都调用upgrade导致链接关闭
		toPut.Http2Upgrade()
	}
	v.(*sync.Pool).Put(toPut)
}

func needRedir(stat int) bool {
	return stat >= 301 && stat < 309 && stat != 304
}

func regnHttpRequest(cli *regn.Client, req *regn.RequestType, resp *regn.ResponseType,
	redir bool, baseUrl *url.URL) (error, string) {
	const maxRedirects = 10 // 限制最大重定向次数，防止死循环
	redirects := strings.Builder{}
	currentRedirectCount := 0

	// 关键：复制baseUrl（避免修改外部传入的URL对象，防止污染）
	prevUrl, err := url.Parse(baseUrl.String())
	if err != nil {
		return fmt.Errorf("parse base URL failed: %v", err), ""
	}

	firstUrl := true

	for {
		// 执行请求（HTTP2/HTTP1模式由客户端自动管理）
		if err = cli.Do(req, resp); err != nil {
			return fmt.Errorf("request failed: %v", err), redirects.String()
		}

		stat := resp.StatusCode()
		// 不需要重定向：直接退出
		if !redir || !needRedir(stat) {
			break
		}

		if firstUrl {
			firstUrl = false
			redirects.WriteString(prevUrl.String())
		}

		// 超过最大重定向次数
		if currentRedirectCount >= maxRedirects {
			return fmt.Errorf("too many redirects (max allowed: %d)", maxRedirects), redirects.String()
		}

		// 重定向必须包含Location头（HTTP标准要求）
		loc := resp.Header.Get("Location")
		if loc == "" {
			return fmt.Errorf("redirect status %d without Location header", stat), redirects.String()
		}

		// 解析重定向URL：基于上一次重定向的URL（符合HTTP标准）
		locUrl, err := url.Parse(loc)
		if err != nil {
			return fmt.Errorf("invalid redirect location '%s': %v", loc, err), redirects.String()
		}
		nextUrl := prevUrl.ResolveReference(locUrl) // 相对路径基于上一次URL解析

		// HTTP2模式下重定向到HTTP明文，请求降级为HTTP1（regn不允许http2明文，会panic）
		switch nextUrl.Scheme {
		case "http":
			req.HttpDowngrade()
			resp.HttpDowngrade()
		case "https":
			req.Http2Upgrade()
			resp.Http2Upgrade()
		default:
			return fmt.Errorf("unsupported scheme '%s' in redirect location [url: %s]",
				nextUrl.Scheme, nextUrl.String()), redirects.String()
		}
		if nextUrl.Scheme != prevUrl.Scheme {
			if nextUrl.Scheme == "http" {
				cli.HttpDowngrade()
			} else {
				cli.Http2Upgrade()
			}
		}

		// 302/303重定向切换为GET并清空Body
		if stat == 302 || stat == 303 {
			req.SetMethod("GET")
			req.SetBody(nil)
		}

		// 更新请求URL和重定向链
		req.SetURL(nextUrl.String())
		if redirects.Len() > 0 {
			redirects.WriteString(" -> ")
		}
		redirects.WriteString(nextUrl.String())

		// 准备下一次循环：以上一次重定向URL作为新基准
		prevUrl = nextUrl
		currentRedirectCount++
	}

	return nil, redirects.String()
}

// DoRequestHttpNew 尝试改用regn库，避免之前的net/http库hpack panic问题
func DoRequestHttpNew(reqCtx *fuzzTypes.RequestCtx) (*fuzzTypes.Resp, error) {
	request := reqCtx.Request

	// HTTP/1.x 使用fasthttp库
	if request.HttpSpec.Proto != "HTTP/2" {
		return doRequestFastHttp(reqCtx)
	}

	resp := new(fuzzTypes.Resp)

	proxy, timeout, httpRedirect, retryCodes, retry, retryRegex :=
		reqCtx.Proxy, reqCtx.Timeout, reqCtx.HttpFollowRedirects, reqCtx.RetryCodes, reqCtx.Retry, reqCtx.RetryRegex

	regnReq, u, err := fuzzReq2Regn(request)
	if err != nil {
		return resp, err
	}
	regnResp := regn.Http2Response()

	var redirectChain string
	cli := getRegnCli(proxy, timeout)
	defer putRegnCli(cli, proxy)

	timeStart := time.Now()

	// 主请求和重试逻辑
	for attempt := 0; attempt <= retry; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
		}

		err, redirectChain = regnHttpRequest(cli, regnReq, regnResp, httpRedirect, u)

		if err != nil {
			resp.ErrMsg = err.Error()
			if attempt < retry {
				continue // 继续重试
			}
			break
		}

		// 构建响应

		resp.RawResponse = regnResp.Body()
		resp.Statistic()
		resp.RawResponse = regnResp.Raw()
		resp.HttpResponse = &http.Response{StatusCode: regnResp.StatusCode()}
		resp.HttpRedirectChain = redirectChain

		// 检查是否需要重试
		needRetry := retryCodes.Contains(regnResp.StatusCode()) || common.RegexMatch(resp.RawResponse, retryRegex)

		if !needRetry || attempt >= retry {
			break
		}
	}

	resp.ResponseTime = time.Since(timeStart)
	return resp, err
}

package stageSend

import (
	"bytes"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
	"unsafe"
)

// buildRawHTTPResponse1 将 fasthttp.Response 转为原始 HTTP 响应 []byte
func buildRawHTTPResponse1(resp *fasthttp.Response) []byte {
	var buf bytes.Buffer
	// 写状态行
	fmt.Fprintf(&buf, "%s %d %s\r\n",
		resp.Header.Protocol(),
		resp.StatusCode(),
		resp.Header.StatusMessage())
	// 写 header
	resp.Header.VisitAll(func(key, value []byte) {
		buf.Write(key)
		buf.WriteString(": ")
		buf.Write(value)
		buf.WriteString("\r\n")
	})
	// 空行分隔 header 与 body
	buf.WriteString("\r\n")
	// 写 body
	buf.Write(resp.Body())
	return buf.Bytes()
}

var fhCliPool = sync.Pool{New: func() any {
	return &fasthttp.Client{
		ReadTimeout:                   0,
		WriteTimeout:                  0,
		MaxIdleConnDuration:           90 * time.Second,
		NoDefaultUserAgentHeader:      true,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      1,
			DNSCacheDuration: time.Hour,
		}).Dial,
	}
}}

func unsafeStringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func fuzzReq2FHReq(fr *fuzzTypes.Req, fhr *fasthttp.Request) {
	fhr.Header.SetMethod(fr.HttpSpec.Method)
	u := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(u)
	u.Update(fr.URL)
	if fr.HttpSpec.ForceHttps {
		u.SetScheme("https")
	}
	fhr.SetURI(u)
	fhr.Header.SetProtocol(fr.HttpSpec.Version)
	for _, h := range fr.HttpSpec.Headers {
		indCol := strings.Index(h, ":")
		if indCol == len(h)-1 {
			fhr.Header.Add(h[:indCol], "")
		} else if indCol == -1 {
			fhr.Header.Add(h, "")
		} else {
			fhr.Header.Add(h[:indCol], h[indCol+1:])
		}
	}
	if ua := fhr.Header.Peek("User-Agent"); len(ua) == 0 {
		if HTTPRandomAgent {
			fhr.Header.Set("User-Agent", agents[rand.Int()%len(agents)])
		} else {
			fhr.Header.Set("User-Agent", "milaogiu browser(114.54)")
		}
	}
	fhr.SetBody(unsafeStringToBytes(fr.Data))
}

func getFHCli(pxy string) *fasthttp.Client {
	cli := (fhCliPool.Get()).(*fasthttp.Client)
	if pxy != "" {
		cli.Dial = fasthttpproxy.FasthttpHTTPDialer(pxy)
	}
	return cli
}

func fastHttpRequest(cli *fasthttp.Client, fhReq *fasthttp.Request, fhResp *fasthttp.Response, redirect bool,
	timeout int) (error, string) {

	timeOut := time.Second * time.Duration(timeout)
	rdrChain := strings.Builder{}

	tmp := fasthttp.AcquireRequest()
	reqSend := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(tmp)
	defer fasthttp.ReleaseRequest(reqSend)
	fhReq.CopyTo(tmp)
	tmp.CopyTo(reqSend)

	err := cli.DoTimeout(reqSend, fhResp, timeOut)
	if err != nil {
		return err, ""
	}

	if redirect {
		if stat := fhResp.StatusCode(); stat >= 301 && stat < 309 && stat != 304 {
			loc := fhResp.Header.Peek("Location")
			for i := 0; i < 10; i++ {
				rdrChain.WriteString(tmp.URI().String())

				// 更新 URI
				tmp.URI().UpdateBytes(loc)
				if len(tmp.URI().Path()) == 0 {
					tmp.URI().SetPath("/")
				}
				// 302/303 改成 GET
				if stat == 302 || stat == 303 {
					tmp.Header.SetMethod("GET")
					tmp.SetBody(nil)
				}

				fhResp.Reset()
				reqSend.Reset()
				tmp.CopyTo(reqSend)
				err = cli.DoTimeout(reqSend, fhResp, timeOut)
				if err != nil {
					rdrChain.WriteString(" -> ")
					rdrChain.Write(loc)
					return err, rdrChain.String()
				}

				// 检查是否还需要重定向
				if stat = fhResp.StatusCode(); stat < 301 || stat >= 309 || stat == 304 {
					rdrChain.WriteString(" -> ")
					rdrChain.WriteString(tmp.URI().String())
					break
				}

				loc = fhResp.Header.Peek("Location")

				// 未到达最后一次仍然满足重定向，写入分隔符
				if i != 9 {
					rdrChain.WriteString(" -> ")
				}
			}
		}
	}

	return nil, rdrChain.String()
}

func sendRequestFastHttp(req *fuzzTypes.Req, timeout int, httpRedirect bool, retry int,
	retryCode, retryRegex, proxy string) (*fuzzTypes.Resp, error) {
	resp := new(fuzzTypes.Resp)

	fhReq := fasthttp.AcquireRequest()
	fhResp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(fhReq)
	defer fasthttp.ReleaseResponse(fhResp)

	fuzzReq2FHReq(req, fhReq)

	cli := getFHCli(proxy)
	defer func() {
		// 重置为默认拨号器
		cli.Dial = (&fasthttp.TCPDialer{
			Concurrency:      1,
			DNSCacheDuration: time.Hour,
		}).Dial
		fhCliPool.Put(cli)
	}()

	timeStart := time.Now()
	// 发送请求（带重试和重定向）
	err, rdr := fastHttpRequest(cli, fhReq, fhResp, httpRedirect, timeout)
	rtyCodes := strings.Split(retryCode, ",")
	if retry > 0 {
		// 重试逻辑
		if (retryRegex != "" && common.RegexMatch(fhResp.Body(), retryRegex)) ||
			containRetryCode(fhResp.StatusCode(), rtyCodes) || err != nil {
			for ; retry > 0; retry-- {
				fhResp.Reset()
				err, rdr = fastHttpRequest(cli, fhReq, fhResp, httpRedirect, timeout)
				if !(retryRegex != "" && common.RegexMatch(fhResp.Body(), retryRegex)) &&
					!containRetryCode(fhResp.StatusCode(), rtyCodes) && err == nil {
					break
				}
			}
		}
	}
	// 统计时间为第一次+重试的总时间
	resp.ResponseTime = time.Since(timeStart)
	resp.RawResponse = buildRawHTTPResponse1(fhResp)
	// 填充httpResponse对象（仅需填充status code，因为过滤与匹配只用到status code）
	resp.HttpResponse = &http.Response{StatusCode: fhResp.StatusCode()}
	resp.HttpRedirectChain = rdr
	resp.Size = len(resp.RawResponse)
	resp.Words = countWords(resp.RawResponse)
	resp.Lines = countLines(resp.RawResponse)
	if err != nil {
		resp.ErrMsg = err.Error()
	}
	return resp, err
}

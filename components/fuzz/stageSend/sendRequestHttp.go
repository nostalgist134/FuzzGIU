package stageSend

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var agents = []string{
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36 OPR/26.0.1656.60",
	"Opera/8.0 (Windows NT 5.1; U; en)",
	"Mozilla/5.0 (Windows NT 5.1; U; en; rv:1.8.1) Gecko/20061208 Firefox/2.0.0 Opera 9.50",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; en) Opera 9.50",
	"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
	"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
	"Opera/9.80 (Android 2.3.4; Linux; Opera Mobi/build-1107180945; U; en-GB) Presto/2.8.149 Version/11.10",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; rv:34.0) Gecko/20100101 Firefox/34.0",
	"Mozilla/5.0 (X11; U; Linux x86_64; zh-CN; rv:1.9.2.10) Gecko/20100922 Ubuntu/10.10 (maverick) Firefox/3.6.10",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10.6; rv,2.0.1) Gecko/20100101 Firefox/4.0.1",
	"Mozilla/5.0 (Windows NT 6.1; rv,2.0.1) Gecko/20100101 Firefox/4.0.1",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/534.57.2 (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",
	"MAC：Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.122 Safari/537.36",
	"Windows：Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	"Mozilla/5.0 (iPhone; U; CPU iPhone OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/5.0 (iPhone; U; CPU iPhone OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/5.0 (iPad; U; CPU OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.71 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11",
	"Mozilla/5.0 (Windows; U; Windows NT 6.1; en-US) AppleWebKit/534.16 (KHTML, like Gecko) Chrome/10.0.648.133 Safari/534.16",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64; Trident/7.0; rv:11.0) like Gecko",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/536.11 (KHTML, like Gecko) Chrome/20.0.1132.11 TaoBrowser/2.0 Safari/536.11",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.1 (KHTML, like Gecko) Chrome/21.0.1180.71 Safari/537.1 LBBROWSER",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; LBBROWSER)",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; QQDownload 732; .NET4.0C; .NET4.0E; LBBROWSER)",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; WOW64; Trident/5.0; SLCC2; .NET CLR 2.0.50727; .NET CLR 3.5.30729; .NET CLR 3.0.30729; Media Center PC 6.0; .NET4.0C; .NET4.0E; QQBrowser/7.0.3698.400)",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1; SV1; QQDownload 732; .NET4.0C; .NET4.0E)",
	"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.84 Safari/535.11 SE 2.X MetaSr 1.0",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Trident/4.0; SV1; QQDownload 732; .NET4.0C; .NET4.0E; SE 2.X MetaSr 1.0)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Trident/4.0; SE 2.X MetaSr 1.0; SE 2.X MetaSr 1.0; .NET CLR 2.0.50727; SE 2.X MetaSr 1.0)",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Maxthon/4.4.3.4000 Chrome/30.0.1599.101 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_0) AppleWebKit/535.11 (KHTML, like Gecko) Chrome/17.0.963.56 Safari/535.11",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/38.0.2125.122 UBrowser/4.0.3214.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.87 UBrowser/6.2.4094.1 Safari/537.36",
	"Mozilla/5.0 (iPhone; U; CPU iPhone OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/5.0 (iPod; U; CPU iPhone OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/5.0 (iPad; U; CPU OS 4_2_1 like Mac OS X; zh-cn) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8C148 Safari/6533.18.5",
	"Mozilla/5.0 (iPad; U; CPU OS 4_3_3 like Mac OS X; en-us) AppleWebKit/533.17.9 (KHTML, like Gecko) Version/5.0.2 Mobile/8J2 Safari/6533.18.5",
	"Mozilla/5.0 (Linux; U; Android 2.2.1; zh-cn; HTC_Wildfire_A3333 Build/FRG83D) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	"Mozilla/5.0 (Linux; U; Android 2.3.7; en-us; Nexus One Build/FRF91) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	"MQQBrowser/26 Mozilla/5.0 (Linux; U; Android 2.3.7; zh-cn; MB200 Build/GRJ22; CyanogenMod-7) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	"Opera/9.80 (Android 2.3.4; Linux; Opera Mobi/build-1107180945; U; en-GB) Presto/2.8.149 Version/11.10",
	"Mozilla/5.0 (Linux; U; Android 3.0; en-us; Xoom Build/HRI39) AppleWebKit/534.13 (KHTML, like Gecko) Version/4.0 Safari/534.13",
	"Mozilla/5.0 (BlackBerry; U; BlackBerry 9800; en) AppleWebKit/534.1+ (KHTML, like Gecko) Version/6.0.0.337 Mobile Safari/534.1+",
	"Mozilla/5.0 (hp-tablet; Linux; hpwOS/3.0.0; U; en-US) AppleWebKit/534.6 (KHTML, like Gecko) wOSBrowser/233.70 Safari/534.6 TouchPad/1.0",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows NT 6.1; Trident/5.0;",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 6.0)",
	"Mozilla/4.0 (compatible; MSIE 8.0; Windows NT 6.0; Trident/4.0)",
	"Mozilla/4.0 (compatible; MSIE 6.0; Windows NT 5.1)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; TencentTraveler 4.0)",
	"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Avant Browser)",
	"Mozilla/5.0 (Linux; U; Android 2.3.7; en-us; Nexus One Build/FRF91) AppleWebKit/533.1 (KHTML, like Gecko) Version/4.0 Mobile Safari/533.1",
	"Mozilla/5.0 (SymbianOS/9.4; Series60/5.0 NokiaN97-1/20.0.019; Profile/MIDP-2.1 Configuration/CLDC-1.1) AppleWebKit/525 (KHTML, like Gecko) BrowserNG/7.1.18124",
	"Mozilla/5.0 (compatible; MSIE 9.0; Windows Phone OS 7.5; Trident/5.0; IEMobile/9.0; HTC; Titan)",
	"UCWEB7.0.2.37/28/999",
	"NOKIA5700/ UCWEB7.0.2.37/28/999",
	"Openwave/ UCWEB7.0.2.37/28/999",
	"Openwave/ UCWEB7.0.2.37/28/999",
}

var HTTPRandomAgent = false

func buildRawHTTPResponse(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}
	var raw bytes.Buffer

	// 状态行
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
		if HTTPRandomAgent {
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

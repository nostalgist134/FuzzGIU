package stageSend

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"net/url"
)

/*func sendRequestDebug(meta *fuzzTypes.SendMeta) *fuzzTypes.Resp {
	reqJ, _ := json.Marshal(request)

	fmt.Printf("[DEBUG] sending %s with circumstance:\n", string(reqJ))
	fmt.Printf("[DEBUG] proxy: %s\n", proxy)
	fmt.Printf("[DEBUG] timeout: %d\n", timeout)
	fmt.Printf("[DEBUG] retry: %d\n", retry)
	fmt.Printf("[DEBUG] retryCode: %s\n", retryCode)
	fmt.Printf("[DEBUG] retryRegex: %s\n", retryRegex)
	//fmt.Printf("[DEBUG] Sending Request %s\n", meta.Request.URL)
	sCode := rand.Int() % 300
	httpResp := &http.Response{StatusCode: sCode, Header: http.Header{}}
	return &fuzzTypes.Resp{
		HttpResponse:      httpResp,
		HttpRedirectChain: "NISHIGIU->WOSHIGIU->MILAOGIU",
		Size:              3,
		Words:             4,
		Lines:             5,
		RawResponse:       []byte("HTTP/1.1 " + strconv.Itoa(sCode) + " OK\r\n\r\n"),
		ErrMsg:         nil,
	}
}*/

func SendRequest(meta *fuzzTypes.SendMeta) *fuzzTypes.Resp {
	u, err := url.Parse(meta.Request.URL)
	var retResp *fuzzTypes.Resp
	if err != nil { // 无法解析URL
		retResp = &fuzzTypes.Resp{}
		retResp.ErrMsg = err.Error()
		return retResp
	}
	switch u.Scheme {
	case "http", "https", "":
		resp, sendErr := sendRequestHttp(meta.Request, meta.Timeout, meta.HttpFollowRedirects,
			meta.Retry, meta.RetryCode, meta.RetryRegex, meta.Proxy)
		if sendErr != nil && resp != nil && resp.ErrMsg == "" {
			resp.ErrMsg = sendErr.Error()
		}
		retResp = resp
	case "ws", "wss":
		retResp = sendRequestWs(meta.Request, meta.Timeout, meta.Retry, meta.RetryRegex)
	case "dns":
		retResp = sendRequestDns(meta.Request, meta.Timeout)
	default:
		p := fuzzTypes.Plugin{Name: u.Scheme}
		retResp = plugin.SendRequest(p, meta)
	}
	if retResp == nil {
		return &fuzzTypes.Resp{ErrMsg: "nil response"}
	}
	if retResp.RawResponse == nil {
		retResp.RawResponse = []byte("")
	}
	return retResp
}

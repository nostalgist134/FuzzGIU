package stageSend

import (
	"FuzzGIU/components/fuzzTypes"
	"FuzzGIU/components/plugin"
	"encoding/json"
	"errors"
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
		RespError:         nil,
	}
}*/

func SendRequest(meta *fuzzTypes.SendMeta) *fuzzTypes.Resp {
	u, err := url.Parse(meta.Request.URL)
	if err != nil { // 无法解析的URL默认发给http
		resp, sendErr := sendRequestHttp(meta.Request, meta.Timeout, meta.HttpFollowRedirects,
			meta.Retry, meta.RetryCode, meta.RetryRegex, meta.Proxy)
		if sendErr != nil && resp != nil && resp.RespError == nil {
			resp.RespError = sendErr
		}
		return resp
	}
	var retResp *fuzzTypes.Resp
	switch u.Scheme {
	case "http", "https":
		resp, sendErr := sendRequestHttp(meta.Request, meta.Timeout, meta.HttpFollowRedirects,
			meta.Retry, meta.RetryCode, meta.RetryRegex, meta.Proxy)
		if sendErr != nil && resp != nil && resp.RespError == nil {
			resp.RespError = sendErr
		}
		retResp = resp
	case "ws", "wss":
		retResp = sendRequestWs(meta.Request, meta.Timeout, meta.Retry, meta.RetryRegex)
	case "dns":
		retResp = sendRequestDns(meta.Request, meta.Timeout)
	default:
		p := plugin.Plugin{Name: u.Scheme}
		sendMetaJson, _ := json.Marshal(meta)
		retResp = (plugin.Call(plugin.PTypeReqSender, p, sendMetaJson, nil)).(*fuzzTypes.Resp)
	}
	if retResp == nil {
		return &fuzzTypes.Resp{RespError: errors.New("nil response")}
	}
	if retResp.RawResponse == nil {
		retResp.RawResponse = []byte("")
	}
	return retResp
}

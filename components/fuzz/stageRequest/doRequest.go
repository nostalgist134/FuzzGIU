package stageRequest

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageRequest/requestHttp"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageRequest/requestWs"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"net/url"
)

// DoRequest 根据RequestCtx请求上下文发送请求
func DoRequest(rCtx *fuzzTypes.RequestCtx, scheme string) *fuzzTypes.Resp {
	if rCtx == nil || rCtx.Request == nil {
		return &fuzzTypes.Resp{ErrMsg: "nil request to send"}
	}

	var (
		ret     *fuzzTypes.Resp
		uScheme string
	)

	// 若请求使用的scheme已经成功预解析，就不必再调用url.Parse
	if scheme != "" {
		uScheme = scheme
	} else {
		if rCtx.Request.URL == "" { // URL为空
			ret = &fuzzTypes.Resp{ErrMsg: "empty url to request"}
			return ret
		}
		u, err := url.Parse(rCtx.Request.URL)
		if err != nil { // 无法解析URL
			ret = &fuzzTypes.Resp{ErrMsg: err.Error()}
			return ret
		}
		uScheme = u.Scheme
	}

	switch uScheme {
	case "http", "https", "": // 若没有scheme，默认使用http
		resp, sendErr := requestHttp.DoRequestHttp(rCtx)
		if sendErr != nil && resp != nil && resp.ErrMsg == "" {
			resp.ErrMsg = sendErr.Error()
		}
		ret = resp
	case "ws", "wss":
		ret = requestWs.DoRequestWs(rCtx.Request, rCtx.Timeout, rCtx.Retry, rCtx.RetryRegex)
	default:
		p := fuzzTypes.Plugin{Name: uScheme}
		ret = plugin.DoRequest(p, rCtx)
	}

	if ret == nil {
		ret = &fuzzTypes.Resp{ErrMsg: "nil response"}
	}

	if ret.RawResponse == nil {
		ret.RawResponse = []byte("")
	}
	return ret
}

package stageDoReq

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageDoReq/doHttp"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageDoReq/doWs"
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
		u, err := url.Parse(rCtx.Request.URL)
		if err != nil { // 无法解析URL
			ret = &fuzzTypes.Resp{}
			ret.ErrMsg = err.Error()
			return ret
		}
		uScheme = u.Scheme
	}

	switch uScheme {
	case "http", "https", "":
		resp, sendErr := doHttp.DoRequestHttp(rCtx, rCtx.Timeout, rCtx.HttpFollowRedirects,
			rCtx.Retry, rCtx.RetryCode, rCtx.RetryRegex, rCtx.Proxy)
		if sendErr != nil && resp != nil && resp.ErrMsg == "" {
			resp.ErrMsg = sendErr.Error()
		}
		ret = resp
	case "ws", "wss":
		ret = doWs.DoRequestWs(rCtx.Request, rCtx.Timeout, rCtx.Retry, rCtx.RetryRegex)
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

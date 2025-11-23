package requestHttp

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"testing"
)

func TestDoRequestFastHttp(t *testing.T) {
	req := &fuzzTypes.Req{
		URL: "https://www.baibaoxiang.vip/files",
		HttpSpec: fuzzTypes.HTTPSpec{
			RandomAgent: true,
		},
		Fields: nil,
		Data:   nil,
	}
	rc := &fuzzTypes.RequestCtx{
		Request:             req,
		Proxy:               "http://127.0.0.1:8080",
		Retry:               0,
		RetryCode:           "",
		RetryRegex:          "",
		Timeout:             10,
		HttpFollowRedirects: true,
	}
	resp, err := doRequestFastHttp(rc)
	fmt.Println(err)
	fmt.Println(string(resp.RawResponse))
	fmt.Println(resp.HttpRedirectChain)
}

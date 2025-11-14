package requestHttp

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"testing"
)

func TestRawRsp(t *testing.T) {
	fhResp := &fasthttp.Response{}
	fhResp.Header.Set("NISHIGIU", "WOSHIGIU")
	fhResp.Header.Set("MILAOGIU", "NISHIGIU")
	fhResp.SetBodyRaw([]byte("NISHIGIUWOSHGIUMILAOGIU"))
	r1, r2 := buildRawHTTPResponse1(fhResp)
	fmt.Println(string(r1))
	fmt.Println(string(r2))
}

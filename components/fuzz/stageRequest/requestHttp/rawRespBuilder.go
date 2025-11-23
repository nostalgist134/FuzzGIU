package requestHttp

import (
	"bytes"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"net/http"
)

func buildRawHTTPResponse(resp *http.Response) ([]byte, []byte, error) {
	if resp == nil {
		return nil, nil, nil
	}
	var raw bytes.Buffer

	// 状态行
	raw.WriteString(fmt.Sprintf("HTTP/%d.%d %s\r\n",
		resp.ProtoMajor, resp.ProtoMinor, resp.Status))

	// 响应头
	for k, vals := range resp.Header {
		for _, v := range vals {
			raw.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	raw.WriteString("\r\n")

	var bodyBytes []byte
	var err error
	// 响应体
	if resp.Body != nil {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, err
		}
		raw.Write(bodyBytes)

		// 重新填充 resp.Body 以便后续代码还能使用它
		resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return raw.Bytes(), bodyBytes, nil
}

// buildRawHTTPResponse1 将 fasthttp.Response 转为原始 HTTP 响应 []byte
func buildRawHTTPResponse1(resp *fasthttp.Response) ([]byte, []byte) {
	var buf bytes.Buffer

	// 写 header
	buf.Write(resp.Header.Header())
	// 写 body
	respBody := resp.Body()
	buf.Write(respBody)
	return buf.Bytes(), respBody
}

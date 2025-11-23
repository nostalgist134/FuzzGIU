package libfgiu

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"io"
	"net/url"
	"os"
	"strings"
)

// toHttpRequest 尝试将文件内容解析为http请求
func toHttpRequest(fileName string) (*fuzzTypes.Req, error) {
	raw, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewReader(raw))

	// 读取请求行
	startLine, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read request line: %w", err)
	}
	startLine = bytes.TrimRight(startLine, "\r\n")
	parts := bytes.Fields(startLine)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid request line: %q", startLine)
	}
	method := string(parts[0])
	path := string(parts[1])
	proto := string(parts[2])

	// 读取 headers
	var rawHeaders [][]byte
	var host string
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading headers: %w", err)
		}
		if len(line) > 5 && bytes.Equal(line[:5], []byte("Host:")) {
			host = strings.TrimSpace(string(line[6:]))
		}
		line = bytes.TrimRight(line, "\r\n")
		if len(line) == 0 {
			break // 空行：header 结束
		}
		rawHeaders = append(rawHeaders, line)
	}

	// 构造完整的URL（如果可以）
	if host != "" {
		u, err := url.Parse(path)
		if err != nil {
			u = &url.URL{}
			u.Path = path
			u.Host = host
		} else if u.Host == "" {
			u.Host = host
		}
		if u.Scheme == "" {
			u.Scheme = "http"
		}
		path = u.String()
	}

	// 读取 body
	bodyBuf := new(bytes.Buffer)
	if _, err := io.Copy(bodyBuf, reader); err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	body := bodyBuf.Bytes()

	// 构造 Req 结构体
	req := &fuzzTypes.Req{
		URL:  path,
		Data: body,
		HttpSpec: fuzzTypes.HTTPSpec{
			Method: method,
			Proto:  proto,
		},
	}

	// 处理 Headers，保存所有原始 header 行
	for _, h := range rawHeaders {
		req.HttpSpec.Headers = append(req.HttpSpec.Headers, string(h))
	}

	// 如果不是合法的 Request 请求，返回 error
	if req.HttpSpec.Method == "" || req.URL == "" {
		return nil, fmt.Errorf("invalid http request")
	}

	return req, nil
}

// toJsonRequest 尝试将文件内容解析为json格式的Req结构体
func toJsonRequest(fileName string) (*fuzzTypes.Req, error) {
	raw, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	req := new(fuzzTypes.Req)
	err = json.Unmarshal(raw, req)
	return req, err
}

func rawData(fileName string) ([]byte, error) {
	return os.ReadFile(fileName)
}

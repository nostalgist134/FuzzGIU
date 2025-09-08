package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"io"
	"os"
	"strings"
)

// parseHttpRequest 尝试将raw request解析为http请求
func parseHttpRequest(fileName string) (*fuzzTypes.Req, error) {
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
	version := string(parts[2])

	// 读取 headers
	var rawHeaders [][]byte
	var host string
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading headers: %w", err)
		}
		if len(line) > 5 && string(line[:5]) == "Host:" {
			host = strings.TrimSpace(string(line[6:]))
		}
		line = bytes.TrimRight(line, "\r\n")
		if len(line) == 0 {
			break // 空行：header 结束
		}
		rawHeaders = append(rawHeaders, line)
	}
	if len(path) < 7 || (path[:8] != "https://" || path[:7] != "http://") && host != "" { // 构造完整的URL
		if path[0] != '/' {
			path = "/" + path
		}
		path = "https://" + host + path
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
		Data: string(body),
		HttpSpec: fuzzTypes.HTTPSpec{
			Method:  method,
			Version: version,
		},
	}

	// 处理 Headers，保存所有原始 header 行
	for _, h := range rawHeaders {
		req.HttpSpec.Headers = append(req.HttpSpec.Headers, string(h))
	}

	// 判断是否需要强制 HTTPS (Host 字段判断)
	for _, h := range rawHeaders {
		if strings.HasPrefix(string(h), "Host:") {
			host := strings.TrimSpace(strings.TrimPrefix(string(h), "Host:"))
			if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
				req.HttpSpec.ForceHttps = true
			}
			break
		}
	}

	// 如果不是合法的 Request 请求，返回 error
	if req.HttpSpec.Method == "" || req.URL == "" {
		return nil, fmt.Errorf("invalid Request request")
	}

	return req, nil
}

func jsonRequest(fileName string) (*fuzzTypes.Req, error) {
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

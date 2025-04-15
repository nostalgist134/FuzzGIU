package common

import (
	"FuzzGIU/components/fuzzTypes"
	"sync"
)

var reqPool = sync.Pool{
	New: func() interface{} { return new(fuzzTypes.Req) },
}

func GetNewReq(originalReq *fuzzTypes.Req) *fuzzTypes.Req {
	newReq := (reqPool.Get()).(*fuzzTypes.Req)
	if originalReq == nil {
		return newReq
	}
	*newReq = fuzzTypes.Req{
		URL:  originalReq.URL,
		Data: originalReq.Data,
		HttpSpec: struct {
			Method     string   `json:"method"`
			Headers    []string `json:"headers"`
			Version    string   `json:"version"`
			ForceHttps bool     `json:"force_https"`
		}{
			Method: originalReq.HttpSpec.Method,
			// patchLog#4: 此处分配header时将从原req结构中复制而不是直接使用
			Headers:    append([]string{}, originalReq.HttpSpec.Headers...),
			Version:    originalReq.HttpSpec.Version,
			ForceHttps: originalReq.HttpSpec.ForceHttps,
		},
	}
	return newReq
}

func PutReq(r *fuzzTypes.Req) {
	reqPool.Put(r)
}

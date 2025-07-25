package common

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
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
		HttpSpec: fuzzTypes.HTTPSpec{
			Method:     originalReq.HttpSpec.Method,
			Version:    originalReq.HttpSpec.Version,
			ForceHttps: originalReq.HttpSpec.ForceHttps,
		},
	}
	//patchLog#4.1: 修改了判断Http头的逻辑，现在如果原请求头为nil，则新请求也会为nil，避免了append nil导致panic
	if originalReq.HttpSpec.Headers == nil {
		newReq.HttpSpec.Headers = nil
	} else {
		newReq.HttpSpec.Headers = append([]string{}, originalReq.HttpSpec.Headers...)
	}
	return newReq
}

func PutReq(r *fuzzTypes.Req) {
	reqPool.Put(r)
}

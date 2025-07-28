package common

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
)

var reqPool = sync.Pool{
	New: func() any { return new(fuzzTypes.Req) },
}

// GetNewReq 从池中获取一个新的Req结构，并使用originalReq复制
func GetNewReq() *fuzzTypes.Req {
	return (reqPool.Get()).(*fuzzTypes.Req)
}

// PutReq 放回用完的Req结构
func PutReq(r *fuzzTypes.Req) {
	reqPool.Put(r)
}

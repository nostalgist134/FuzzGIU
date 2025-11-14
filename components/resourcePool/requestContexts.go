package resourcePool

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
)

var rcPool = sync.Pool{
	New: func() any { return new(fuzzTypes.RequestCtx) },
}

// GetReqCtx 从池中获取一个新的taskCtx结构
func GetReqCtx() *fuzzTypes.RequestCtx {
	return (rcPool.Get()).(*fuzzTypes.RequestCtx)
}

// PutReqCtx TaskCtx回池
func PutReqCtx(toPut *fuzzTypes.RequestCtx) {
	if toPut == nil {
		return
	}
	*toPut = fuzzTypes.RequestCtx{}
	rcPool.Put(toPut)
}

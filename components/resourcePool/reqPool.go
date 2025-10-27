package resourcePool

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
)

var reqPool = sync.Pool{
	New: func() any { return new(fuzzTypes.Req) },
}

// GetReq 从池中获取一个新的Req结构
func GetReq() *fuzzTypes.Req {
	return (reqPool.Get()).(*fuzzTypes.Req)
}

// PutReq 放回用完的Req结构
func PutReq(toPut *fuzzTypes.Req) {
	if toPut == nil {
		return
	}
	StringSlices.Put(toPut.HttpSpec.Headers)
	FieldSlices.Put(toPut.Fields)
	*toPut = fuzzTypes.Req{}
	reqPool.Put(toPut)
}

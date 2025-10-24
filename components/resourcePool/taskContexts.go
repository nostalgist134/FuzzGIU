package resourcePool

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"sync"
)

var tcPool = sync.Pool{
	New: func() any { return new(fuzzCtx.TaskCtx) },
}

// GetTaskCtx 从池中获取一个新的taskCtx结构
func GetTaskCtx() *fuzzCtx.TaskCtx {
	return (tcPool.Get()).(*fuzzCtx.TaskCtx)
}

// PutTaskCtx TaskCtx回池
func PutTaskCtx(toPut *fuzzCtx.TaskCtx) {
	if toPut == nil {
		return
	}
	*toPut = fuzzCtx.TaskCtx{}
	reactionPool.Put(toPut)
}

package fuzzCtx

import (
	"sync"
)

var tcPool = sync.Pool{
	New: func() any { return new(TaskCtx) },
}

// GetTaskCtx 从池中获取一个新的taskCtx结构
func GetTaskCtx() *TaskCtx {
	return (tcPool.Get()).(*TaskCtx)
}

// PutTaskCtx TaskCtx回池
func PutTaskCtx(toPut *TaskCtx) {
	if toPut == nil {
		return
	}
	*toPut = TaskCtx{}
	tcPool.Put(toPut)
}

package fuzzCtx

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/tmplReplace"
	"sync"
)

// TaskCtx 单次循环上下文 inspired by ethread win kernel
type TaskCtx struct {
	IterInd      int
	SnipLen      int
	USchemeCache string
	Keywords     []string
	Payloads     []string
	RepTmpl      *tmplReplace.ReplaceTemplate
	JobCtx       *JobCtx
	PlProc       [][]fuzzTypes.Plugin
	ViaReaction  *fuzzTypes.Reaction
}

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

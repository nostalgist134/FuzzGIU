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
	toPut.Keywords = toPut.Keywords[:0]
	toPut.Payloads = toPut.Payloads[:0]
	toPut.PlProc = toPut.PlProc[:0]

	toPut.RepTmpl = nil
	toPut.JobCtx = nil
	toPut.ViaReaction = nil

	toPut.IterInd = 0
	toPut.SnipLen = 0
	toPut.USchemeCache = ""

	tcPool.Put(toPut)
}

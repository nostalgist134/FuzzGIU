package fuzzCtx

import (
	"context"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"sync"
)

type JobCtx struct {
	JobId     int
	ParentId  int
	Job       *fuzzTypes.Fuzz
	RP        IFaceRP // 改为使用interface，这样就避免引用rp包导致引用循环
	OutputCtx *output.Ctx
	GlobCtx   context.Context
	Cancel    context.CancelFunc
	occupied  sync.WaitGroup
}

func (jc *JobCtx) Close() error {
	jc.occupied.Wait()
	jc.RP.ReleaseSelf()
	return jc.OutputCtx.Close()
}

func (jc *JobCtx) Stop() {
	if jc.Cancel != nil {
		jc.Cancel()
	}
}

// Occupy 将jobCtx标记为占用，防止关闭
func (jc *JobCtx) Occupy() {
	jc.occupied.Add(1)
}

// Release 将jobCtx占用数减一
func (jc *JobCtx) Release() {
	jc.occupied.Done()
}

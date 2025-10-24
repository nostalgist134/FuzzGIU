package fuzzCtx

import (
	"context"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
)

type JobCtx struct {
	JobId     int
	ParentId  int
	Job       *fuzzTypes.Fuzz
	RP        IFaceRP // 改为使用interface，这样就避免引用rp包导致引用循环
	OutputCtx *output.Ctx
	GlobCtx   context.Context
	Cancel    context.CancelFunc
}

func (jc *JobCtx) Close() error {
	jc.RP.ReleaseSelf()
	return jc.OutputCtx.Close()
}

func (jc *JobCtx) Stop() {
	if jc.Cancel != nil {
		jc.Cancel()
	}
}

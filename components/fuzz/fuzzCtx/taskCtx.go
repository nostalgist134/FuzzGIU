package fuzzCtx

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/tmplReplace"
)

// TaskCtx 用于描述单次循环中的环境 inspired by ethread win kernel
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

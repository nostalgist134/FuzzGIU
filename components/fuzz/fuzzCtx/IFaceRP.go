package fuzzCtx

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"time"
)

// IFaceRP 协程池接口，用于避免循环import
type IFaceRP interface {
	Submit(*TaskCtx, int8, time.Duration) bool
	RegisterExecutor(func(*TaskCtx) *fuzzTypes.Reaction, int8)
	Wait(time.Duration) bool
	Status() int8
	Start()
	Stop()
	Pause()
	Resume()
	GetSingleResult() *fuzzTypes.Reaction
	Resize(int)
	WaitResume()
	Clear()
	ReleaseSelf()
}

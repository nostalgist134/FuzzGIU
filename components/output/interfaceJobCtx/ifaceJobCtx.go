package interfaceJobCtx

import "github.com/nostalgist134/FuzzGIU/components/fuzzTypes"

type IFaceJobCtx interface {
	Close() error
	Stop()
	Pause()
	Resume()
	Occupy()
	Release()
	GetJobInfo() *fuzzTypes.Fuzz
}

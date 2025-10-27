package outCtx

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	co "github.com/nostalgist134/FuzzGIU/components/output/chanOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	fo "github.com/nostalgist134/FuzzGIU/components/output/fileOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/httpOutput"
	stdout "github.com/nostalgist134/FuzzGIU/components/output/stdoutOutput"
	tOut "github.com/nostalgist134/FuzzGIU/components/output/tviewOutput"
)

type OutputCtx struct {
	Id             int
	TviewOutputCtx *tOut.Ctx
	FileOutputCtx  *fo.Ctx
	ChanOutputCtx  *co.Ctx
	StdoutCtx      *stdout.Ctx
	HttpCtx        *httpOutput.Ctx
	OutSetting     fuzzTypes.OutputSetting
	Counter        *counter.Counter
}

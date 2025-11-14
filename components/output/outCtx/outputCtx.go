package outCtx

import (
	co "github.com/nostalgist134/FuzzGIU/components/output/chanOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	fo "github.com/nostalgist134/FuzzGIU/components/output/fileOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/httpOutput"
	stdout "github.com/nostalgist134/FuzzGIU/components/output/stdoutOutput"
	tvw "github.com/nostalgist134/FuzzGIU/components/output/tviewOutput"
	"sync"
)

type OutputCtx struct {
	Id             int
	TviewOutputCtx *tvw.Ctx
	FileOutputCtx  *fo.Ctx
	ChanOutputCtx  *co.Ctx
	StdoutCtx      *stdout.Ctx
	HttpCtx        *httpOutput.Ctx
	ToWhere        int32
	Counter        *counter.Counter
	Wg             sync.WaitGroup
}

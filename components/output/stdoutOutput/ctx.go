package stdoutOutput

import (
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
)

type Ctx struct {
	id        int
	closed    bool
	outputFmt string
	cntrStop  chan struct{}
	cntrReg   chan struct{}
	okToClose chan struct{}
	counter   *counter.Counter
}

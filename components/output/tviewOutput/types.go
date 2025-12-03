package tviewOutput

import (
	"context"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/output/interfaceJobCtx"
	"github.com/rivo/tview"
	"sync"
	"sync/atomic"
)

type Ctx struct {
	id           int
	job          *fuzzTypes.Fuzz
	app          *tview.Application
	flx          *tview.Flex
	textViews    []*tview.TextView
	focus        int
	counter      *counter.Counter
	jobCtx       interfaceJobCtx.IFaceJobCtx
	outputFormat string
	verbosity    int
	closed       bool
	endCounter   chan struct{}
	startCounter chan struct{}
	lockOnLog    atomic.Bool
	lockOnOutput atomic.Bool
	tagName      string
}

type tviewScreen struct {
	wg                  sync.WaitGroup
	wgAdded             atomic.Int64
	tviewApp            *tview.Application
	pages               *tview.Pages
	listJobs            *tview.List
	listFlx             *tview.Flex
	activeTviewCtxs     sync.Map
	ctx                 context.Context
	cancel              context.CancelFunc
	listJobsNameToIndex sync.Map
}

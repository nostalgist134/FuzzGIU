package tviewOutput

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/output/interfaceJobCtx"
	"github.com/rivo/tview"
)

type Ctx struct {
	id        int
	job       *fuzzTypes.Fuzz
	app       *tview.Application
	flx       *tview.Flex
	textViews []*tview.TextView
	counter   *counter.Counter
	jobCtx    interfaceJobCtx.IFaceJobCtx
	closed    bool
}

type tviewScreen struct {
	tviewApp  *tview.Application
	pages     *tview.Pages
	pageNames []string
	list      *tview.List
}

package tviewOutput

import (
	"github.com/gizak/termui/v3/widgets"
	"sync"
)

type screenOutputRegion struct {
	Pg             *widgets.Paragraph
	lines          []string
	mu             sync.Mutex
	lineInd        int
	lineLeft       int
	maxRenderLines int
	rendered       bool
	renderBuffer   []string
	TopCorner      struct {
		X int
		Y int
	}
	BottomCorner struct {
		X int
		Y int
	}
}

type Screen struct {
	renderMu     sync.Mutex
	logo         *screenOutputRegion
	globInfo     screenOutputRegion
	counterFrame screenOutputRegion
	outputs      screenOutputRegion
	logs         screenOutputRegion
	selectInd    int
}

type Ctx struct {
}

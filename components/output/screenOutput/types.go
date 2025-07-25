package output

import (
	"github.com/gizak/termui/v3/widgets"
	"sync"
)

type screenOutputRegion struct {
	Pg             *widgets.Paragraph
	lines          []string
	mu             sync.Mutex
	lineInd        int
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

type screenOutputFrame struct {
	renderMu      sync.Mutex
	logo          screenOutputRegion
	globInfo      screenOutputRegion
	counterFrame  screenOutputRegion
	outputObjects screenOutputRegion
	logs          screenOutputRegion
}

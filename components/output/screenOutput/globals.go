package output

import (
	"sync"
	"sync/atomic"
)

const (
	logoMaxLines     = 8
	globInfoMaxLines = 4
	counterMaxLines  = 1
	outputMaxLines   = 8
	logMaxLines      = 2
	leastHeight      = logMaxLines + globInfoMaxLines + counterMaxLines + outputMaxLines + logoMaxLines + 3*5 - 4

	titleGlobInfo      = "GLOBAL_INFORMATION"
	titleOutput        = "OUTPUT"
	titleCounter       = "PROGRESS"
	titlePausedCounter = "PROGRESS(PAUSED)"
	titleLogger        = "LOGS"
	titleLockedOutput  = "OUTPUT(LOCKED)"

	directionUp    = int8(0)
	directionDown  = int8(1)
	directionLeft  = int8(2)
	directionRight = int8(3)

	selectGlobInfo = 0
	selectOutput   = 1
	selectLogs     = 2
)

// screenOutput 输出屏幕
var screenOutput *screenOutputFrame

// selectableRegions 标识可以被选中（可以上下滑动）的输出区域
var selectableRegions []*screenOutputRegion

var indSelect = 0

var logo = "     GIUGIUGIUGIU                            GIUGIUGIUGI GIUGIUGI     GI\n      GI                                    IU            GI  UG     UG\n     UG                                    GI            UG  IU     IU\n    IUGIUGIUGGI   IU#GIUGIUGIU#GIUGIUGIU# UG            IU  GI     GI\n   GI       UG   UG       GIU       GIU  IU       GIU  GI  UG     UG\n  UG       IU   GI      GIU       GIU   GI        IU  UG  IU     IU\n IU       GI   IU     GIU       GIU    UG        UG  IU  GI     GI\nGIU        UGIUGIU GIUGIUGIU#GIUGIUGIU#IUGIUGIUGIU GIUGIU UGIUGIU"
var counterFmt = "task: %d/%d    job: %d/%d    rate: %dr/s    duration: [%02d:%02d:%02d]"
var hintWindowTooSmall = "THE WINDOW SEEMS TOO SMALL TO DISPLAY ALL INFORMATION, RECOMMEND RESIZE"

var posLogo = []int{0, 0, 0, logoMaxLines + 2}

var outputHasInit = atomic.Bool{}
var outputLocked = true
var hasOutput = false

var wg = sync.WaitGroup{}

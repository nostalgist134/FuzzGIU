package tviewOutput

import (
	"github.com/rivo/tview"
	"sync"
)

var indSelect = 0

var logo = "     GIUGIUGIUGIU                            GIUGIUGIUGI GIUGIUGI     GI\n      GI                                    IU            GI  UG     UG\n     UG                                    GI            UG  IU     IU\n    IUGIUGIUGGI   IU#GIUGIUGIU#GIUGIUGIU# UG            IU  GI     GI\n   GI       UG   UG       GIU       GIU  IU       GIU  GI  UG     UG\n  UG       IU   GI      GIU       GIU   GI        IU  UG  IU     IU\n IU       GI   IU     GIU       GIU    UG        UG  IU  GI     GI\nGIU        UGIUGIU GIUGIUGIU#GIUGIUGIU#IUGIUGIUGIU GIUGIU UGIUGIU"

var (
	proportions   = []int{6, 2, 9, 2}
	titles        = []string{"JOB_INFO", "OUTPUT", "PROGRESS", "LOGS"}
	wg            = sync.WaitGroup{}
	tviewApp      *tview.Application
	appCreateOnce = sync.Once{}
	screen        *tviewScreen
)

package output

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/nostalgist134/FuzzGIU/components/rp"
	"os"
)

// eventListener 监听事件
func eventListener() {
	defer func() {
		ScreenClose()
		fmt.Printf("Now exitting...")
		wg.Done()
		os.Exit(0)
	}()
	for e := range ui.PollEvents() {
		switch e.ID {
		case "w":
			if indSelect > 0 {
				indSelect--
				switchHighLightRegion(indSelect + 1)
			}
		case "s":
			if indSelect < len(selectableRegions)-1 {
				indSelect++
				switchHighLightRegion(indSelect - 1)
			}
		case "L":
			screenOutput.outputObjects.mu.Lock()
			if !outputLocked {
				if len(screenOutput.outputObjects.lines)-outputMaxLines-1 < 0 {
					screenOutput.outputObjects.lineInd = 0
				} else {
					screenOutput.outputObjects.lineInd = len(screenOutput.outputObjects.lines) - outputMaxLines - 1
				}
				screenOutput.outputObjects.render(titleLockedOutput, true)
			} else {
				screenOutput.outputObjects.render(titleOutput, true)
			}
			outputLocked = !outputLocked
			screenOutput.outputObjects.mu.Unlock()
		case "<Up>", "k":
			selectableRegions[indSelect].scroll(directionUp)
		case "<Down>", "j":
			selectableRegions[indSelect].scroll(directionDown)
		case "<Left>", "h":
			selectableRegions[indSelect].scroll(directionLeft)
		case "<Right>", "l":
			selectableRegions[indSelect].scroll(directionRight)
		case "c":
			if indSelect != selectGlobInfo {
				selectableRegions[indSelect].clear()
			}
		case "q":
			return
		case "p":
			if rp.CurrentRp != nil {
				rp.CurrentRp.Pause()
				screenOutput.counterFrame.render(titlePausedCounter)
			}
		case "r":
			if rp.CurrentRp != nil {
				rp.CurrentRp.Resume()
				screenOutput.counterFrame.render(titleCounter)
			}
		case "<Resize>":
			// 命令行窗口大小被调整，调整宽度并重新渲染全部窗口
			w, h := ui.TerminalDimensions()
			screenOutput.logo.mu.Lock()
			if h >= leastHeight {
				screenOutput.logo.Pg.Title = ""
			} else if screenOutput.logo.Pg.Title == "" {
				screenOutput.logo.Pg.Title = hintWindowTooSmall
			}
			logoLines := splitLines(logo)
			screenOutput.logo.lines = logoLines
			centeredLines(screenOutput.logo.lines, w)
			screenOutput.logo.mu.Unlock()
			screenOutput.logo.render("")
			screenOutput.counterFrame.setRect([]int{-1, -1, w, -1})
			screenOutput.counterFrame.render("")
			screenOutput.outputObjects.setRect([]int{-1, -1, w, -1})
			screenOutput.outputObjects.render("")
			screenOutput.logs.setRect([]int{-1, -1, w, -1})
			screenOutput.logs.render("")
			screenOutput.globInfo.setRect([]int{-1, -1, w, -1})
			screenOutput.globInfo.render("")
		}
	}
}

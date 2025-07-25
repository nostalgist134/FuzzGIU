package output

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/common"
	"time"
)

func Log(log string) {
	if !outputHasInit.Load() {
		return
	}
	if firstLog {
		screenOutput.logs.clear()
		firstLog = false
	}
	screenOutput.logs.mu.Lock()
	defer screenOutput.logs.mu.Unlock()
	screenOutput.logs.append(splitLines(log))
	lenLines := len(screenOutput.logs.lines)
	if lenLines >= 2 {
		screenOutput.logs.lineInd = lenLines - 2
	} else {
		screenOutput.logs.lineInd = 0
	}
	screenOutput.logs.render("", true)
}

func renderCounter() {
	screenOutput.counterFrame.mu.Lock()
	defer screenOutput.counterFrame.mu.Unlock()
	c := common.GetCounter()
	timeLapsed := common.GetTimeLapsed()
	h := int(timeLapsed.Hours())
	m := int(timeLapsed.Minutes()) % 60
	s := int(timeLapsed.Seconds()) % 60
	screenOutput.counterFrame.lines = []string{fmt.Sprintf(counterFmt, c[0], c[1], c[2], c[3], common.GetCurrentRate(),
		h, m, s)}
	screenOutput.counterFrame.render("", true)
}

func getNextParaPos(pos []int, maxLines int) []int {
	if len(pos) < 4 {
		return nil
	}
	return []int{0, pos[3] + 1, pos[2], pos[3] + maxLines + 3}
}

// InitOutputScreen 初始化输出窗口
func InitOutputScreen(globInfo *fuzzTypes.Fuzz) {
	HasInit := outputHasInit.Load()
	if !HasInit {
		outputHasInit.Store(true)
		if err := ui.Init(); err != nil {
			fmt.Printf("%v\n", err)
		}
	}
	w, h := ui.TerminalDimensions()
	if !HasInit {
		screenOutput = new(screenOutputFrame)
		screenOutput.logo.init(logoMaxLines, true)
		screenOutput.logs.init(logMaxLines)
		screenOutput.outputObjects.init(outputMaxLines)
		screenOutput.counterFrame.init(counterMaxLines)
		screenOutput.globInfo.init(globInfoMaxLines)
	}
	// 渲染logo
	posLogo[2] = w
	logoLines := splitLines(logo)
	centeredLines(logoLines, w)
	screenOutput.logo.lines = logoLines
	screenOutput.logo.setRect(posLogo)
	if h < leastHeight {
		screenOutput.logo.render(hintWindowTooSmall)
	} else {
		screenOutput.logo.render("")
	}
	// 渲染全局信息窗口
	posGlobInfo := getNextParaPos(posLogo, globInfoMaxLines)
	screenOutput.globInfo.setLines(genInfoLines(globInfo))
	screenOutput.globInfo.setRect(posGlobInfo)
	screenOutput.globInfo.render(titleGlobInfo)
	// 渲染计数器窗口
	counterPos := getNextParaPos(posGlobInfo, counterMaxLines)
	screenOutput.counterFrame.setRect(counterPos)
	screenOutput.counterFrame.render(titleCounter)
	// 渲染输出窗口
	outputPos := getNextParaPos(counterPos, outputMaxLines)
	screenOutput.outputObjects.setRect(outputPos)
	screenOutput.outputObjects.render(titleLockedOutput)
	// 渲染日志记录窗口
	logPos := getNextParaPos(outputPos, logMaxLines)
	screenOutput.logs.setRect(logPos)
	if !HasInit {
		screenOutput.logs.append([]string{"W/S to select window to control, <Up>/K to scroll up, <Down>/J to scroll " +
			"down, Q to quit.", "Output window will lock to the latest output object by default, L to unlock it"})
	}
	screenOutput.logs.render(titleLogger)
	if !HasInit {
		// 设置选中的窗口为可选窗口数组中的第一个，并渲染
		selectableRegions = []*screenOutputRegion{&screenOutput.globInfo, &screenOutput.outputObjects,
			&screenOutput.logs}
		switchHighLightRegion(-1)
		go func() {
			wg.Add(1)
			eventListener()
		}()
		// 每0.15秒更新一次计数器
		go func() {
			wg.Add(1)
			for {
				if !outputHasInit.Load() {
					wg.Done()
					return
				}
				renderCounter()
				time.Sleep(200 * time.Millisecond)
			}
		}()
	}
}

func ScreenObjOutput(obj *common.OutObj) {
	if !outputHasInit.Load() {
		return
	}
	screenOutput.outputObjects.mu.Lock()
	defer screenOutput.outputObjects.mu.Unlock()
	output := common.FormatObjOutput(obj, "native", true)
	lines := splitLines(string(output))
	screenOutput.outputObjects.append(lines)
	if outputLocked && hasOutput {
		screenOutput.outputObjects.lineInd = len(screenOutput.outputObjects.lines) - len(lines)
	}
	hasOutput = true
	screenOutput.outputObjects.render("", true)
}

func FinishOutputScreen() {
	if !outputHasInit.Load() {
		return
	}
	screenOutput.globInfo.clear()
}

func WaitForScreenQuit() {
	wg.Wait()
}

func ScreenClose() {
	outputHasInit.Store(false)
	screenOutput.renderMu.Lock()
	defer screenOutput.renderMu.Unlock()
	ui.Close()
}

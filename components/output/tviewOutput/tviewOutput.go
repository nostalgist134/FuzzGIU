package tviewOutput

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/output/interfaceJobCtx"
	"github.com/nostalgist134/FuzzGIU/components/output/outputErrors"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/rivo/tview"
	"log"
)

func vimKey(k *tcell.EventKey) *tcell.EventKey {
	switch k.Name() {
	case "h":
		return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
	case "j":
		return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	case "k":
		return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
	case "l":
		return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
	}
	return k
}

func newTextViewAndFlex(jobInfo *fuzzTypes.Fuzz) (textViews []*tview.TextView, flx *tview.Flex) {
	textViews = make([]*tview.TextView, 4)
	flx = tview.NewFlex()
	for i, _ := range textViews {
		textViews[i] = tview.NewTextView()
		textViews[i].SetFocusFunc(func() { // 为textView设置选中时边框变为蓝色
			textViews[i].SetBorderColor(tcell.ColorBlue)
		})
		textViews[i].SetBlurFunc(func() { // 未选择是变回白色
			textViews[i].SetBorderColor(tcell.ColorWhite)
		})
		textViews[i].SetInputCapture(vimKey) // 为每个textView添加vim风格按键映射
		if i == 0 {
			textViews[i].SetText(stringifyJobInfo(jobInfo))
		} else if i == 2 {
			textViews[i].SetDynamicColors(true)
		}
		textViews[i].SetTitle(titles[i]).SetTitleAlign(tview.AlignLeft).SetBorder(true)
		flx.AddItem(textViews[i], 0, proportions[i], i == 0)
	}
	return
}

func NewTviewOutputCtx(outSetting *fuzzTypes.OutputSetting, jobCtx interfaceJobCtx.IFaceJobCtx, id int) (*Ctx, error) {
	if outSetting.ToWhere|outputFlag.OutToStdout != 0 {
		return nil, outputErrors.ErrTviewConflict
	}

	// 创建tviewApplication，只执行一次
	appCreateOnce.Do(func() {
		if tviewApp == nil {
			tviewApp = tview.NewApplication()
		}
		screen = &tviewScreen{
			tviewApp:  tviewApp,
			pages:     tview.NewPages(),
			pageNames: []string{"list-jobs"},
			list:      tview.NewList(),
		}
		screen.pages.AddPage(screen.pageNames[0], screen.list, true, true)
		tviewApp.SetRoot(screen.pages, false)
		go func() {
			wg.Add(1)
			defer wg.Done()
			if err := tviewApp.Run(); err != nil {
				log.Fatal(err)
			}
		}()
	})

	textViews, flx := newTextViewAndFlex(jobCtx.GetJobInfo())

	screen.tviewApp.QueueUpdate(func() {
		pageName := fmt.Sprintf("job#%d", id)
		screen.list.AddItem(pageName, "", 0, func() {
			screen.pages.SwitchToPage(pageName)
		})
	})
	return nil, nil
}

func (c *Ctx) Output() {

}

func (c *Ctx) Close() {

}

func WaitTviewQuit() {
	wg.Wait()
}

func StopTviewOutput() {
	tviewApp.Stop()
}

func (c *Ctx) RegisterCounter(counter *counter.Counter) {
}

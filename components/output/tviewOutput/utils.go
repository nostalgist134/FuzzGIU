package tviewOutput

import (
	"github.com/gdamore/tcell/v2"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/rivo/tview"
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
	flx = tview.NewFlex().SetDirection(tview.FlexRow).SetFullScreen(true)
	for i, _ := range textViews {
		textViews[i] = tview.NewTextView()
		textViews[i].SetFocusFunc(func() { // 为textView设置选中时边框变为蓝色
			textViews[i].SetBorderColor(tcell.ColorBlue)
		})
		textViews[i].SetBlurFunc(func() { // 未选择变回白色
			textViews[i].SetBorderColor(tcell.ColorWhite)
		})
		textViews[i].SetInputCapture(vimKey) // 为每个textView添加vim风格按键映射
		switch i {
		case IndJobInfo:
			textViews[i].SetText(stringifyJobInfo(jobInfo))
		case IndOutput:
			textViews[i].SetDynamicColors(true)
		case IndLogs:
		default:
		}
		// 标题统一采用左对齐，并且统一采用边框
		textViews[i].SetTitle(titles[i]).SetTitleAlign(tview.AlignLeft).SetBorder(true)
		// 将textView对象添加到flex中，并且设置第一个为聚焦对象(i==0)
		flx.AddItem(textViews[i], 0, proportions[i], i == 0)
	}
	return
}

func getListItemByName(lst *tview.List, name string) int {
	if lst == nil {
		return -1
	}
	for i := 0; i < lst.GetItemCount(); i++ {
		if mainName, _ := lst.GetItemText(i); mainName == name {
			return i
		}
	}
	return -1
}

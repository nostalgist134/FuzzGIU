package tviewOutput

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/output/interfaceJobCtx"
	"github.com/nostalgist134/FuzzGIU/components/output/outputErrors"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"github.com/rivo/tview"
	"log"
	"time"
)

func initOnce() {
	screen = &tviewScreen{
		tviewApp:  tview.NewApplication(),
		pages:     tview.NewPages(),
		pageNames: []string{"list-jobs"},
		listJobs:  tview.NewList(),
		listFlx:   tview.NewFlex(),
	}
	screen.listJobs.SetInputCapture(vimKey)
	screen.listJobs.SetBorder(true)
	screen.listJobs.SetTitle("RUNNING_JOBS").SetTitleAlign(tview.AlignLeft)
	screen.listFlx.SetFullScreen(true)
	screen.listFlx.AddItem(screen.listJobs, 0, 1, true)
	// 目录页
	screen.pages.AddPage(screen.pageNames[0], screen.listFlx, true, true)
	screen.pages.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		switch key.Key() {
		case tcell.KeyCtrlC: // 按下ctrl+c就退出
			QuitTview()
			return nil
		case tcell.KeyCtrlR: // 按下ctrl+r就切换回目录页
			screen.pages.SwitchToPage(screen.pageNames[0])
			screen.tviewApp.SetFocus(screen.listJobs)
			return nil
		default:
		}
		return key
	})
	screen.tviewApp.SetRoot(screen.pages, false)
	go func() {
		screen.wg.Add(1)
		defer screen.wg.Done()
		if err := screen.tviewApp.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		for {
			screen.tviewApp.Draw()
			time.Sleep(15 * time.Millisecond)
		}
	}()
}

// NewTviewOutputCtx 创建一个新的tview子窗口
func NewTviewOutputCtx(outSetting *fuzzTypes.OutputSetting, jobCtx interfaceJobCtx.IFaceJobCtx, id int) (*Ctx, error) {
	if outSetting.ToWhere&outputFlag.OutToStdout != 0 {
		return nil, outputErrors.ErrTviewConflict
	}

	// 创建tviewApplication，只执行一次。这里采用initOnce而不是直接使用init函数的原因
	// 在于：工具不一定使用这个输出流，如果使用init函数就会导致无论使用与否都会创建tview窗
	// 口，这显然是不符合常理的，因此只有显式调用这个函数才会使用
	appCreateOnce.Do(initOnce)

	textViews, flx := newTextViewAndFlex(jobCtx.GetJobInfo())

	tviewCtx := &Ctx{
		app:          screen.tviewApp,
		textViews:    textViews,
		flx:          flx,
		outputFormat: "native",
		verbosity:    outSetting.Verbosity,
		startCounter: make(chan struct{}),
		endCounter:   make(chan struct{}),
		jobCtx:       jobCtx,
		id:           id,
	}
	tviewCtx.occupied.Add(1)

	flx.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		switch key.Name() {
		case "Ctrl+Up", "Ctrl+K", "Ctrl+W": // 切换到上一个窗口
			if tviewCtx.focus > 0 {
				tviewCtx.focus--
				tviewCtx.app.SetFocus(tviewCtx.textViews[tviewCtx.focus])
			}
			return nil
		case "Ctrl+Down", "Ctrl+J", "Ctrl+S": // 切换到下一个窗口
			if tviewCtx.focus < len(tviewCtx.textViews)-1 {
				tviewCtx.focus++
				tviewCtx.app.SetFocus(tviewCtx.textViews[tviewCtx.focus])
			}
			return nil
		case "Rune[p]":
			tviewCtx.textViews[IndCounter].SetTitle(titles[IndCounterPaused])
			tviewCtx.jobCtx.Pause()
			return nil
		case "Rune[r]":
			tviewCtx.textViews[IndCounter].SetTitle(titles[IndCounter])
			tviewCtx.jobCtx.Resume()
			return nil
		case "Rune[q]":
			tviewCtx.quitOnce.Do(func() {
				tviewCtx.occupied.Done() // 实现在任务全部完成后不自动退出，而是等待用户按下q退出
			})
			tviewCtx.jobCtx.Stop()
			tviewCtx.jobCtx.Resume()
			return nil
		}
		return key
	})

	tviewCtx.textViews[IndOutput].SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		switch key.Key() {
		case tcell.KeyCtrlL:
			tviewCtx.lockOnOutput.Store(true)
			tviewCtx.textViews[IndOutput].SetTitle(titles[IndOutputLocked])
			tviewCtx.textViews[IndOutput].ScrollToEnd()
			return nil
		case tcell.KeyCtrlU:
			tviewCtx.lockOnOutput.Store(false)
			tviewCtx.textViews[IndOutput].SetTitle(titles[IndOutput])
			return nil
		default:
			if key.Name() == "Rune[c]" {
				tviewCtx.textViews[IndOutput].SetText("")
				return nil
			}
			return key
		}
	})

	tviewCtx.textViews[IndLogs].SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		switch key.Key() {
		case tcell.KeyCtrlL:
			tviewCtx.lockOnLog.Store(true)
			tviewCtx.textViews[IndLogs].SetTitle(titles[IndLogsLocked])
			tviewCtx.textViews[IndLogs].ScrollToEnd()
			return nil
		case tcell.KeyCtrlU:
			tviewCtx.lockOnLog.Store(false)
			tviewCtx.textViews[IndLogs].SetTitle(titles[IndLogs])
			return nil
		default:
			if key.Name() == "Rune[c]" {
				tviewCtx.textViews[IndLogs].SetText("")
				return nil
			}
			return key
		}
	})

	flx.SetFocusFunc(func() { // 与下面的函数配合，当flx被聚焦时，将聚焦传递到上次选中的textView上
		tviewCtx.app.SetFocus(tviewCtx.textViews[tviewCtx.focus])
	})

	tviewCtx.app.QueueUpdate(func() { // tview作者真是神人，有的组件方法就加锁，有的就不加，我还得自己进去看再处理
		pageName := fmt.Sprintf("job#%d", id)
		screen.pages.AddPage(pageName, flx, false, true) // 添加一个名为job#id的页
		screen.listJobs.AddItem(pageName, "", 0, func() {
			// 选中对应list时，切换到这个页，并聚焦到flx上
			screen.pages.SwitchToPage(pageName)
			tviewCtx.app.SetFocus(tviewCtx.flx)
		})
	})

	go func() {
		<-tviewCtx.startCounter
		for {
			select {
			case <-tviewCtx.endCounter:
				return
			default:
				tviewCtx.textViews[IndCounter].SetText(tviewCtx.counter.ToFmt())
				time.Sleep(225 * time.Millisecond)
			}
		}
	}()
	return tviewCtx, nil
}

func (c *Ctx) Output(o *outputable.OutObj) error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	_, err := fmt.Fprintln(c.textViews[IndOutput], o.ToFormatStr(c.outputFormat, true, c.verbosity))
	if err != nil {
		return err
	}
	if c.lockOnOutput.Load() { // 其实用mutex+bool会更好，不过我懒得写那么多了，而且如果这里用了log也要用
		c.textViews[IndOutput].ScrollToEnd()
	}
	return nil
}

func (c *Ctx) Log(l *outputable.Log) error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	_, err := fmt.Fprintln(c.textViews[IndLogs], l.ToFormatStr(c.outputFormat))
	if err != nil {
		return err
	}
	if c.lockOnLog.Load() {
		c.textViews[IndLogs].ScrollToEnd()
	}
	return nil
}

func (c *Ctx) Close() error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	c.flx.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		switch key.Name() {
		case "Ctrl+Up", "Ctrl+K", "Ctrl+W": // 切换到上一个窗口
			if c.focus > 0 {
				c.focus--
				c.app.SetFocus(c.textViews[c.focus])
			}
			return nil
		case "Ctrl+Down", "Ctrl+J", "Ctrl+S": // 切换到下一个窗口
			if c.focus < len(c.textViews)-1 {
				c.focus++
				c.app.SetFocus(c.textViews[c.focus])
			}
			return nil
		case "Rune[q]":
			c.quitOnce.Do(func() {
				c.occupied.Done() // 实现在任务全部完成后不自动退出，而是等待用户按下q退出
			})
			return nil
		}
		return key
	})
	screen.tviewApp.QueueUpdate(func() { // 在菜单页面将对应的项标记为完成
		itemInd := getListItemByName(screen.listJobs, fmt.Sprintf("job#%d", c.id))
		if itemInd == -1 {
			return
		}
		_, secondary := screen.listJobs.GetItemText(itemInd)
		screen.listJobs.SetItemText(itemInd, fmt.Sprintf("job#%d(done)", c.id), secondary)
	})
	logJobDone := &outputable.Log{
		Msg:  "job is already done, press q to quit",
		Time: time.Now(),
	}
	c.Log(logJobDone)
	// 等待直到用户按下q键退出
	c.occupied.Wait()
	screen.tviewApp.QueueUpdate(func() {
		itemInd := getListItemByName(screen.listJobs, fmt.Sprintf("job#%d(done)", c.id))
		if itemInd == -1 {
			return
		}
		screen.pages.SwitchToPage("list-jobs")
		screen.tviewApp.SetFocus(screen.listJobs) // 回到菜单页面
		screen.listJobs.RemoveItem(itemInd)       // 将当前项从list中移除
	})
	c.endCounter <- struct{}{}
	if c.counter == nil {
		c.startCounter <- struct{}{}
	}
	c.closed = true
	return nil
}

func (c *Ctx) RegisterCounter(counter *counter.Counter) error {
	if counter == nil {
		return outputErrors.ErrRegisterNilCounter
	} else if c.closed {
		return outputErrors.ErrCtxClosed
	}
	c.counter = counter
	c.startCounter <- struct{}{}
	return nil
}

func QuitTview() {
	appStopOnce.Do(func() {
		if screen != nil && screen.tviewApp != nil {
			screen.tviewApp.Stop()
		}
	})
}

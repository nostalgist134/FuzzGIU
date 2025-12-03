package tviewOutput

// 恭喜此包成为整个项目中最复杂的包，超过tmplReplace与plugin，单独这一个包就700+行，难绷

import (
	"context"
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
	"os"
	"sync/atomic"
	"time"
)

func quitAll() {
	screen.activeTviewCtxs.Range(func(_, v any) bool {
		if c, ok := v.(*Ctx); ok && c.jobCtx != nil {
			c.jobCtx.Stop()
		}
		return true
	})
	screen.activeTviewCtxs.Clear()
	for i := screen.wgAdded.Load(); i > 0; i = screen.wgAdded.Load() {
		screen.wg.Done() // 真是他妈的傻逼，能加能减能等，为什么不能查看大小
		screen.wgAdded.Add(-1)
	}
	if screen.tviewApp != nil {
		screen.tviewApp.Stop()
	}
	os.Exit(0)
}

func initOnce() {
	screen = &tviewScreen{
		tviewApp: tview.NewApplication(),
		pages:    tview.NewPages(),
		listJobs: tview.NewList(),
		listFlx:  tview.NewFlex(),
	}
	screen.ctx, screen.cancel = context.WithCancel(context.Background())
	screen.tviewApp.SetInputCapture(func(k *tcell.EventKey) *tcell.EventKey {
		if k.Key() == tcell.KeyCtrlC {
			quitAll()
			return nil
		}
		return k
	})
	screen.listJobs.SetInputCapture(vimKey)
	screen.listJobs.SetBorder(true)
	screen.listJobs.SetTitle("JOBS").SetTitleAlign(tview.AlignLeft)
	screen.listJobs.SetSecondaryTextColor(tcell.ColorWhite)
	screen.listFlx.SetFullScreen(true)
	screen.listFlx.AddItem(screen.listJobs, 0, 1, true)
	// 目录页
	screen.pages.AddPage("list-jobs", screen.listFlx, true, true)
	screen.pages.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
		switch key.Key() {
		case tcell.KeyCtrlR: // 按下ctrl+r就切换回目录页
			screen.pages.SwitchToPage("list-jobs")
			screen.tviewApp.SetFocus(screen.listJobs)
			return nil
		default:
			return key
		}
	})
	screen.tviewApp.SetRoot(screen.pages, false)
	go func() {
		if err := screen.tviewApp.Run(); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		for {
			select {
			case <-screen.ctx.Done():
				return
			default:
				screen.tviewApp.Draw()
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()
}

func lockableInputCap(tviewCtx *Ctx, ind, lockedInd int,
	atomicBool *atomic.Bool) func(key *tcell.EventKey) *tcell.EventKey {
	return func(key *tcell.EventKey) *tcell.EventKey {
		switch key.Key() {
		case tcell.KeyCtrlL:
			atomicBool.Store(true)
			tviewCtx.textViews[ind].SetTitle(titles[lockedInd])
			tviewCtx.textViews[ind].ScrollToEnd()
			return nil
		case tcell.KeyCtrlU:
			atomicBool.Store(false)
			tviewCtx.textViews[ind].SetTitle(titles[ind])
			return nil
		default:
			if key.Name() == "Rune[c]" {
				tviewCtx.textViews[ind].SetText("")
				return nil
			}
			return key
		}
	}
}

func counterProgress(c *counter.Counter) string {
	s := c.Snapshot()
	if s.Errors.Completed == 0 {
		return fmt.Sprintf(
			"  [#2dffff]progress[-][[#76bdff]%d[-]/[#76bdff]%d[-]]   errors[[yellow]%d[-]]   derivedJobs[[green]%d[-]]",
			s.TaskProgress.Completed,
			s.TaskProgress.Total,
			s.Errors.Completed,
			s.DerivedJobs)
	}
	return fmt.Sprintf(
		"  [#2dffff]progress[-][[#76bdff]%d[-]/[#76bdff]%d[-]]  [red]errors[-][[red]%d[-]]   derivedJobs[[green]%d[-]]",
		s.TaskProgress.Completed,
		s.TaskProgress.Total,
		s.Errors.Completed,
		s.DerivedJobs)
}

func getTagName(jobCtx interfaceJobCtx.IFaceJobCtx) string {
	if parId := jobCtx.GetParentId(); parId != 0 {
		return fmt.Sprintf("job#[#007dff]%d[-](derived by [#7d7dff]#%d[-])", jobCtx.GetJobId(), parId)
	}
	return fmt.Sprintf("job#[#007dff]%d[-]", jobCtx.GetJobId())
}

func (c *Ctx) updateTagName() string {
	c.tagName += "([#ff7d00]done[-])"
	return c.tagName
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

	screen.wgAdded.Add(1)
	screen.wg.Add(1)

	textViews, flx := newTextViewAndFlex(jobCtx.GetJobInfo())

	c := &Ctx{
		app:          screen.tviewApp,
		textViews:    textViews,
		flx:          flx,
		outputFormat: "native",
		verbosity:    outSetting.Verbosity,
		startCounter: make(chan struct{}),
		endCounter:   make(chan struct{}),
		jobCtx:       jobCtx,
		id:           id,
		tagName:      getTagName(jobCtx),
	}

	flx.SetInputCapture(func(key *tcell.EventKey) *tcell.EventKey {
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
		case "Rune[p]":
			c.textViews[IndCounter].SetTitle(titles[IndCounterPaused])
			c.jobCtx.Pause()
			return nil
		case "Rune[r]":
			c.textViews[IndCounter].SetTitle(titles[IndCounter])
			c.jobCtx.Resume()
			return nil
		case "Rune[q]":
			screen.pages.SwitchToPage("list-jobs")
			screen.tviewApp.SetFocus(screen.listJobs) // 回到菜单页面
			screen.removeListJobItemByName(c.tagName)
			c.jobCtx.Stop()
			c.jobCtx.Resume()
			screen.wg.Done()
			screen.wgAdded.Add(-1)
			return nil
		}
		return key
	})

	c.textViews[IndOutput].
		SetInputCapture(lockableInputCap(c, IndOutput, IndOutputLocked, &c.lockOnOutput))

	c.textViews[IndLogs].
		SetInputCapture(lockableInputCap(c, IndLogs, IndLogsLocked, &c.lockOnLog))

	flx.SetFocusFunc(func() { // 与下面的函数配合，当flx被聚焦时，将聚焦传递到上次选中的textView上
		c.app.SetFocus(c.textViews[c.focus])
	})

	c.app.QueueUpdate(func() { // tview作者真是神人，有的组件方法就加锁，有的就不加，我还得自己进去看再处理
		pageName := c.tagName
		screen.pages.AddPage(pageName, flx, false, false) // 添加一个名为job#id的页
		screen.addListJobItem(pageName, "", func() {
			// 选中对应list时，切换到这个页，并聚焦到flx上
			screen.pages.SwitchToPage(pageName)
			c.app.SetFocus(c.flx)
		})
	})

	go func() {
		<-c.startCounter
		for {
			select {
			case <-c.endCounter:
				s := c.counter.Snapshot()
				s.TaskRate = 0
				c.textViews[IndCounter].SetText(s.ToFmt())
				return
			default:
				c.textViews[IndCounter].SetText(c.counter.ToFmt())
				c.app.QueueUpdate(func() {
					ind := screen.getListItemIndexByName(c.tagName)
					if ind == -1 {
						return
					}
					screen.listJobs.SetItemText(ind, c.tagName, counterProgress(c.counter))
				})
				time.Sleep(225 * time.Millisecond)
			}
		}
	}()
	screen.activeTviewCtxs.Store(id, c)
	return c, nil
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
	screen.tviewApp.QueueUpdate(func() { // 在菜单页面将对应的项标记为完成
		old := c.tagName
		screen.updateItemName(old, c.updateTagName())
		ind := screen.getListItemIndexByName(c.tagName)
		if ind == -1 {
			return
		}
		screen.listJobs.SetItemText(ind, c.tagName, counterProgress(c.counter))
	})
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
		case "Rune[q]": // tview任务窗口目前设置为即使任务退出了也还是可以通过目录回来看，直到用户在任务页面按下q退出
			screen.removeListJobItemByName(c.tagName)
			screen.pages.SwitchToPage("list-jobs")
			screen.tviewApp.SetFocus(screen.listJobs) // 回到菜单页面
			screen.wg.Done()
			screen.wgAdded.Add(-1)
			return nil
		}
		return key
	})
	logJobDone := &outputable.Log{
		Msg:  "job is already done, press q to quit",
		Time: time.Now(),
	}
	c.Log(logJobDone)
	c.endCounter <- struct{}{}
	if c.counter == nil {
		c.startCounter <- struct{}{}
	}
	c.closed = true
	screen.activeTviewCtxs.Delete(c.id)
	c.jobCtx = nil
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
			screen.wg.Wait()
			screen.cancel()
			screen.tviewApp.Stop()
		}
	})
}

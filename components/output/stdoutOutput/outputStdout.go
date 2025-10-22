package stdoutOutput

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"time"
	"unsafe"
)

type counterInterior struct {
	XMLName      xml.Name         `json:"-" xml:"progress"`
	TaskRate     int64            `json:"task_rate,omitempty" xml:"task_rate,omitempty"`
	JobProgress  counter.Progress `json:"job_progress,omitempty" xml:"job_progress,omitempty"`
	TaskProgress counter.Progress `json:"task_progress,omitempty" xml:"task_progress,omitempty"`
}

func (c *counterInterior) ToFormatStr(format string) string {
	var (
		fmtBytes []byte
		err      error
	)
	switch format {
	case "xml":
		fmtBytes, err = xml.Marshal(c)
		if err != nil {
			return ""
		}
	case "json", "json-line":
		fmtBytes, err = json.Marshal(c)
		if err != nil {
			return ""
		}
	case "native":
		return fmt.Sprintf("REQ:[%d / %d]   JOB:[%d / %d]   RATE:[%d r/s]",
			c.TaskProgress.Completed, c.TaskProgress.Total, c.JobProgress.Completed,
			c.JobProgress.Total, c.TaskRate)
	default:
		return ""
	}
	return unsafe.String(&fmtBytes[0], len(fmtBytes)) // 省一点空间，毕竟计数器打印还是挺频繁的
}

// NewStdoutCtx 创建一个新的stdout fuzzCtx
func NewStdoutCtx(outSetting *fuzzTypes.OutputSetting, id int) (*Ctx, error) {
	if outSetting.ToWhere&outputFlag.OutToTview != 0 { // 不能同时启用tview输出和stdout
		return nil, errTviewConflict
	}

	c := &Ctx{
		id:        id,
		outputFmt: outSetting.OutputFormat,
		cntrReg:   make(chan struct{}),
		cntrStop:  make(chan struct{}),
		okToClose: make(chan struct{}),
	}

	fmt.Printf("stdout_%d_begin\n", c.id)
	ticker := time.NewTicker(400 * time.Millisecond)

	go func() {
		defer ticker.Stop()
		<-c.cntrReg
		for {
			select {
			case <-c.cntrStop:
				c.okToClose <- struct{}{}
				return
			case <-ticker.C:
				snapshot := c.counter.Snapshot()
				interior := counterInterior{
					TaskRate:     snapshot.TaskRate,
					JobProgress:  snapshot.JobProgress,
					TaskProgress: snapshot.TaskProgress,
				}
				fmt.Printf("[#%d COUNTER] %s\n", c.id, interior.ToFormatStr(c.outputFmt))
			}
		}
	}()
	return c, nil
}

// Output 将outObj输出到标准输出流上
func (c *Ctx) Output(obj *outputable.OutObj) error {
	if c.closed {
		return errCtxClosed
	}
	// 由于标准输出流是协程安全的，因此没必要加锁保护
	fmt.Printf("[#%d OUTPUT] %s\n", c.id, obj.ToFormatStr(c.outputFmt, false, 0))

	return nil
}

// Close 关闭标准输出上下文
func (c *Ctx) Close() error {
	if c.closed {
		return errCtxClosed
	}
	c.closed = true

	// 若没有调用RegisterCounter（也就是c.counter==nil），往cntrReg管道中发送，这样可以避免计数器协程卡在<-c.cntrReg
	if c.counter == nil {
		c.cntrReg <- struct{}{}
	}

	// 发送信号关闭
	c.cntrStop <- struct{}{}

	// 等计数器协程完全结束，避免在下面输出结束信息后还打印计数器
	<-c.okToClose

	fmt.Printf("stdout_%d_end\n", c.id)

	close(c.cntrReg)
	close(c.cntrStop)
	close(c.okToClose)

	return nil
}

// Log 向stdout输出一条日志
func (c *Ctx) Log(log *outputable.Log) error {
	if c.closed {
		return errCtxClosed
	}
	fmt.Printf("[#%d LOG] %s\n", c.id, log.ToFormatStr(c.outputFmt))
	return nil
}

// RegisterCounter 将一个counter注册到当前上下文中
func (c *Ctx) RegisterCounter(cntr *counter.Counter) error {
	if cntr == nil {
		return errRegisterNil
	} else if c.closed {
		return errCtxClosed
	}
	c.counter = cntr
	c.cntrReg <- struct{}{}
	return nil
}

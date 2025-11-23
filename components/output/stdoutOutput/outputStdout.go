package stdoutOutput

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/output/outputErrors"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"time"
)

// NewStdoutCtx 创建一个新的stdout fuzzCtx
func NewStdoutCtx(outSetting *fuzzTypes.OutputSetting, id int) (*Ctx, error) {
	if outSetting.ToWhere&outputFlag.OutToTview != 0 { // 不能同时启用tview输出和stdout
		return nil, outputErrors.ErrTviewConflict
	}

	c := &Ctx{
		id:              id,
		outputFmt:       outSetting.OutputFormat,
		outputVerbosity: outSetting.Verbosity,
		cntrReg:         make(chan struct{}),
		cntrStop:        make(chan struct{}),
		okToClose:       make(chan struct{}),
	}

	fmt.Printf("stdout_%d_begin\n", c.id)
	ticker := time.NewTicker(500 * time.Millisecond)

	go func() {
		defer ticker.Stop()
		<-c.cntrReg
		for {
			select {
			case <-c.cntrStop:
				c.okToClose <- struct{}{}
				return
			case <-ticker.C:
				fmt.Printf("[#%d COUNTER] %s\n", c.id, c.counter.ToFmt())
			}
		}
	}()
	return c, nil
}

// Output 将outObj输出到标准输出流上
func (c *Ctx) Output(obj *outputable.OutObj) error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	// 由于标准输出流是协程安全的，因此没必要加锁保护
	fmt.Printf("[#%d OUTPUT] %s\n", c.id, obj.ToFormatStr(c.outputFmt, false, c.outputVerbosity))

	return nil
}

// Close 关闭标准输出上下文
func (c *Ctx) Close() error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	c.closed = true

	// 发送信号关闭
	c.cntrStop <- struct{}{}

	// 若没有调用RegisterCounter（也就是c.counter==nil），往cntrReg管道中发送，这样可以避免计数器协程卡在<-c.cntrReg
	if c.counter == nil {
		c.cntrReg <- struct{}{}
	}

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
		return outputErrors.ErrCtxClosed
	}
	fmt.Printf("[#%d LOG] %s\n", c.id, log.ToFormatStr(c.outputFmt))
	return nil
}

// RegisterCounter 将一个counter注册到当前上下文中
func (c *Ctx) RegisterCounter(cntr *counter.Counter) error {
	if cntr == nil {
		return outputErrors.ErrRegisterNilCounter
	} else if c.closed {
		return outputErrors.ErrCtxClosed
	}
	c.counter = cntr
	c.cntrReg <- struct{}{}
	return nil
}

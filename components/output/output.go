package output

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	co "github.com/nostalgist134/FuzzGIU/components/output/chanOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	fo "github.com/nostalgist134/FuzzGIU/components/output/fileOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/httpOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/interfaceJobCtx"
	"github.com/nostalgist134/FuzzGIU/components/output/outCtx"
	"github.com/nostalgist134/FuzzGIU/components/output/outputErrors"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	so "github.com/nostalgist134/FuzzGIU/components/output/stdoutOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/tviewOutput"
	"sync"
	"time"
)

/*
output 包用于处理FuzzGIU的输出结果，主要提供3类函数：InitOutput、Output和Finish。
	NewOutputCtx函数根据输出设置生成一个输出上下文
	Output函数用于根据输出上下文向指定输出流输出结果
	Close函数用于关闭一个输出上下文

	仅有Output之间、Output与Log间、Log之间、Output或Log与Close的并发调用是保证协程安全的（通过子Ctx中的协程安全方法与waitGroup保证），
	其它函数间的调用均不协程安全，需自行确定。代码会假设函数的调用遵循NewOutputCtx->Output或者Log（可以并发）->Close的过程
*/

type (
	OutObj = outputable.OutObj
	Log    = outputable.Log
	Ctx    outCtx.OutputCtx // outCtx.OutputCtx仅定义类成员，类方法在此文件中定义
)

var (
	outObjPool                = sync.Pool{New: func() any { return new(OutObj) }}
	errDbOutputNotImplemented = errors.New("db output not implemented yet")
)

func GetOutputObj() *OutObj {
	return outObjPool.Get().(*OutObj)
}

func PutOutputObj(obj *OutObj) {
	*obj = OutObj{}
	outObjPool.Put(obj)
}

// NewOutputCtx 根据输出设置，生成一个输出上下文，从而在之后可以使用此上下文进行输出
func NewOutputCtx(outSetting *fuzzTypes.OutputSetting, jobCtx interfaceJobCtx.IFaceJobCtx, jid int) (*Ctx, error) {
	if outSetting == nil {
		return nil, outputErrors.ErrNilOutputSetting
	}
	if jobCtx == nil {
		return nil, outputErrors.ErrNilJobCtx
	}

	toWhere := outSetting.ToWhere
	format := outSetting.OutputFormat

	if !outputable.FormatSupported(format) {
		return nil, fmt.Errorf("unsupported output format '%s'", format)
	}

	outputCtx := new(Ctx)

	outputCtx.Counter = new(counter.Counter)

	var err error

	if toWhere&outputFlag.OutToFile != 0 {
		fc, err1 := fo.NewFileOutputCtx(outSetting, jid)
		err = err1
		if err1 == nil {
			outputCtx.FileOutputCtx = fc
			outputCtx.ToWhere |= outputFlag.OutToFile
		}
	}

	if toWhere&outputFlag.OutToChan != 0 {
		cc, err1 := co.NewOutputChanCtx(outSetting, jid)
		err = errors.Join(err, err1)
		if err1 == nil {
			outputCtx.ChanOutputCtx = cc
			outputCtx.ToWhere |= outputFlag.OutToChan
		}
	}

	if toWhere&outputFlag.OutToStdout != 0 {
		sc, err1 := so.NewStdoutCtx(outSetting, jid)
		err = errors.Join(err, err1)
		if err1 == nil {
			outputCtx.StdoutCtx = sc
			outputCtx.ToWhere |= outputFlag.OutToStdout
		}

		if sc != nil { // Stdout对计数器的输出需要手动注册
			err2 := sc.RegisterCounter(outputCtx.Counter)
			err = errors.Join(err, err2)
		}
	}

	if toWhere&outputFlag.OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
	}

	if toWhere&outputFlag.OutToHttp != 0 {
		httpCtx, err1 := httpOutput.NewHttpOutputCtx(outSetting, jid)
		err = errors.Join(err, err1)
		if err1 == nil {
			outputCtx.HttpCtx = httpCtx
			outputCtx.ToWhere |= outputFlag.OutToHttp
		}
	}

	if toWhere&outputFlag.OutToTview != 0 {
		tviewCtx, err1 := tviewOutput.NewTviewOutputCtx(outSetting, jobCtx, jid)
		err = errors.Join(err, err1)
		if err1 == nil {
			outputCtx.TviewOutputCtx = tviewCtx
			outputCtx.ToWhere |= outputFlag.OutToTview
		}

		if tviewCtx != nil {
			err2 := tviewCtx.RegisterCounter(outputCtx.Counter)
			err = errors.Join(err, err2)
		}
	}

	return outputCtx, nil
}

// Output 向输出上下文中输出数据
func (c *Ctx) Output(obj *OutObj) error {
	c.Wg.Add(1)
	defer c.Wg.Done()
	toWhere := c.ToWhere

	c.Counter.Complete(counter.CntrOut)

	var err error

	if toWhere&outputFlag.OutToFile != 0 {
		err = errors.Join(err, c.FileOutputCtx.Output(obj))
	}

	if toWhere&outputFlag.OutToChan != 0 {
		err = errors.Join(err, c.ChanOutputCtx.Output(obj))
	}

	if toWhere&outputFlag.OutToStdout != 0 {
		err = errors.Join(err, c.StdoutCtx.Output(obj))
	}

	if toWhere&outputFlag.OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
	}

	if toWhere&outputFlag.OutToHttp != 0 {
		err = errors.Join(err, c.HttpCtx.Output(obj))
	}

	if toWhere&outputFlag.OutToTview != 0 {
		err = errors.Join(err, c.TviewOutputCtx.Output(obj))
	}

	return err
}

// Close 关闭一个输出上下文（此函数会阻塞直到所有输出、日志都完成才真正关闭）
func (c *Ctx) Close() error {
	c.Wg.Wait()
	toWhere := c.ToWhere
	var err error

	if toWhere&outputFlag.OutToFile != 0 {
		err = errors.Join(err, c.FileOutputCtx.Close())
	}

	if toWhere&outputFlag.OutToChan != 0 {
		err = errors.Join(err, c.ChanOutputCtx.Close())
	}

	if toWhere&outputFlag.OutToStdout != 0 {
		err = errors.Join(err, c.StdoutCtx.Close())
	}

	if toWhere&outputFlag.OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
	}

	if toWhere&outputFlag.OutToHttp != 0 {
		err = errors.Join(err, c.HttpCtx.Close())
	}

	if toWhere&outputFlag.OutToTview != 0 {
		err = errors.Join(err, c.TviewOutputCtx.Close())
	}

	c.Counter.StopRecordTaskRate()

	return err
}

// Log 输出一条日志
func (c *Ctx) Log(log *outputable.Log) error {
	c.Wg.Add(1)
	defer c.Wg.Done()

	toWhere := c.ToWhere
	var err error

	if toWhere&outputFlag.OutToFile != 0 {
		err = errors.Join(err, c.FileOutputCtx.Log(log))
	}

	if toWhere&outputFlag.OutToChan != 0 {
		err = errors.Join(err, c.ChanOutputCtx.Log(log))
	}

	if toWhere&outputFlag.OutToStdout != 0 {
		err = errors.Join(err, c.StdoutCtx.Log(log))
	}

	if toWhere&outputFlag.OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
	}

	if toWhere&outputFlag.OutToHttp != 0 {
		err = errors.Join(err, c.HttpCtx.Log(log))
	}

	if toWhere&outputFlag.OutToTview != 0 {
		err = errors.Join(err, c.TviewOutputCtx.Log(log))
	}

	return err
}

// LogFmtMsg 格式化输出日志
func (c *Ctx) LogFmtMsg(format string, a ...any) error {
	l := outputable.Log{
		Jid:  c.Id,
		Msg:  fmt.Sprintf(format, a...),
		Time: time.Now(),
	}
	return c.Log(&l)
}

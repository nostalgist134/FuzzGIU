package output

import (
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	co "github.com/nostalgist134/FuzzGIU/components/output/chanOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	fo "github.com/nostalgist134/FuzzGIU/components/output/fileOutput"
	"github.com/nostalgist134/FuzzGIU/components/output/outCtx"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	so "github.com/nostalgist134/FuzzGIU/components/output/stdoutOutput"
	"time"
)

/*
output 包用于处理FuzzGIU的输出结果，主要提供3类函数：InitOutput、Output和Finish。
	NewOutputCtx函数根据输出设置生成一个输出上下文
	Output函数用于根据输出上下文向指定输出流输出结果
	Close函数用于关闭一个输出上下文

	注意：仅有Output之间、Output与Log间、Log间的并发调用是保证协程安全的（通过子Ctx中的协程安全方法保证），Output与其它函数间的调用、其它函数
	间的调用均不协程安全，需自行确定。代码会假设函数的调用遵循NewOutputCtx->Output或者Log（可以并发）->Close的过程
*/

type (
	OutObj  = outputable.OutObj
	Log     = outputable.Log
	Counter = counter.Counter
	Ctx     outCtx.OutputCtx // outCtx.OutputCtx仅定义类成员，类方法在此文件中定义
)

var errDbOutputNotImplemented = errors.New("db output not implemented yet")

// NewOutputCtx 根据输出设置，生成一个输出上下文，从而在之后可以使用此上下文进行输出
func NewOutputCtx(outSetting *fuzzTypes.OutputSetting, jid int) (*Ctx, error) {
	toWhere := outSetting.ToWhere
	format := outSetting.OutputFormat

	if !outputable.FormatSupported(format) {
		return nil, fmt.Errorf("unsupported output format '%s'", format)
	}

	oc := new(Ctx)

	oc.Counter = &Counter{
		StartTime: time.Now(),
	}

	var err error

	if toWhere&OutToFile != 0 {
		fc, err1 := fo.NewFileOutputCtx(outSetting, jid)
		err = errors.Join(err, err1)

		oc.FileOutputCtx = fc
	}

	if toWhere&OutToChan != 0 {
		cc, err1 := co.NewOutputChanCtx(outSetting, jid)
		err = errors.Join(err, err1)

		oc.ChanOutputCtx = cc
	}

	if toWhere&OutToStdout != 0 {
		sc, err1 := so.NewStdoutCtx(outSetting, jid)
		err = errors.Join(err, err1)
		oc.StdoutCtx = sc

		if sc != nil { // Stdout对计数器的输出需要手动注册
			err2 := sc.RegisterCounter(oc.Counter)
			err = errors.Join(err, err2)
		}
	}

	if toWhere&OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
	}

	if toWhere&OutToTview != 0 { // todo
	}

	return oc, nil
}

// Output 向输出上下文中输出数据
func (c *Ctx) Output(obj *OutObj) error {
	toWhere := c.OutSetting.ToWhere

	var err error

	if toWhere&OutToFile != 0 {
		err = errors.Join(err, c.FileOutputCtx.Output(obj))
	}

	if toWhere&OutToChan != 0 {
		err = errors.Join(err, c.ChanOutputCtx.Output(obj))
	}

	if toWhere&OutToStdout != 0 {
		err = errors.Join(err, c.StdoutCtx.Output(obj))
	}

	if toWhere&OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
	}

	return err
}

// Close 关闭一个输出上下文
func (c *Ctx) Close() error {
	toWhere := c.OutSetting.ToWhere
	var err error

	if toWhere&OutToFile != 0 {
		err = errors.Join(err, c.FileOutputCtx.Close())
	}

	if toWhere&OutToChan != 0 {
		err = errors.Join(err, c.ChanOutputCtx.Close())
	}

	if toWhere&OutToStdout != 0 {
		err = errors.Join(err, c.StdoutCtx.Close())
	}

	if toWhere&OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
	}

	return err
}

// Log 输出一条日志
func (c *Ctx) Log(log *outputable.Log) error {
	toWhere := c.OutSetting.ToWhere
	var err error

	if toWhere&OutToFile != 0 {
		err = errors.Join(err, c.FileOutputCtx.Log(log))
	}

	if toWhere&OutToChan != 0 {
		err = errors.Join(err, c.ChanOutputCtx.Log(log))
	}

	if toWhere&OutToStdout != 0 {
		err = errors.Join(err, c.StdoutCtx.Log(log))
	}

	if toWhere&OutToDB != 0 {
		err = errors.Join(err, errDbOutputNotImplemented)
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

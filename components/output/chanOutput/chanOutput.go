package chanOutput

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
)

type Ctx struct {
	outChan chan *outputable.OutObj
	closed  bool
}

// NewOutputChanCtx 创建一个新的输出管道
func NewOutputChanCtx(outSetting *fuzzTypes.OutputSetting, _ int) (*Ctx, error) {
	chanSize := outSetting.ChanSize
	c := &Ctx{make(chan *outputable.OutObj, chanSize), false}
	return c, nil
}

// Output 向管道推送一个OutObj
func (c *Ctx) Output(obj *outputable.OutObj) error {
	if c.closed {
		return errChanClosed
	}
	select {
	case c.outChan <- obj:
		return nil
	default:
		return errChanFull
	}
}

// Close 关闭管道输出上下文，注意：不保证协程安全，需调用方自行判断，new以后若要关闭，需要确定已经没有协程在调用Output，然后才能调用
func (c *Ctx) Close() error {
	if c.closed {
		return errChanClosed
	}
	c.closed = true
	close(c.outChan)
	return nil
}

func (c *Ctx) Log(_ *outputable.Log) error {
	return errLogNotImplemented
}

// OutputChan 获取当前ctx的输出管道（只读）
func (c *Ctx) OutputChan() <-chan *outputable.OutObj {
	return c.outChan
}

// GetSingleOutObj 从管道中取出一个输出对象
func (c *Ctx) GetSingleOutObj() (*outputable.OutObj, bool) {
	select {
	case obj, ok := <-c.outChan:
		return obj, ok
	default:
		return nil, false
	}
}

// GetChanCap 获取管道的最大大小
func (c *Ctx) GetChanCap() int {
	return cap(c.outChan)
}

// GetChanCurUsed 获取管道当前占用情况
func (c *Ctx) GetChanCurUsed() int {
	return len(c.outChan)
}

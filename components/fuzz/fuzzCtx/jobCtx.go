package fuzzCtx

import (
	"context"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"sync"
)

// JobCtx 单个fuzz任务上下文 inspired by eprocess win kernel
type JobCtx struct {
	JobId     int
	ParentId  int
	Job       *fuzzTypes.Fuzz
	RP        IFaceRP // 改为使用interface，这样就避免引用rp包导致引用循环
	OutputCtx *output.Ctx
	GlobCtx   context.Context
	Cancel    context.CancelFunc
	occupied  sync.WaitGroup
	closeOnce sync.Once
}

// Close 关闭任务上下文，释放其资源
func (jc *JobCtx) Close() error {
	jc.occupied.Wait()
	jc.RP.ReleaseSelf()
	jc.Stop()
	return jc.OutputCtx.Close()
}

// Stop 停止当前任务
func (jc *JobCtx) Stop() {
	if jc.Cancel != nil {
		jc.Cancel()
	}
	jc.RP.Resume() // 让rp继续运行，防止卡在暂停状态接收不到停止信息
}

// Pause 暂停当前任务
func (jc *JobCtx) Pause() {
	if jc.RP != nil {
		jc.RP.Pause()
	}
}

// Resume 继续执行当前任务
func (jc *JobCtx) Resume() {
	if jc.RP != nil {
		jc.RP.Resume()
	}
}

// Occupy 将jobCtx标记为占用，防止关闭
func (jc *JobCtx) Occupy() {
	jc.occupied.Add(1)
}

// Release 将jobCtx占用数减一
func (jc *JobCtx) Release() {
	jc.occupied.Done()
}

// GetJobInfo 获取当前任务的任务信息
func (jc *JobCtx) GetJobInfo() *fuzzTypes.Fuzz {
	return jc.Job
}

package rp

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
	"time"
)

const (
	StatStop    = 0
	StatRunning = 1
	StatPause   = 2

	ExecMajor = int8(0)
	ExecMinor = int8(1)
)

type tsk struct {
	arg       *fuzzCtx.TaskCtx
	whichExec int8
}

type RoutinePool struct {
	tasks       chan tsk
	results     chan *fuzzTypes.Reaction
	resizeMu    sync.Mutex
	concurrency int

	wg        sync.WaitGroup
	quit      chan struct{}
	status    int8
	statMu    sync.Mutex
	cond      *sync.Cond
	executors [2]func(*fuzzCtx.TaskCtx) *fuzzTypes.Reaction
}

// nopExecutor 空执行器，作为默认值避免空指针
func nopExecutor(*fuzzCtx.TaskCtx) *fuzzTypes.Reaction {
	return nil
}

// newRoutinePool 创建一个新的协程池
func newRoutinePool(concurrency int) *RoutinePool {
	routinePool := &RoutinePool{
		concurrency: concurrency,
		executors:   [2]func(*fuzzCtx.TaskCtx) *fuzzTypes.Reaction{nopExecutor, nopExecutor},
	}
	routinePool.cond = sync.NewCond(&routinePool.statMu)
	return routinePool
}

// RegisterExecutor 注册任务执行函数，如果协程池是运行状态，或注册下标不为ExecMinor或ExecMajor，则退出
// 特别注意：协程池在运行状态时是不能调用此函数的，必须先暂停
func (p *RoutinePool) RegisterExecutor(executor func(*fuzzCtx.TaskCtx) *fuzzTypes.Reaction, which int8) {
	if which != ExecMinor && which != ExecMajor {
		return
	}
	p.statMu.Lock()
	defer p.statMu.Unlock()
	if p.status == StatRunning {
		return
	}
	p.executors[which] = executor
}

// Start 启动所有 worker
func (p *RoutinePool) Start() {
	p.statMu.Lock()
	defer p.statMu.Unlock()
	if p.status == StatRunning {
		return
	}
	if p.status == StatStop {
		p.tasks = make(chan tsk, 8192)
		p.results = make(chan *fuzzTypes.Reaction, 8192)
		p.quit = make(chan struct{})
	}
	p.status = StatRunning
	p.resizeMu.Lock()
	defer p.resizeMu.Unlock()
	for i := 0; i < p.concurrency; i++ {
		go p.worker()
	}
}

func (p *RoutinePool) worker() {
	for {
		// 先检查退出信号
		select {
		case <-p.quit:
			return
		default:
		}
		// 检查状态并决定是否等待
		p.statMu.Lock()
		switch p.status {
		case StatStop:
			p.statMu.Unlock()
			return
		case StatPause:
			// 处于暂停状态，等待唤醒
			p.cond.Wait()
			p.statMu.Unlock()
		case StatRunning:
			// 处于运行状态，释放锁并尝试获取任务
			p.statMu.Unlock()
			select {
			case task, ok := <-p.tasks:
				if !ok { // 管道关闭且没有值了，退出
					return
				}
				result := p.executors[task.whichExec](task.arg)
				p.results <- result
				p.wg.Done()
			case <-p.quit:
				return
			}
		}
	}
}

// Submit 添加任务
func (p *RoutinePool) Submit(execArg *fuzzCtx.TaskCtx, whichExec int8, timeout time.Duration) bool {
	p.statMu.Lock()
	if p.status != StatRunning {
		p.statMu.Unlock()
		return false
	}
	p.statMu.Unlock()

	p.wg.Add(1)

	task := tsk{
		arg:       execArg,
		whichExec: whichExec,
	}

	if timeout < 0 {
		p.tasks <- task
		return true
	}

	select {
	case p.tasks <- task:
		return true
	case <-time.After(timeout):
		p.wg.Done()
		return false
	}
}

// Wait 等待协程池若干时间
func (p *RoutinePool) Wait(maxTime time.Duration) bool {
	if maxTime < 0 {
		p.wg.Wait()
		return true
	}
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return true
	case <-time.After(maxTime):
		return false
	}
}

// Stop 关闭管道，停止所有 worker
func (p *RoutinePool) Stop() {
	p.statMu.Lock()
	if p.status == StatStop {
		p.statMu.Unlock()
		return
	}
	p.statMu.Unlock()

	p.Clear()

	p.statMu.Lock()
	defer p.statMu.Unlock()
	close(p.quit)
	close(p.tasks)
	close(p.results)
	p.status = StatStop
	p.cond.Broadcast()
}

// Pause 暂停调度
func (p *RoutinePool) Pause() {
	p.statMu.Lock()
	defer p.statMu.Unlock()
	if p.status == StatRunning {
		p.status = StatPause
		p.cond.Broadcast()
	}
}

// Resume 恢复调度
func (p *RoutinePool) Resume() {
	p.statMu.Lock()
	defer p.statMu.Unlock()
	if p.status == StatPause {
		p.status = StatRunning
		p.cond.Broadcast()
	}
}

func (p *RoutinePool) Resize(size int) {
	if p.Status() == StatStop {
		return
	}
	p.resizeMu.Lock()
	defer p.resizeMu.Unlock()
	if size == p.concurrency || size < 0 {
		return
	}
	if size > p.concurrency {
		for i := 0; i < size-p.concurrency; i++ {
			go p.worker()
		}
	} else {
		for i := 0; i < p.concurrency-size; i++ {
			p.quit <- struct{}{}
		}
	}
	p.concurrency = size
}

// GetSingleResult 获取单个任务结果
func (p *RoutinePool) GetSingleResult() *fuzzTypes.Reaction {
	select {
	case r := <-p.results:
		return r
	default:
		return nil
	}
}

func (p *RoutinePool) Status() int8 {
	p.statMu.Lock()
	defer p.statMu.Unlock()
	s := p.status
	return s
}

func (p *RoutinePool) WaitResume() {
	p.statMu.Lock()
	defer p.statMu.Unlock()

	for p.status != StatRunning {
		if p.status == StatStop {
			return // 已停止，直接返回
		}
		p.cond.Wait()
	}
}

// Clear 清空任务队列
func (p *RoutinePool) Clear() {
	p.statMu.Lock()
	if p.status == StatStop {
		p.statMu.Unlock()
		return
	}
	p.statMu.Unlock()

	p.Pause()
	defer p.Resume()
	for {
		select {
		case <-p.tasks:
			p.wg.Done()
		case <-p.results:
		default:
			return
		}
	}
}

// ReleaseSelf 将自身放入池中
func (p *RoutinePool) ReleaseSelf() {
	putRp(p)
}

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

// todo: 这个包循环import了，具体的是rp import fuzzCtx, fuzzCtx.JobCtx.RP -> import rp
// 可能的解决方案：1.在这个包里再声明一次taskCtx，然后之后调用都强转
// 	2.tsk.arg与executor的参数类型全改为使用unsafe.Pointer，不过这样需要改的内容会很多，而且需要另起一个池来管理execCtx
//	已解决，在fuzzCtx中声明了一个rp接口，然后使用的时候不直接用rp类型而是用接口，这样就避免了

type tsk struct {
	arg       *fuzzCtx.TaskCtx
	whichExec int8
}

type RoutinePool struct {
	tasks       chan tsk
	results     chan *fuzzTypes.Reaction
	concurrency int

	wg        sync.WaitGroup
	quit      chan struct{}
	status    int8
	mu        sync.Mutex
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
	routinePool.cond = sync.NewCond(&routinePool.mu)
	return routinePool
}

// RegisterExecutor 注册任务执行函数，如果协程池是运行状态，或注册下标不为ExecMinor或ExecMajor，则退出
// 特别注意：协程池在运行状态时是不能调用此函数的，必须先暂停
func (p *RoutinePool) RegisterExecutor(executor func(*fuzzCtx.TaskCtx) *fuzzTypes.Reaction, which int8) {
	if which != ExecMinor && which != ExecMajor {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == StatRunning {
		return
	}
	p.executors[which] = executor
}

// Start 启动所有 worker
func (p *RoutinePool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == StatRunning {
		return
	}
	if p.status == StatStop {
		p.tasks = make(chan tsk, 8192)
		p.results = make(chan *fuzzTypes.Reaction, 8192)
		p.quit = make(chan struct{})
	}
	p.status = StatRunning
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
		p.mu.Lock()
		switch p.status {
		case StatStop:
			p.mu.Unlock()
			return
		case StatPause:
			// 处于暂停状态，等待唤醒
			p.cond.Wait()
			p.mu.Unlock()
		case StatRunning:
			// 处于运行状态，释放锁并尝试获取任务
			p.mu.Unlock()

			// 阻塞等待任务或退出信号，避免忙循环
			select {
			case task, ok := <-p.tasks:
				if !ok {
					continue
				}
				result := p.executors[task.whichExec](task.arg)
				p.results <- result
				p.wg.Done()
			case <-p.quit:
				return
			case <-time.After(5 * time.Millisecond): // 短暂超时
			}
		}
	}
}

// Submit 添加任务
func (p *RoutinePool) Submit(execArg *fuzzCtx.TaskCtx, whichExec int8, timeout time.Duration) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.status != StatRunning {
		return false
	}

	p.wg.Add(1)

	// 提交函数、执行可以用指针，这样可以减少栈分配，但是tsk结构必须是字面值复制，不然又要写一个资源池来管理
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
	p.mu.Lock()
	if p.status == StatStop {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

	p.Clear()

	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.quit)
	close(p.tasks)
	close(p.results)
	p.status = StatStop
	p.cond.Broadcast()
}

// Pause 暂停调度
func (p *RoutinePool) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == StatRunning {
		p.status = StatPause
		p.cond.Broadcast()
	}
}

// Resume 恢复调度
func (p *RoutinePool) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == StatPause {
		p.status = StatRunning
		p.cond.Broadcast()
	}
}

func (p *RoutinePool) Resize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if size == p.concurrency || size < 0 || p.status == StatStop {
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
	p.mu.Lock()
	defer p.mu.Unlock()
	s := p.status
	return s
}

func (p *RoutinePool) WaitResume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.status != StatRunning {
		if p.status == StatStop {
			return // 已停止，直接返回
		}
		p.cond.Wait()
	}
}

// Clear 清空任务队列
func (p *RoutinePool) Clear() {
	p.mu.Lock()
	if p.status == StatStop {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()

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

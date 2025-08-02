package rp

import (
	"sync"
	"time"

	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
)

const (
	StatStop    = 0
	StatRunning = 1
	StatPause   = 2
)

type Task func() *fuzzTypes.Reaction

type RoutinePool struct {
	tasks       chan Task
	results     chan *fuzzTypes.Reaction
	concurrency int

	wg     sync.WaitGroup
	quit   chan struct{}
	status int8
	mu     sync.Mutex
	cond   *sync.Cond
}

var CurrentRp *RoutinePool

// New 创建一个新的协程池
func New(concurrency int) *RoutinePool {
	wp := &RoutinePool{
		concurrency: concurrency,
	}
	wp.cond = sync.NewCond(&wp.mu)
	CurrentRp = wp
	return CurrentRp
}

// Start 启动所有 worker
func (p *RoutinePool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == StatRunning {
		return
	}
	if p.status == StatStop {
		p.tasks = make(chan Task, 8192)
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
				result := task()
				p.results <- result
				p.wg.Done()
			case <-p.quit:
				return
			case <-time.After(10 * time.Millisecond):
				// 短暂超时，让worker有机会检查状态变化
			}
		}
	}
}

func (p *RoutinePool) waitForResume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.status == StatPause {
		select {
		case <-p.quit:
			return
		default:
			p.cond.Wait()
		}
	}
}

// Submit 添加任务
func (p *RoutinePool) Submit(task Task, timeout time.Duration) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status != StatRunning {
		return false
	}
	p.wg.Add(1)
	if timeout < 0 {
		p.tasks <- task
		return true
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case p.tasks <- task:
		return true
	case <-timer.C:
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

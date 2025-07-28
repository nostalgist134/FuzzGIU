package wp

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
	"time"
)

/*
状态转移规则:
1.statStop->statRunning 可行，但是管道要重建
2.statStop->statPause 不可行
3.statRunning->statStop 可行，关闭管道
4.statRunning->statPause 可行，向pause管道发信息
5.statPause->statRunning 可行，向resume管道发信息
6.statPause->statStop 可行，因为waitforresume会根据quit管道进行退出
*/

const (
	statStop    = 0
	statRunning = 1
	statPause   = 2
)

type Task func() *fuzzTypes.Reaction

type WorkerPool struct {
	tasks       chan Task
	results     chan *fuzzTypes.Reaction
	concurrency int

	wg     sync.WaitGroup
	quit   chan struct{}
	pause  chan struct{}
	resume chan struct{}
	status int8
	mu     sync.Mutex
}

var CurrentWp *WorkerPool

// New 创建一个新的协程池
func New(concurrency int) *WorkerPool {
	CurrentWp = &WorkerPool{
		concurrency: concurrency,
		pause:       make(chan struct{}),
		resume:      make(chan struct{}),
	}
	return CurrentWp
}

// Start 启动所有 worker
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == statRunning {
		return
	}
	// 从停止状态恢复时，重新创建管道
	if p.status == statStop {
		p.tasks = make(chan Task, 8192)
		p.results = make(chan *fuzzTypes.Reaction, 8192)
		p.quit = make(chan struct{})
	}
	p.status = statRunning
	for i := 0; i < p.concurrency; i++ {
		go p.worker()
	}
}

func (p *WorkerPool) worker() {
	for {
		select {
		case <-p.quit:
			return
		case <-p.pause:
			p.waitForResume()
		case task, ok := <-p.tasks:
			if !ok {
				continue
			}
			result := task()
			p.results <- result
			p.wg.Done()
		}
	}
}

func (p *WorkerPool) waitForResume() {
	for {
		select {
		case <-p.resume:
			return
		case <-p.quit:
			return
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Submit 添加任务
func (p *WorkerPool) Submit(task Task, timeout time.Duration) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status != statRunning {
		return false
	}
	p.wg.Add(1)
	if timeout < 0 {
		// 无限等待
		p.tasks <- task
		return true
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case p.tasks <- task:
		return true
	case <-timer.C:
		p.wg.Done() // 撤销加进去的任务计数
		return false
	}
}

// Wait 等待协程池若干时间（maxTime设为负值则不限时间），如果等待完成返回true，否则返回false
func (p *WorkerPool) Wait(maxTime time.Duration) bool {
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
func (p *WorkerPool) Stop() {
	// 如果已经关闭则退出
	p.mu.Lock()
	if p.status == statStop {
		p.mu.Unlock()
		return
	}
	p.mu.Unlock()
	// 清空管道中的数据
	p.Clear()
	// 关闭管道
	p.mu.Lock()
	defer p.mu.Unlock()
	close(p.quit)
	close(p.tasks)
	close(p.results)
	p.status = statStop
}

// Pause 暂停调度
func (p *WorkerPool) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == statRunning {
		p.status = statPause
		for i := 0; i < p.concurrency; i++ {
			p.pause <- struct{}{}
		}
	}
}

// Resume 恢复调度
func (p *WorkerPool) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == statPause {
		p.status = statRunning
		for i := 0; i < p.concurrency; i++ {
			p.resume <- struct{}{}
		}
	}
}

func (p *WorkerPool) Resize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if size == p.concurrency || size < 0 || p.status == statStop {
		return
	} else {
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
}

// GetSingleResult 获取单个任务结果
func (p *WorkerPool) GetSingleResult() *fuzzTypes.Reaction {
	select {
	case r := <-p.results:
		return r
	default:
		return nil
	}
}

// Clear 清空任务队列
func (p *WorkerPool) Clear() {
	p.mu.Lock()
	if p.status == statStop {
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

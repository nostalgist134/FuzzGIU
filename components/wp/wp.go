package wp

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
	"time"
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
	paused bool
	mu     sync.Mutex
}

var Wp *WorkerPool

// New 创建一个新的协程池
func New(concurrency int) *WorkerPool {
	Wp = &WorkerPool{
		tasks:       make(chan Task, 8192),
		results:     make(chan *fuzzTypes.Reaction, 8192),
		concurrency: concurrency,
		quit:        make(chan struct{}),
		pause:       make(chan struct{}),
		resume:      make(chan struct{}),
	}
	return Wp
}

// Start 启动所有 worker
func (p *WorkerPool) Start() {
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
		case task := <-p.tasks:
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
func (p *WorkerPool) Submit(task Task) {
	p.wg.Add(1)
	p.tasks <- task
}

// Wait 等待协程池若干时间（maxTime设为负值则不限时间），如果等待完成返回true，否则返回false
func (p *WorkerPool) Wait(maxTime time.Duration) bool {
	if maxTime < 0 {
		p.wg.Wait()
		return true
	}
	waitDone := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
		return true
	case <-time.After(maxTime):
		return false
	}
}

// Stop 停止所有 worker
func (p *WorkerPool) Stop() {
	close(p.quit)
	close(p.results) // 关闭结果通道，避免 GetResChan 中泄露
}

// Pause 暂停调度
func (p *WorkerPool) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.paused {
		p.paused = true
		for i := 0; i < p.concurrency; i++ {
			p.pause <- struct{}{}
		}
	}
}

// Resume 恢复调度
func (p *WorkerPool) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.paused {
		p.paused = false
		for i := 0; i < p.concurrency; i++ {
			p.resume <- struct{}{}
		}
	}
}

func (p *WorkerPool) Resize(size int) {
	if size == p.concurrency || size < 0 {
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
	p.Pause()
	defer p.Resume()
	for {
		select {
		case <-p.tasks:
			p.wg.Done()
		default:
			return
		}
	}
}

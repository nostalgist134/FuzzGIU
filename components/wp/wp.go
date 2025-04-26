package wp

import (
	"FuzzGIU/components/fuzzTypes"
	"sync"
	"time"
)

type TaskFunc func() *fuzzTypes.Reaction

type WorkerPool struct {
	tasks       chan TaskFunc
	results     chan *fuzzTypes.Reaction
	concurrency int

	wg         sync.WaitGroup
	stopChan   chan struct{}
	pauseChan  chan struct{}
	resumeChan chan struct{}
	paused     bool
	mu         sync.Mutex
}

// NewWorkerPool 创建一个新的协程池
func NewWorkerPool(concurrency int) *WorkerPool {
	return &WorkerPool{
		tasks:       make(chan TaskFunc, 4096),
		results:     make(chan *fuzzTypes.Reaction, 4096),
		concurrency: concurrency,
		stopChan:    make(chan struct{}),
		pauseChan:   make(chan struct{}),
		resumeChan:  make(chan struct{}),
	}
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
		case <-p.stopChan:
			return
		case <-p.pauseChan:
			p.waitForResume()
		case task := <-p.tasks:
			if task != nil {
				result := task()
				p.results <- result
				p.wg.Done()
			}
		}
	}
}

func (p *WorkerPool) waitForResume() {
	for {
		select {
		case <-p.resumeChan:
			return
		case <-p.stopChan:
			return
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// Submit 添加任务
func (p *WorkerPool) Submit(task TaskFunc) {
	p.wg.Add(1)
	p.tasks <- task
}

// Wait 等待所有任务完成
func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

// Stop 停止所有 worker
func (p *WorkerPool) Stop() {
	close(p.stopChan)
	close(p.results) // 关闭结果通道，避免 GetResult 中泄露
}

// Pause 暂停调度
func (p *WorkerPool) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.paused {
		p.paused = true
		for i := 0; i < p.concurrency; i++ {
			p.pauseChan <- struct{}{}
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
			p.resumeChan <- struct{}{}
		}
	}
}

// GetResult 获取一个用于读取任务返回值的 channel
func (p *WorkerPool) GetResult() <-chan *fuzzTypes.Reaction {
	return p.results
}

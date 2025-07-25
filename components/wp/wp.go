package wp

import (
	"FuzzGIU/components/fuzzTypes"
	"sync"
	"time"
)

type Task func() *fuzzTypes.Reaction

type WorkerPool struct {
	tasks       chan Task
	results     chan *fuzzTypes.Reaction
	concurrency int

	wg         sync.WaitGroup
	stopChan   chan struct{}
	pauseChan  chan struct{}
	resumeChan chan struct{}
	paused     bool
	mu         sync.Mutex
}

var Wp *WorkerPool

// New 创建一个新的协程池
func New(concurrency int) *WorkerPool {
	Wp = &WorkerPool{
		tasks:       make(chan Task, 8192),
		results:     make(chan *fuzzTypes.Reaction, 8192),
		concurrency: concurrency,
		stopChan:    make(chan struct{}),
		pauseChan:   make(chan struct{}),
		resumeChan:  make(chan struct{}),
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
		case <-p.stopChan:
			return
		case <-p.pauseChan:
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
func (p *WorkerPool) Submit(task Task) {
	p.wg.Add(1)
	p.tasks <- task
}

// Wait 等待所有任务完成
func (p *WorkerPool) Wait(maxTime time.Duration) bool {
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
	close(p.stopChan)
	close(p.results) // 关闭结果通道，避免 GetResChan 中泄露
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

func (p *WorkerPool) Resize(size int) {
	if size == p.concurrency {
		return
	} else {
		if size > p.concurrency {
			for i := 0; i < size-p.concurrency; i++ {
				go p.worker()
			}
		} else {
			for i := 0; i < p.concurrency-size; i++ {
				p.stopChan <- struct{}{}
			}
		}
		p.concurrency = size
	}
}

func (p *WorkerPool) GetSingleResult() *fuzzTypes.Reaction {
	select {
	case r := <-p.results:
		return r
	default:
		return nil
	}
}

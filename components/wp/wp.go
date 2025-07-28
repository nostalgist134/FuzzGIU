package wp

import (
	"sync"
	"time"

	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
)

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
	status int8
	mu     sync.Mutex
	cond   *sync.Cond
}

var CurrentWp *WorkerPool

// New 创建一个新的协程池
func New(concurrency int) *WorkerPool {
	wp := &WorkerPool{
		concurrency: concurrency,
	}
	wp.cond = sync.NewCond(&wp.mu)
	CurrentWp = wp
	return CurrentWp
}

// Start 启动所有 worker
func (p *WorkerPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == statRunning {
		return
	}
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
		// 先检查退出信号
		select {
		case <-p.quit:
			return
		default:
		}

		// 检查状态并决定是否等待
		p.mu.Lock()
		switch p.status {
		case statStop:
			p.mu.Unlock()
			return
		case statPause:
			// 处于暂停状态，等待唤醒
			p.cond.Wait()
			p.mu.Unlock()
		case statRunning:
			// 处于运行状态，释放锁并尝试获取任务
			p.mu.Unlock()

			// 阻塞等待任务或退出信号，避免忙循环
			select {
			case task, ok := <-p.tasks:
				if !ok {
					continue
				}
				result := task()
				// 非阻塞发送结果，避免通道满时阻塞worker
				select {
				case p.results <- result:
				default:
					// 可以在这里添加结果处理失败的逻辑
				}
				p.wg.Done()
			case <-p.quit:
				return
			case <-time.After(10 * time.Millisecond):
				// 短暂超时，让worker有机会检查状态变化
			}
		}
	}
}

func (p *WorkerPool) waitForResume() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for p.status == statPause {
		select {
		case <-p.quit:
			return
		default:
			p.cond.Wait()
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
	p.mu.Lock()
	if p.status == statStop {
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
	p.status = statStop
	p.cond.Broadcast()
}

// Pause 暂停调度
func (p *WorkerPool) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == statRunning {
		p.status = statPause
		p.cond.Broadcast()
	}
}

// Resume 恢复调度
func (p *WorkerPool) Resume() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.status == statPause {
		p.status = statRunning
		p.cond.Broadcast()
	}
}

func (p *WorkerPool) Resize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if size == p.concurrency || size < 0 || p.status == statStop {
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

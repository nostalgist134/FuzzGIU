package libfgiu

import (
	"context"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
	"time"
)

type jobExecutor func(*fuzzCtx.JobCtx) (int, time.Duration, []*fuzzTypes.Fuzz, error)

type result struct {
	jid        int
	timeLapsed time.Duration
	newJobs    []*fuzzTypes.Fuzz
	err        error
}

// jobExecPool 丐版协程池，用于并发执行fuzz任务
type jobExecPool struct {
	mu          sync.Mutex
	concurrency int
	jobQueue    chan *fuzzCtx.JobCtx
	results     chan result
	runningJobs sync.Map
	executor    jobExecutor
	// 说实话还是不如用一个stop chan
	quitCtx context.Context
	cancel  context.CancelFunc
}

func nopExec(*fuzzCtx.JobCtx) (int, time.Duration, []*fuzzTypes.Fuzz, error) {
	return 0, 0, nil, nil
}

func newJobExecPool(concurrency int, resultLen int, quitCtx context.Context,
	cancelFunc context.CancelFunc) *jobExecPool {
	if concurrency < 1 { // 懒得再写error了，直接返回nil得
		return nil
	}
	return &jobExecPool{
		concurrency: concurrency,
		jobQueue:    make(chan *fuzzCtx.JobCtx, concurrency*2),
		executor:    nopExec,
		results:     make(chan result, resultLen),
		quitCtx:     quitCtx,
		cancel:      cancelFunc,
	}
}

func (jp *jobExecPool) registerExecutor(executor jobExecutor) {
	jp.mu.Lock()
	defer jp.mu.Unlock()
	jp.executor = executor
}

func (jp *jobExecPool) submit(jobCtx *fuzzCtx.JobCtx) bool {
	select {
	case jp.jobQueue <- jobCtx:
		return true
	default:
		return false
	}
}

func (jp *jobExecPool) getResult() (res result, ok bool) {
	select {
	case res = <-jp.results:
		ok = true
		return
	default:
		return result{}, false
	}
}

func (jp *jobExecPool) worker() {
	for {
		select {
		case job := <-jp.jobQueue:
			jp.runningJobs.Store(job.JobId, job)
			jid, timeLapsed, newJobs, err := jp.executor(job)
			jp.results <- result{jid, timeLapsed, newJobs, err}
			jp.runningJobs.Delete(job.JobId)
		case <-jp.quitCtx.Done():
			return
		}
	}
}

func (jp *jobExecPool) start() {
	for i := 0; i < jp.concurrency; i++ {
		go jp.worker()
	}
}

func (jp *jobExecPool) stop() {
	jp.cancel()
}

func (jp *jobExecPool) findRunningJobById(jid int) (job *fuzzCtx.JobCtx, exist bool) {
	if j, ok := jp.runningJobs.Load(jid); ok {
		return j.(*fuzzCtx.JobCtx), true
	}
	return nil, false
}

package libfgiu

import (
	"context"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"log"
	"sync"
	"time"
)

const (
	FuzzerStatInit    = 0
	FuzzerStatRunning = 1
	FuzzerStatUsed    = 2
)

var (
	errJobQuFull     = errors.New("job queue is full")
	errJobPoolNil    = errors.New("job pool is nil")
	errFuzzerStopped = errors.New("fuzzer is already stopped")
	errNotStarted    = errors.New("fuzzer is not started yet")
)

type pendingJob struct {
	job                *fuzzTypes.Fuzz
	parentId           int
	parentUnderHttpApi bool
}

// Fuzzer 用来执行模糊测试任务，允许多个任务并发执行，内部维护一个任务协程池
// 注意：此结构是一次性的，也就是说调用Stop之后就不能再调用Start启动，否则可
// 能导致未定义行为，必须使用NewFuzzer重新获取
type Fuzzer struct {
	stat        int8
	statMux     sync.Mutex
	jp          *jobExecPool
	cancel      context.CancelFunc
	ctx         context.Context
	pendingJobs []pendingJob
	s           *httpService
}

type WebApiConfig struct {
	ServAddr     string
	Tls          bool
	CertFileName string
	CertKeyName  string
}

func (f *Fuzzer) daemon() {
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
			// 消费本地 pending 队列
			done := 0
			for ; done < len(f.pendingJobs); done++ {
				p := f.pendingJobs[done]
				jc, err := fuzz.NewJobCtx(p.job, p.parentId, f.ctx, f.cancel)
				if err != nil {
					log.Printf("failed to init job, reason: %v", err)
				}
				if !f.jp.submit(jc) {
					break // 池满，立即停止
				}
			}
			// 保留未提交部分
			f.pendingJobs = f.pendingJobs[done:]

			// 消费协程池回捞结果
			res, ok := f.jp.getResult()
			if !ok {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			done = 0
			for ; done < len(res.newJobs); done++ {
				j := res.newJobs[done]
				jc, err := fuzz.NewJobCtx(j, res.jid, f.ctx, f.cancel)
				if err != nil {
					log.Printf("failed to init job, reason: %v", err)
					continue
				}
				if !f.jp.submit(jc) {
					break // 池满，立即停
				}
			}

			if done < len(res.newJobs) {
				for _, j := range res.newJobs[done:] {
					f.pendingJobs = append(f.pendingJobs, pendingJob{
						job:      j,
						parentId: res.jid,
					})
				}
			}

			// 3. 防止 CPU 空转
			if len(f.pendingJobs) == 0 && !ok {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

// NewFuzzer 获取一个Fuzzer对象
// concurrency 指定任务并发池的大小
// apiConf 指定是否启动api模式及启动的配置，只要指定了这个参数就会启动api，但是若指定nil会使用默认配置
func NewFuzzer(concurrency int, apiConf ...*WebApiConfig) (*Fuzzer, error) {
	quitCtx, cancel := context.WithCancel(context.Background())

	f := &Fuzzer{
		stat:   FuzzerStatInit,
		jp:     newJobExecPool(concurrency, concurrency*20, quitCtx, cancel),
		ctx:    quitCtx,
		cancel: cancel,
	}

	if len(apiConf) > 0 {
		var err error
		err = f.startHttpApi(apiConf[0])
		if err != nil {
			return f, err
		}
	}

	f.jp.registerExecutor(fuzz.DoJobByCtx)
	return f, nil
}

// Do 用于阻塞运行一个fuzz任务
func (f *Fuzzer) Do(job *fuzzTypes.Fuzz) (jid int, timeLapsed time.Duration, newJobs []*fuzzTypes.Fuzz, err error) {
	if f.Status() == FuzzerStatUsed {
		return 0, 0, nil, errFuzzerStopped
	}
	var jobCtx *fuzzCtx.JobCtx
	jobCtx, err = fuzz.NewJobCtx(job, 0, f.ctx, f.cancel)
	if err != nil {
		return
	}
	jid, timeLapsed, newJobs, err = fuzz.DoJobByCtx(jobCtx)
	return
}

// Submit 用于非阻塞执行一个fuzz任务（提交到任务池中）
func (f *Fuzzer) Submit(job *fuzzTypes.Fuzz) error {
	f.statMux.Lock()
	switch f.stat {
	case FuzzerStatInit:
		return errNotStarted
	case FuzzerStatUsed:
		return errFuzzerStopped
	default:
	}
	f.statMux.Unlock()

	if f.jp == nil {
		return errJobPoolNil
	}
	if err := fuzz.ValidateJob(job); err != nil {
		return err
	}
	jc, err := fuzz.NewJobCtx(job, 0, f.ctx, f.cancel) // 使用submit提交的job其parentId都为0，代表最上层
	if err != nil {
		return err
	}

	if !f.jp.submit(jc) {
		return errJobQuFull
	}
	return nil
}

// Start 启动Fuzzer的任务池，在此之后可使用Submit方法向其中提交任务
func (f *Fuzzer) Start() *Fuzzer {
	f.statMux.Lock()
	defer f.statMux.Unlock()
	switch f.stat {
	case FuzzerStatRunning, FuzzerStatUsed:
		return f
	default:
	}
	if f.jp == nil {
		return f
	}
	f.jp.start()
	go f.daemon()
	f.stat = FuzzerStatRunning
	return f
}

// Stop 停止fuzzer的运行，并停止所有任务的运行
func (f *Fuzzer) Stop() {
	f.statMux.Lock()
	defer f.statMux.Unlock()
	if f.stat == FuzzerStatUsed {
		return
	}
	f.stat = FuzzerStatUsed
	f.s.e.Shutdown(f.ctx)
	f.cancel()
}

// Status 获取fuzzer当前的状态
func (f *Fuzzer) Status() int8 {
	f.statMux.Lock()
	defer f.statMux.Unlock()
	return f.stat
}

// GetJob 获取当前协程池中一个正在运行的任务的任务上下文，并且标记1次占用，防止使用时就被关闭
// 注意，获取到jobCtx后禁止更改，否则可能出现并发安全问题，并且需要手动调用jobCtx.Release
// 释放，否则这个job永远不会完成
func (f *Fuzzer) GetJob(jid int) (jobCtx *fuzzCtx.JobCtx, ok bool) {
	if f.jp == nil {
		return
	}
	jobCtx, ok = f.jp.findRunningJobById(jid)
	if !ok {
		return
	}
	jobCtx.Occupy()
	return
}

// GetJobIds 获取当前任务池中运行的所有任务
func (f *Fuzzer) GetJobIds() []int {
	if f.jp == nil {
		return nil
	}
	return f.jp.getRunningJobIds()
}

// StopJob 停止一个任务
func (f *Fuzzer) StopJob(jid int) error {
	if f.jp == nil {
		return errJobPoolNil
	}
	jc, ok := f.jp.findRunningJobById(jid)
	if !ok {
		return fmt.Errorf("job#%d not exist", jid)
	}
	jc.Stop()
	return nil
}

func (f *Fuzzer) GetApiToken() string {
	if f.s != nil {
		return f.s.accessToken
	}
	return ""
}

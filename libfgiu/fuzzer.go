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
	job      *fuzzTypes.Fuzz
	parentId int
}

// Fuzzer 用来执行模糊测试任务，允许多个任务并发执行，内部维护一个任务协程池
// 注意：此结构是一次性的，也就是说调用Stop之后就不能再调用Start启动，否则可
// 能导致未定义行为，必须使用NewFuzzer重新获取
// 这个结构体可以在其它的go代码中使用，只要遵循上面的原则就行
type Fuzzer struct {
	stat        int8
	statMux     sync.Mutex
	jp          *jobExecPool
	cancel      context.CancelFunc
	ctx         context.Context
	pendingJobs []pendingJob
	muPending   sync.Mutex
	s           *httpService
}

// WebApiConfig 若要使用web api，指定web api的设置
type WebApiConfig struct {
	ServAddr     string
	TLS          bool
	CertFileName string
	CertKeyName  string
}

func (f *Fuzzer) daemon() {
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
			var (
				i            = 0
				jobCtx       *fuzzCtx.JobCtx
				err          error
				lastConsumed bool
			)
			f.muPending.Lock()
			// 循环1：消耗pendingJobs队列
			for ; i < len(f.pendingJobs); i++ {
				p := f.pendingJobs[i]
				jobCtx, err = fuzz.NewJobCtx(p.job, p.parentId, f.ctx, f.cancel)
				if err != nil {
					log.Printf("[FUZZER] failed to init job: %v\n", err)
				} else if !f.jp.submit(jobCtx) {
					err = jobCtx.Close() // 关闭输出上下文，避免资源泄漏
					if err != nil {
						log.Printf("[JOB_CONTEXT] close error: %v\n", err)
					}
					break
				} else if i == len(f.pendingJobs)-1 { // pendingJobs最后一个元素也被消费了，标记为true
					lastConsumed = true
				}
			}

			if lastConsumed { // 如果最后一个元素也被消费了，则置空
				f.pendingJobs = []pendingJob{}
			} else { // 切到未消耗的部分
				f.pendingJobs = f.pendingJobs[i:]
			}

			// 循环2：从jp的结果队列中取衍生任务，直到jp结果队列为空
			for {
				res, ok := f.jp.getResult()
				if !ok {
					break
				}
				lastConsumed = false
				// 尝试提交newJobs切片中任务，直到jp队列满或者提交全部完成
				for i = 0; i < len(res.newJobs); i++ {
					jobCtx, err = fuzz.NewJobCtx(res.newJobs[i], res.jid, f.ctx, f.cancel)
					if err != nil {
						log.Printf("[FUZZER] failed to init job: %v\n", err)
					} else if !f.jp.submit(jobCtx) {
						err = jobCtx.Close()
						if err != nil {
							log.Printf("[JOB_CONTEXT] close error: %v\n", err)
						}
						break
					} else if i == len(res.newJobs)-1 {
						lastConsumed = true
					}
				}
				if !lastConsumed { // 将剩余未提交的任务存入pendingJobs队列中
					for ; i < len(res.newJobs); i++ {
						f.pendingJobs = append(f.pendingJobs, pendingJob{res.newJobs[i], res.jid})
					}
				}
			}
			f.muPending.Unlock()
			time.Sleep(25 * time.Millisecond) // 短暂休眠，避免空转
		}
	}
}

// NewFuzzer 获取一个Fuzzer对象，如果需要，可以将libfgiu包作为库使用，大部分的细节已经包装好了
// concurrency 指定任务并发池的大小
// apiConf 指定是否启动api模式及启动的配置，只要指定了这个参数就会启动api，如果要自定义配置，则指定具体配置
func NewFuzzer(concurrency int, apiConf ...WebApiConfig) (*Fuzzer, error) {
	quitCtx, cancel := context.WithCancel(context.Background())

	jp, err := newJobExecPool(concurrency, concurrency*20, quitCtx, cancel)
	if err != nil {
		return nil, err
	}

	f := &Fuzzer{
		stat:   FuzzerStatInit,
		jp:     jp,
		ctx:    quitCtx,
		cancel: cancel,
	}

	f.jp.registerExecutor(fuzz.DoJobByCtx) // 使用DoJobByCtx作为执行函数
	if len(apiConf) > 0 {
		err = f.startHttpApi(apiConf[0])
		if err != nil {
			return f, err
		}
	}
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
// 返回提交任务的id和错误
func (f *Fuzzer) Submit(job *fuzzTypes.Fuzz) (int, error) {
	f.statMux.Lock()
	switch f.stat {
	case FuzzerStatInit:
		return -1, errNotStarted
	case FuzzerStatUsed:
		return -1, errFuzzerStopped
	default:
	}
	f.statMux.Unlock()

	if f.jp == nil {
		return -1, errJobPoolNil
	}
	jc, err := fuzz.NewJobCtx(job, 0, f.ctx, f.cancel) // 使用submit提交的job其parentId都为0，代表最上层
	if err != nil {
		return -1, err
	}

	if !f.jp.submit(jc) {
		return -1, errJobQuFull
	}
	return jc.JobId, nil
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

// Wait 等待fuzzer对象直到其不再执行任何任务
func (f *Fuzzer) Wait() {
	for {
		if f.s != nil {
			f.s.wait()
		}
		f.jp.wait()
		if len(f.jp.jobQueue) == 0 && len(f.jp.results) == 0 {
			f.muPending.Lock()
			// 等待结束条件：
			// 1.jobExecPool的任务队列为空
			// 2.jobExecPool的结果队列为空
			// 3.pendingJobs队列长度为0
			if len(f.pendingJobs) == 0 {
				f.muPending.Unlock()
				return
			}
			f.muPending.Unlock()
			time.Sleep(50 * time.Millisecond)
			continue
		}
	}
}

// Stop 停止fuzzer的运行，并停止所有任务的运行
func (f *Fuzzer) Stop() error {
	f.statMux.Lock()
	defer f.statMux.Unlock()
	if f.stat == FuzzerStatUsed {
		return errFuzzerStopped
	}
	f.stat = FuzzerStatUsed
	var err error
	if f.s != nil && f.s.e != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
		defer cancel()
		err = f.s.e.Shutdown(ctx)
	}
	f.cancel()
	return err
}

// Status 获取fuzzer当前的状态
func (f *Fuzzer) Status() int8 {
	f.statMux.Lock()
	defer f.statMux.Unlock()
	return f.stat
}

// GetJob 获取当前协程池中一个正在运行的任务的任务上下文，并且标记1次占用，防止使用时就被关闭
// 注意，目前版本暂不支持获取到jobCtx后更改，否则可能出现并发安全问题，获取后需要手动调用
// jobCtx.Release方法释放，否则会导致关闭时阻塞
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
	return jc.Close()
}

// GetApiToken 如果启动了http api模式，获取api模式的token
func (f *Fuzzer) GetApiToken() string {
	if f.s != nil {
		return f.s.accessToken
	}
	return ""
}

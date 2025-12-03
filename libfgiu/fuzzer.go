package libfgiu

import (
	"context"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"log"
	"sync"
	"time"
)

const (
	FuzzerStatInit    = 0
	FuzzerStatRunning = 1
	FuzzerStatStopped = 2
)

var (
	errJobQuFull            = errors.New("job queue is full")
	errJobPoolNil           = errors.New("job pool is nil")
	errFuzzerStopped        = errors.New("fuzzer is already stopped")
	errHttpApiDisallowTview = errors.New("tview output is not allowed in http api mode")
)

// Fuzzer 用来执行模糊测试任务，允许多个任务并发执行，内部维护一个任务协程池
// 注意：此结构是一次性的，也就是说调用Stop之后就不能再调用Start启动，否则可
// 能导致未定义行为，必须使用NewFuzzer重新获取（因为我实在是懒得调它的状态机
// 了，越调可能出现的问题越多，就这样写还更好）
// Fuzzer对象只能通过NewFuzzer函数获取，不能直接声明后就使用
// 这个结构体可以在其它的go代码中使用，只要遵循上面的原则就行
type Fuzzer struct {
	stat        int8
	muStat      sync.Mutex
	idle        bool
	condIdle    *sync.Cond
	jp          *jobExecPool
	cancel      context.CancelFunc
	quitCtx     context.Context
	pendingJobs []*fuzzCtx.JobCtx
	muApi       sync.Mutex
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
		case <-f.quitCtx.Done():
			for _, pendingJob := range f.pendingJobs {
				pendingJob.Close()
			}
			f.pendingJobs = nil
			return
		default:
			var (
				i            = 0
				idle         = true
				jobCtx       *fuzzCtx.JobCtx
				err          error
				lastConsumed bool
				ctx          context.Context
				cancel       context.CancelFunc
			)

			if len(f.pendingJobs) > 0 { // 如果有任务缓冲队列非空，判断为非空闲
				idle = false
			}
			// 循环1：消耗pendingJobs队列，由于pendingJobs只在此函数中调用，且此函数不启动多次，因此无需加锁
			for ; i < len(f.pendingJobs); i++ {
				p := f.pendingJobs[i]
				if !f.jp.submit(p) {
					break
				} else if i == len(f.pendingJobs)-1 { // pendingJobs最后一个元素也被消费了，标记为true
					lastConsumed = true
				}
			}

			if lastConsumed { // 如果最后一个元素也被消费了，则置空
				f.pendingJobs = []*fuzzCtx.JobCtx{}
			} else { // 切到未消耗的部分
				f.pendingJobs = f.pendingJobs[i:]
			}

			// 循环2：从jp的结果队列中取衍生任务，直到jp结果队列为空
			for {
				res, ok := f.jp.getResult()
				if !ok {
					break
				}
				if len(res.newJobs) > 0 { // 如果有衍生任务，判断为非空闲
					idle = false
				}
				// 尝试提交newJobs切片中任务，若提交失败，则将任务存入pendingJobs队列
				for i = 0; i < len(res.newJobs); i++ {
					ctx, cancel = context.WithCancel(f.quitCtx)
					jobCtx, err = fuzz.NewJobCtx(res.newJobs[i], res.jid, ctx, cancel)
					if err != nil {
						log.Printf("[FUZZER] failed to init job: %v\n", err)
					} else if !f.jp.submit(jobCtx) {
						f.pendingJobs = append(f.pendingJobs, jobCtx)
					}
				}
			}

			// 若任务池在运行/等待任务数非0、结果队列或任务队列非空，判断为非空闲
			if f.jp.activePendingCnt.Load() > 0 || len(f.jp.results) > 0 || len(f.jp.jobQueue) > 0 {
				idle = false
			}
			f.condIdle.L.Lock()
			if f.idle = idle; idle { // 若为空闲状态
				f.condIdle.L.Unlock()
				f.condIdle.Broadcast()
			} else {
				f.condIdle.L.Unlock()
			}
			time.Sleep(50 * time.Millisecond) // 短暂休眠，避免空转
		}
	}
}

// NewFuzzer 获取一个Fuzzer对象
// concurrency 指定任务并发池的大小
// apiConf 指定是否启动api模式及启动的配置，只要指定了这个参数就会启动api，如果要自定义配置，则指定具体配置
func NewFuzzer(concurrency int, apiConf ...WebApiConfig) (*Fuzzer, error) {
	quitCtx, cancel := context.WithCancel(context.Background())
	jpCtx, jpCancel := context.WithCancel(quitCtx)

	jp, err := newJobExecPool(concurrency, concurrency*20, jpCtx, jpCancel)
	if err != nil {
		cancel()
		return nil, err
	}

	f := &Fuzzer{
		stat:    FuzzerStatInit,
		jp:      jp,
		quitCtx: quitCtx,
		cancel:  cancel,
		idle:    true,
	}
	f.condIdle = sync.NewCond(new(sync.Mutex))

	f.jp.registerExecutor(fuzz.DoJobByCtx) // 使用DoJobByCtx作为执行函数
	if len(apiConf) > 0 {
		err = f.StartHttpApi(apiConf[0])
		if err != nil {
			return f, err
		}
	}
	return f, nil
}

// Do 用于阻塞运行一个fuzz任务
func (f *Fuzzer) Do(job *fuzzTypes.Fuzz) (jid int, timeLapsed time.Duration, newJobs []*fuzzTypes.Fuzz, err error) {
	if f.Status() == FuzzerStatStopped {
		return 0, 0, nil, errFuzzerStopped
	}
	var jobCtx *fuzzCtx.JobCtx
	ctx, cancel := context.WithCancel(f.quitCtx)
	jobCtx, err = fuzz.NewJobCtx(job, 0, ctx, cancel)
	if err != nil {
		return
	}
	jid, timeLapsed, newJobs, err = fuzz.DoJobByCtx(jobCtx)
	return
}

// Submit 用于非阻塞执行一个fuzz任务（提交到任务池中）
// 返回提交任务的id和错误
func (f *Fuzzer) Submit(job *fuzzTypes.Fuzz) (int, error) {
	f.muStat.Lock()
	defer f.muStat.Unlock()
	if f.stat == FuzzerStatStopped {
		return -1, errFuzzerStopped
	}

	if f.jp == nil {
		return -1, errJobPoolNil
	}
	if job != nil && job.Control.OutSetting.ToWhere&outputFlag.OutToTview != 0 && f.s != nil {
		return -1, errHttpApiDisallowTview
	}
	ctx, cancel := context.WithCancel(f.quitCtx)
	jc, err := fuzz.NewJobCtx(job, 0, ctx, cancel) // 使用submit提交的job其parentId都为0，代表最上层
	if err != nil {
		return -1, err
	}

	f.condIdle.L.Lock()
	defer f.condIdle.L.Unlock()
	if !f.jp.submit(jc) {
		return -1, errors.Join(jc.Close(), errJobQuFull)
	}
	f.idle = false
	return jc.JobId, nil
}

// Start 启动Fuzzer的任务池，在此之后可使用Submit方法向其中提交任务
func (f *Fuzzer) Start() error {
	f.muStat.Lock()
	defer f.muStat.Unlock()
	switch f.stat {
	case FuzzerStatRunning:
		return nil
	case FuzzerStatStopped:
		return errFuzzerStopped
	default:
	}
	f.jp.start()
	go f.daemon()
	f.stat = FuzzerStatRunning
	return nil
}

// Wait 等待fuzzer对象直到其处于空闲状态（即没有任务执行，也没有待执行的任务）
func (f *Fuzzer) Wait() {
	if f.Status() == FuzzerStatStopped {
		return
	}
	if f.s != nil {
		f.s.wait()
	}

	f.jp.wait()
	f.condIdle.L.Lock()
	for !f.idle {
		f.condIdle.Wait()
	}
	f.condIdle.L.Unlock()
}

// Stop 停止fuzzer的运行，并停止所有任务的运行
func (f *Fuzzer) Stop() error {
	f.muStat.Lock()
	defer f.muStat.Unlock()
	if f.stat == FuzzerStatStopped {
		return errFuzzerStopped
	}
	f.stat = FuzzerStatStopped
	f.jp.stop()
	f.cancel()

	f.condIdle.L.Lock()
	f.idle = true
	f.condIdle.L.Unlock()
	f.condIdle.Broadcast()
	return f.StopHttpApi()
}

// Status 获取fuzzer当前的状态
func (f *Fuzzer) Status() int8 {
	f.muStat.Lock()
	defer f.muStat.Unlock()
	return f.stat
}

// GetJob 获取当前协程池中一个正在运行的任务的任务上下文，并且标记1次占用，防止使用时就被关闭
// 注意，目前版本暂不支持获取到jobCtx后更改，否则可能出现并发安全问题，获取后需要手动调用
// jobCtx.Release方法释放，否则会导致关闭时阻塞
func (f *Fuzzer) GetJob(jid int) (jobCtx *fuzzCtx.JobCtx, ok bool) {
	if f.Status() == FuzzerStatStopped {
		return nil, false
	}
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
	if f.jp == nil || f.Status() == FuzzerStatStopped {
		return nil
	}
	return f.jp.getRunningJobIds()
}

// StopJob 停止一个任务
func (f *Fuzzer) StopJob(jid int) error {
	if f.Status() == FuzzerStatStopped {
		return errFuzzerStopped
	}
	if f.jp == nil {
		return errJobPoolNil
	}
	jc, ok := f.jp.findRunningJobById(jid)
	if !ok {
		return fmt.Errorf("job#%d not exist", jid)
	}
	return jc.Close()
}

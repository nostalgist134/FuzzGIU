package libfgiu

import (
	"context"
	"errors"
	"github.com/nostalgist134/FuzzGIU/components/fuzz"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"log"
	"net/http"
	"time"
)

const (
	FuzzModeSubmit = iota
	FuzzModePassive

	defPsvAddr = "0.0.0.0:11451"
)

type pendingJob struct {
	job      *fuzzTypes.Fuzz
	parentId int
}

// Fuzzer 用来执行模糊测试任务
type Fuzzer struct {
	runMode     int8
	jp          *jobExecPool
	cancel      context.CancelFunc
	ctx         context.Context
	pendingJobs []pendingJob
}

func (f *Fuzzer) daemon() {
	for {
		select {
		case <-f.ctx.Done():
			return
		default:
			// 1. 先消费本地 pending 队列
			done := 0
			for ; done < len(f.pendingJobs); done++ {
				p := f.pendingJobs[done]
				jc, err := fuzz.NewJobCtx(p.job, p.parentId)
				if err != nil {
					log.Fatal(err)
				}
				if !f.jp.submit(jc) {
					break // 池满，立即停止
				}
			}
			// 保留未提交部分（done 指向第一个未成功的任务）
			f.pendingJobs = f.pendingJobs[done:]

			// 2. 消费协程池回捞结果
			res, ok := f.jp.getResult()
			if !ok {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			done = 0
			for ; done < len(res.newJobs); done++ {
				j := res.newJobs[done]
				jc, err := fuzz.NewJobCtx(j, res.jid)
				if err != nil {
					log.Fatal(err)
				}
				if !f.jp.submit(jc) {
					break // 池满，立即停
				}
			}
			// 把没提交完的重新塞回 pending 队列
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
				time.Sleep(10 * time.Millisecond)
			}
		}
	}
}

func handler(w *http.ResponseWriter, r *http.Request) {

}

func NewFuzzer(runMode int8, concurrency int, passiveAddr ...string) (*Fuzzer, error) {
	quitCtx, cancel := context.WithCancel(context.Background())
	f := &Fuzzer{
		runMode: runMode,
		jp:      newJobExecPool(concurrency, concurrency*20, quitCtx, cancel),
		ctx:     quitCtx,
	}
	if runMode == FuzzModePassive {
		if len(passiveAddr) > 0 {
			http.ListenAndServe(passiveAddr[0], nil)
		}
	}
	f.jp.registerExecutor(fuzz.DoJobByCtx)
	return f, nil
}

// Do 用于阻塞运行一个fuzz任务
func (f *Fuzzer) Do(job *fuzzTypes.Fuzz) (jid int, timeLapsed time.Duration, newJobs []*fuzzTypes.Fuzz, err error) {
	var jobCtx *fuzzCtx.JobCtx
	jobCtx, err = fuzz.NewJobCtx(job, 0)
	if err != nil {
		return
	}
	jid, timeLapsed, newJobs, err = fuzz.DoJobByCtx(jobCtx)
	return
}

// Submit 用于非阻塞执行一个fuzz任务（提交到任务池中）
func (f *Fuzzer) Submit(job *fuzzTypes.Fuzz) error {
	jc, err := fuzz.NewJobCtx(job, 0) // 使用submit提交的job其parentId都为0，代表最上层
	if err != nil {
		return err
	}
	if !f.jp.submit(jc) {
		return errors.New("job queue is full")
	}
	return nil
}

func (f *Fuzzer) Start() *Fuzzer {
	f.jp.start()
	go f.daemon()
	return f
}

func (f *Fuzzer) Stop() {
	f.cancel()
}

// GetJob 获取当前协程池中一个正在运行的任务
func (f *Fuzzer) GetJob(jid int) (*fuzzCtx.JobCtx, bool) {
	return f.jp.findRunningJobById(jid)
}

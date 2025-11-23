package fuzz

import (
	"context"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stageReact"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"github.com/nostalgist134/FuzzGIU/components/rp"
	"github.com/nostalgist134/FuzzGIU/components/tmplReplace"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

var curJobId = atomic.Int64{}

// getJobId 获取一个可用的jobId
func getJobId() int {
	return int(curJobId.Add(1))
}

// trySubmit 尝试提交任务，若提交失败，则先从队列中取出所有结果并处理
func trySubmit(jobCtx *fuzzCtx.JobCtx, task *fuzzCtx.TaskCtx, whichExec int8) (stopJob bool,
	newJobs []*fuzzTypes.Fuzz) {
	routinePool := jobCtx.RP

	var newJobsFromHandle []*fuzzTypes.Fuzz

	for !routinePool.Submit(task, whichExec, time.Millisecond*10) {
		// 若处于暂停状态，则不消耗结果
		routinePool.WaitResume()

		// 将结果队列全部消耗
		for r := routinePool.GetSingleResult(); r != nil; r = routinePool.GetSingleResult() {
			stopJob, _, newJobsFromHandle = handleReaction(jobCtx, r)
			newJobs = append(newJobs, newJobsFromHandle...)

			// 若确定jobStop，就可以不用再取结果了，直接返回
			if stopJob {
				return
			}
		}
	}
	return
}

// handleReaction 根据fuzz设置处理反应
func handleReaction(jobCtx *fuzzCtx.JobCtx, r *fuzzTypes.Reaction) (stopJob bool, addReq bool,
	newJobs []*fuzzTypes.Fuzz) {
	defer resourcePool.PutReaction(r)

	if r.Flag&fuzzTypes.ReactAddJob != 0 && r.NewJob != nil {
		k, p := stageReact.GetReactTraceInfo(r)
		if k != nil && p != nil {
			jobCtx.OutputCtx.LogFmtMsg("Job#%d task with %s:%s added Job", jobCtx.JobId, k, p)
		}
		newJobs = append(newJobs, r.NewJob)
	}

	if r.Flag&fuzzTypes.ReactStopJob != 0 {
		jobCtx.OutputCtx.LogFmtMsg("Job#%d stopped by react", jobCtx.JobId)
		jobCtx.RP.Clear()
		stopJob = true
	}

	if r.Flag&fuzzTypes.ReactAddReq != 0 && r.NewReq != nil {
		addReq = true

		newTask := fuzzCtx.GetTaskCtx()
		*newTask = fuzzCtx.TaskCtx{
			JobCtx:      jobCtx,
			ViaReaction: r,
		}

		// 由于尝试提交任务的过程中，需要执行任务列表中的任务，过程中可能产生新的任务，需要处理
		var newJobsFromTrySub []*fuzzTypes.Fuzz
		stopJob, newJobsFromTrySub = trySubmit(jobCtx, newTask, rp.ExecMinor)
		if newJobsFromTrySub != nil {
			newJobs = append(newJobs, newJobsFromTrySub...)
		}

		// task总数加1
		jobCtx.OutputCtx.Counter.Add(counter.CntrTask, counter.FieldTotal, 1)
	}
	return
}

// drainRp 消耗协程池中的所有任务和结果
func drainRp(jobCtx *fuzzCtx.JobCtx) []*fuzzTypes.Fuzz {
	routinePool := jobCtx.RP

	var addReq, stopJob bool
	var newJobsFromHandle []*fuzzTypes.Fuzz

	newJobs := make([]*fuzzTypes.Fuzz, 0)

	for {
		canStop := true // canStop 标记了结果是否已经消耗完毕
		// 循环1：跑到Rp等待不阻塞（也就是任务队列为空）为止
		for !routinePool.Wait(time.Millisecond * 10) {
			for r := routinePool.GetSingleResult(); r != nil; r = routinePool.GetSingleResult() {
				stopJob, addReq, newJobsFromHandle = handleReaction(jobCtx, r)
				newJobs = append(newJobs, newJobsFromHandle...)
				if stopJob {
					return newJobs
				}
				if addReq {
					canStop = false
				}
			}
		}

		// 循环2：把结果队列的结果全部消耗完毕
		for r := routinePool.GetSingleResult(); r != nil; r = routinePool.GetSingleResult() {
			stopJob, addReq, newJobsFromHandle = handleReaction(jobCtx, r)
			if newJobsFromHandle != nil {
				newJobs = append(newJobs, newJobsFromHandle...)
			}
			if stopJob {
				return newJobs
			}
			if addReq {
				canStop = false
			}
		}

		// 若上面两个循环都跑完了，也没有添加新请求，这种情况下任务队列和结果队列均为空，没可能再有新请求，结束循环
		if canStop {
			break
		}
	}
	return newJobs
}

// getKeywordsPayloads 遍历map，获取一个关键字排列顺序以及对应顺序的payload列表长度与payload列表集
// map的遍历是无序的，不过代码对keyword顺序无所谓，只要确定一次顺序后之后都按这个顺序来就行
func getKeywordsPayloads(job *fuzzTypes.Fuzz) (keywords []string, lengths []int, payloadLists [][]string) {
	n := len(job.Preprocess.PlMeta)
	keywords = make([]string, 0, n)
	lengths = make([]int, 0, n)
	payloadLists = make([][]string, 0, n)
	for kw, pt := range job.Preprocess.PlMeta {
		keywords = append(keywords, kw)
		lengths = append(lengths, len(pt.PlList))
		payloadLists = append(payloadLists, pt.PlList)
	}
	return
}

// tryGetUrlScheme 尝试获取url scheme，若整个fuzz过程中url的scheme不会变化（不包含任何fuzz keyword）则可将其缓存
// 从而避免在SendRequest中反复调用url.Parse消耗资源（主要通过scheme选择内置请求发送器还是插件，因此找scheme就好）
func tryGetUrlScheme(req *fuzzTypes.Req, keywords []string) string {
	u, err := url.Parse(req.URL)
	if err != nil {
		return ""
	}
	scheme := u.Scheme
	for _, k := range keywords {
		if strings.Contains(scheme, k) {
			return ""
		}
	}
	return scheme
}

// doJobInter 执行一个fuzz任务，返回其衍生任务集（衍生任务不会在个函数内提交运行）
func doJobInter(jobCtx *fuzzCtx.JobCtx) (timeLapsed time.Duration, newJobs []*fuzzTypes.Fuzz, err error) {
	timeStart := time.Now()

	job := jobCtx.Job
	outCtx := jobCtx.OutputCtx
	routinePool := jobCtx.RP

	// 递归边界
	if job.React.RecursionControl.RecursionDepth > job.React.RecursionControl.MaxRecursionDepth {
		return
	}

	defer func() {
		timeLapsed = time.Since(timeStart)
		err = errors.Join(err, outCtx.LogFmtMsg("job#%d completed, time spent: %v", jobCtx.JobId, timeLapsed))
	}()

	genPayloads(jobCtx)

	// fuzz关键字的处理
	var (
		iterLength   int
		keywords     []string
		payloadLists [][]string
		lengths      []int
	)

	keywords, lengths, payloadLists = getKeywordsPayloads(job)
	parsedTmpl := tmplReplace.ParseReqTmpl(&job.Preprocess.ReqTemplate, keywords) // 请求模板

	iter := &(job.Control.IterCtrl)
	// 确认迭代终点与使用的执行函数
	if iter.End == 0 {
		iterName := job.Control.IterCtrl.Iterator.Name
		// sniper模式或者递归模式，仅允许单个fuzz关键字
		if iterName == "sniper" || job.React.RecursionControl.MaxRecursionDepth > 0 {
			if iter.Iterator.Name == "sniper" { // sniper模式的迭代长度=关键字的payload列表长度*关键字出现次数
				iterLength = parsedTmpl.KeywordCount(0) * lengths[0]
			}
			routinePool.RegisterExecutor(taskSingleKeyword, rp.ExecMajor)
		} else { // 多关键字模式
			iterLength = iterLen(job.Control.IterCtrl.Iterator, lengths)
			routinePool.RegisterExecutor(taskMultiKeyword, rp.ExecMajor)
		}
		iter.End = iterLength
	}

	job = stagePreprocess.Preprocess(job, jobCtx.OutputCtx)

	routinePool.RegisterExecutor(taskNoKeywords, rp.ExecMinor)
	routinePool.Start()

	outCtx.Counter.Set(counter.CntrTask, counter.FieldTotal, iterLength)
	outCtx.Counter.Set(counter.CntrTask, counter.FieldCompleted, iter.Start)

	var plProcs = make([][]fuzzTypes.Plugin, len(keywords)) // payload处理器插件
	for i, keyword := range keywords {
		plProcs[i] = job.Preprocess.PlMeta[keyword].Processors
	}

	jobStop := false

	uScheme := tryGetUrlScheme(&job.Preprocess.ReqTemplate, keywords)

	iterIndexes := make([]int, len(keywords)) // payload下标组合，每次迭代时改变

	var newJobsTmp []*fuzzTypes.Fuzz

	var tc = fuzzCtx.TaskCtx{
		USchemeCache: uScheme,
		Keywords:     keywords,
		RepTmpl:      parsedTmpl,
		JobCtx:       jobCtx,
		PlProc:       plProcs,
	}

	// fuzz主循环，若循环尾为-1则代表无限循环，什么时候结束取决于迭代器逻辑
	for i := iter.Start; i < iter.End || iter.End == fuzzTypes.InfiniteLoop; i++ {
		// 只有进入fuzz循环了，才能停止任务（其实是我懒得设计那么多select了）
		select {
		case <-jobCtx.GlobCtx.Done():
			return
		default:
		}

		task := fuzzCtx.GetTaskCtx()
		*task = tc

		payloads := resourcePool.StringSlices.Get(len(keywords))

		task.IterInd = i
		// 根据迭代器决定迭代下标，递归/sniper模式不走这个分支
		if iter.Iterator.Name != "sniper" && job.React.RecursionControl.MaxRecursionDepth <= 0 {
			iterIndex(lengths, i, iterIndexes, iter.Iterator)

			hasValid := false
			for j, _ := range keywords { // 根据下标选择每个关键字对应的payload
				if iterIndexes[j] < 0 || iterIndexes[j] >= len(payloadLists[j]) {
					payloads[j] = ""
				} else {
					payloads[j] = payloadLists[j][iterIndexes[j]]
					hasValid = true
				}
			}

			// 若下标全为无效值，则认为迭代结束
			if !hasValid {
				break
			}

			task.Payloads = payloads
		} else { // sniper模式或者递归模式
			snipLen := len(payloadLists[0])
			payload := payloadLists[0][i%snipLen]
			payloads[0] = payload
			task.Payloads = payloads
			task.SnipLen = snipLen
		}

		jobStop, newJobsTmp = trySubmit(jobCtx, task, rp.ExecMajor)
		newJobs = append(newJobs, newJobsTmp...)
		if jobStop {
			return
		}

		time.Sleep(job.Control.Delay)

		// 任务提交后，从结果队列中取出所有结果并处理
		for r := routinePool.GetSingleResult(); r != nil; r = routinePool.GetSingleResult() {
			jobStop, _, newJobsTmp = handleReaction(jobCtx, r)
			newJobs = append(newJobs, newJobsTmp...)
			if jobStop {
				return
			}
		}
	}

	newJobsTmp = drainRp(jobCtx)
	newJobs = append(newJobs, newJobsTmp...)
	return
}

var (
	errNilJobProvided    = errors.New("nil job provided")
	errNilJobCtxProvided = errors.New("nil job context provided")
)

func NewJobCtx(job *fuzzTypes.Fuzz, parentId int, ctx context.Context,
	cancel context.CancelFunc) (jobCtx *fuzzCtx.JobCtx, err error) {
	if job == nil {
		err = errNilJobProvided
		return
	}

	if err = ValidateJob(job); err != nil { // 先校验当前job是否有效
		return nil, fmt.Errorf("failed to validate job: %v", err)
	}
	jid := getJobId()

	// 分配一个新的协程池
	var routinePool *rp.RoutinePool
	routinePool, err = rp.NewRp(job.Control.PoolSize)
	if err != nil {
		return
	}

	if cancel == nil || ctx == nil {
		ctx, cancel = context.WithCancel(context.Background())
	}

	jobCtx = &fuzzCtx.JobCtx{
		JobId:    jid,
		ParentId: parentId,
		RP:       routinePool,
		Job:      job,
		Cancel:   cancel,
		GlobCtx:  ctx,
	}

	// 新建一个输出上下文
	var outCtx *output.Ctx
	outCtx, err = output.NewOutputCtx(&job.Control.OutSetting, jobCtx, jid)
	if err != nil {
		if outCtx != nil {
			err = errors.Join(err, outCtx.Close())
		}
		return
	}
	jobCtx.OutputCtx = outCtx

	// 预加载插件
	if err = preLoadJobPlugin(job); err != nil {
		err = errors.Join(err, outCtx.LogFmtMsg("Job#%d preload plugins failed: %v", jid, err))
		err = errors.Join(err, outCtx.Close())
		return
	}

	return
}

// DoJobByCtx 根据jobCtx执行任务，返回其衍生出的所有任务
func DoJobByCtx(jobCtx *fuzzCtx.JobCtx) (jid int, timeLapsed time.Duration, subJobs []*fuzzTypes.Fuzz, err error) {
	if jobCtx == nil {
		err = errNilJobCtxProvided
		return
	} else if jobCtx.Job == nil {
		err = errNilJobProvided
		return
	}
	jid = jobCtx.JobId
	timeLapsed, subJobs, err = doJobInter(jobCtx)
	err = jobCtx.Close()
	return
}

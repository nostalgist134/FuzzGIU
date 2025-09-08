package common

import (
	"sync/atomic"
	"time"
)

const (
	CntTask   = 0 // CntTask 获取task个数
	TotalTask = 1 // TotalTask 获取task总数
	CntJob    = 2 // CntJob 获取job个数
	TotalJob  = 3 // TotalJob 获取job总数
)

func init() {
	globCounter.timeStart = time.Now()
	// 速率每1秒更新一次
	go func() {
		for {
			start := atomic.LoadInt64(&globCounter.taskCounter.count)
			time.Sleep(1 * time.Second)
			if rateNow := int32(atomic.LoadInt64(&globCounter.taskCounter.count) - start); rateNow >= 0 {
				atomic.StoreInt32(&globCounter.rate, rateNow)
			}
		}
	}()
}

// SetTaskTotal 设置task计数器的总数
func SetTaskTotal(total int64) {
	atomic.StoreInt64(&globCounter.taskCounter.total, total)
}

// SetJobTotal 设置job计数器的总数
func SetJobTotal(total int64) {
	atomic.StoreInt64(&globCounter.jobCounter.total, total)
}

// AddTaskCounter 将task计数器的个数加一
func AddTaskCounter() {
	atomic.AddInt64(&globCounter.taskCounter.count, 1)
}

// AddJobCounter 将job计数器的个数加一
func AddJobCounter() {
	atomic.AddInt64(&globCounter.jobCounter.count, 1)
}

func ClearTaskCounter() {
	atomic.StoreInt64(&globCounter.taskCounter.count, 0)
}

// GetCounter 获取计数器的数据
func GetCounter() []int64 {
	return []int64{atomic.LoadInt64(&globCounter.taskCounter.count),
		atomic.LoadInt64(&globCounter.taskCounter.total),
		atomic.LoadInt64(&globCounter.jobCounter.count),
		atomic.LoadInt64(&globCounter.jobCounter.total)}
}

// GetCounterValue 获取单个计数器数据
func GetCounterValue(which int8) int64 {
	switch which {
	case CntTask:
		return atomic.LoadInt64(&globCounter.taskCounter.count)
	case TotalTask:
		return atomic.LoadInt64(&globCounter.taskCounter.total)
	case CntJob:
		return atomic.LoadInt64(&globCounter.jobCounter.count)
	case TotalJob:
		return atomic.LoadInt64(&globCounter.jobCounter.total)
	}
	return -1
}

// GetTimeLapsed 获取自计数开始的时间
func GetTimeLapsed() time.Duration {
	return time.Since(globCounter.timeStart)
}

// GetCurrentRate 获取当前的速率
func GetCurrentRate() int32 {
	return atomic.LoadInt32(&globCounter.rate)
}

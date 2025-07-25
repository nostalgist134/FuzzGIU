package common

import (
	"sync/atomic"
	"time"
)

func init() {
	globCounter.timeStart = time.Now()
	// 速率每1秒更新一次
	go func() {
		for {
			start := atomic.LoadInt64(&globCounter.taskCounter.count)
			time.Sleep(1 * time.Second)
			atomic.StoreInt32(&globCounter.rate, int32(atomic.LoadInt64(&globCounter.taskCounter.count)-start))
		}
	}()
}

// SetTaskCounter 设置task计数器的总数
func SetTaskCounter(total int64) {
	atomic.StoreInt64(&globCounter.taskCounter.total, total)
}

// SetJobCounter 设置job计数器的总数
func SetJobCounter(total int64) {
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

// GetCounterSingle 获取单个计数器数据
func GetCounterSingle(which int8) int64 {
	switch which {
	case 0:
		return atomic.LoadInt64(&globCounter.taskCounter.count)
	case 1:
		return atomic.LoadInt64(&globCounter.taskCounter.total)
	case 2:
		return atomic.LoadInt64(&globCounter.jobCounter.count)
	case 3:
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

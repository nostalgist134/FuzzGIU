package counter

import (
	"sync"
	"sync/atomic"
	"time"
)

type Progress struct {
	Completed int64 `json:"completed,omitempty" xml:"completed,omitempty"`
	Total     int64 `json:"total,omitempty" xml:"total,omitempty"`
}

type Counter struct {
	StartTime    time.Time `json:"start_time,omitempty" xml:"start_time,omitempty"`
	TaskRate     int64     `json:"task_rate,omitempty" xml:"task_rate,omitempty"`
	JobProgress  Progress  `json:"job_progress,omitempty" xml:"job_progress,omitempty"`
	TaskProgress Progress  `json:"task_progress,omitempty" xml:"task_progress,omitempty"`
	ticker       *time.Ticker
	mu           sync.Mutex
	stop         chan struct{}
}

// StartRecordTaskRate 开始计算速率
func (c *Counter) StartRecordTaskRate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.ticker != nil {
		return
	}
	c.ticker = time.NewTicker(time.Second) // 速率每秒统计1次
	go func() {
		lastCompleted := c.Get(CntrTask, FieldCompleted)
		c.stop = make(chan struct{})
		for {
			select {
			case <-c.stop:
				c.ticker.Stop()
				c.ticker = nil
				return
			case <-c.ticker.C:
				currentCompleted := c.Get(CntrTask, FieldCompleted)
				delta := currentCompleted - lastCompleted
				atomic.StoreInt64(&c.TaskRate, int64(delta))
				lastCompleted = currentCompleted
			}
		}
	}()
}

// StopRecordTaskRate 结束计算速率
func (c *Counter) StopRecordTaskRate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.stop == nil {
		return
	}
	select {
	case c.stop <- struct{}{}:
	default:
	}
	return
}

func (c *Counter) GetTaskRate() int {
	return int(atomic.LoadInt64(&c.TaskRate))
}

// Get 获取计数器的特定字段
func (c *Counter) Get(whichCounter int8, whichField int8) int {
	var pProgress *Progress

	switch whichCounter {
	case CntrJob:
		pProgress = &c.JobProgress
	case CntrTask:
		pProgress = &c.TaskProgress
	default:
		return -1
	}

	switch whichField {
	case FieldCompleted:
		return int(atomic.LoadInt64(&pProgress.Completed))
	case FieldTotal:
		return int(atomic.LoadInt64(&pProgress.Total))
	}

	return -1
}

// Complete 将job或task的completed字段加1
func (c *Counter) Complete(whichCounter int8) {
	var pProgress *Progress
	switch whichCounter {
	case CntrJob:
		pProgress = &c.JobProgress
	case CntrTask:
		pProgress = &c.TaskProgress
	default:
		return
	}
	atomic.AddInt64(&pProgress.Completed, 1)
}

// Set 设置计数器的特定字段
func (c *Counter) Set(whichCounter int8, whichField int8, val int) {
	var pProgress *Progress

	switch whichCounter {
	case CntrJob:
		pProgress = &c.JobProgress
	case CntrTask:
		pProgress = &c.TaskProgress
	default:
		return
	}

	switch whichField {
	case FieldCompleted:
		atomic.StoreInt64(&pProgress.Completed, int64(val))
	case FieldTotal:
		atomic.StoreInt64(&pProgress.Total, int64(val))
	}
}

func (c *Counter) Add(whichCounter int8, whichField int8, delta int) {
	var pProgress *Progress

	switch whichCounter {
	case CntrJob:
		pProgress = &c.JobProgress
	case CntrTask:
		pProgress = &c.TaskProgress
	default:
		return
	}

	switch whichField {
	case FieldCompleted:
		atomic.AddInt64(&pProgress.Completed, int64(delta))
	case FieldTotal:
		atomic.AddInt64(&pProgress.Total, int64(delta))
	}
}

// Clear 将计数器的特定字段清0
func (c *Counter) Clear(whichCounter int8, whichField int8) {
	c.Set(whichCounter, whichField, 0)
}

// TimeAnchor 设置开始时间
func (c *Counter) TimeAnchor() {
	c.StartTime = time.Now()
}

// TimeFromAnchor 从开始时间经过了多久
func (c *Counter) TimeFromAnchor() time.Duration {
	return time.Since(c.StartTime)
}

// Snapshot 获取当前计数器的状态
func (c *Counter) Snapshot() Counter {
	return Counter{
		StartTime: time.Time{},
		TaskRate:  int64(c.GetTaskRate()),
		JobProgress: Progress{
			Completed: int64(c.Get(CntrJob, FieldCompleted)),
			Total:     int64(c.Get(CntrJob, FieldTotal)),
		},
		TaskProgress: Progress{
			Completed: int64(c.Get(CntrTask, FieldCompleted)),
			Total:     int64(c.Get(CntrTask, FieldTotal)),
		},
		ticker: nil,
		mu:     sync.Mutex{},
		stop:   nil,
	}
}

package counter

import (
	"fmt"
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
	Errors       Progress  `json:"errors,omitempty" xml:"errors,omitempty"`
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
	case CntrErrors:
		pProgress = &c.Errors
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
	case CntrErrors:
		pProgress = &c.Errors
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
	case CntrErrors:
		pProgress = &c.Errors
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
	case CntrErrors:
		pProgress = &c.Errors
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
		StartTime: c.StartTime,
		TaskRate:  int64(c.GetTaskRate()),
		TaskProgress: Progress{
			Completed: int64(c.Get(CntrTask, FieldCompleted)),
			Total:     int64(c.Get(CntrTask, FieldTotal)),
		},
		Errors: Progress{
			Completed: int64(c.Get(CntrErrors, FieldCompleted)),
			Total:     int64(c.Get(CntrErrors, FieldTotal)),
		},
	}
}

// formatDuration 把 time.Duration 格式化为 00:00:00 格式（支持负数）
func formatDuration(d time.Duration) string {
	// 1. 提取总秒数（避免浮点数精度问题）
	totalSec := int64(d / time.Second)
	if totalSec == 0 {
		return "00:00:00"
	}

	// 2. 处理负数符号
	sign := ""
	if totalSec < 0 {
		sign = "-"
		totalSec = -totalSec // 转为正数计算
	}

	// 3. 拆解小时、分钟、秒
	hours := totalSec / 3600
	remainingSec := totalSec % 3600
	minutes := remainingSec / 60
	seconds := remainingSec % 60

	// 4. 格式化（%02d 确保不足两位补0）
	return fmt.Sprintf("%s%02d:%02d:%02d", sign, hours, minutes, seconds)
}

func (c *Counter) ToFmt() string {
	s := c.Snapshot()
	return fmt.Sprintf("tasks:[%d / %d]   errors:[%d]   rate:[%d t/s]   duration:[%s]", s.TaskProgress.Completed,
		s.TaskProgress.Total, s.Errors.Completed, s.TaskRate, formatDuration(time.Since(s.StartTime)))
}

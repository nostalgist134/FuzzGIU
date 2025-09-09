package memOutput

import (
	"encoding/json"
	"github.com/nostalgist134/FuzzGIU/components/output/common"
	"sync"
)

const (
	segmentSize = 8192
	maxObjects  = 262144

	logSegmentSize = 1024  // 每段日志条数
	logMaxObjects  = 32768 // 最多保存的日志条数
)

var muMem sync.Mutex

var segments [][]json.RawMessage
var headIndex, count int

var muLog sync.Mutex

// 日志缓冲区
var logSegments = make([][]string, 0, logMaxObjects/logSegmentSize)
var logHeadIndex, logCount int

func InitOutput() {
	segments = make([][]json.RawMessage, 0, maxObjects/segmentSize)
	headIndex = 0
	count = 0
}

// Output 接收 common.OutObj 并以 JSON 形式存储
func Output(obj *common.OutObj) {
	muMem.Lock()
	defer muMem.Unlock()

	data, err := json.Marshal(obj)
	if err != nil {
		// 序列化失败直接忽略
		return
	}

	segNum := headIndex / segmentSize
	offset := headIndex % segmentSize

	if segNum >= len(segments) {
		segments = append(segments, make([]json.RawMessage, segmentSize))
	}

	segments[segNum][offset] = data

	headIndex++
	if count < maxObjects {
		count++
	} else {
		headIndex = headIndex % maxObjects
	}
}

// GetAllObjects 返回所有 JSON 数据
func GetAllObjects() []json.RawMessage {
	muMem.Lock()
	defer muMem.Unlock()

	if count == 0 {
		return []json.RawMessage{}
	}

	result := make([]json.RawMessage, count)
	for i := 0; i < count; i++ {
		idx := (headIndex - count + i + maxObjects) % maxObjects
		segNum := idx / segmentSize
		offset := idx % segmentSize
		result[i] = segments[segNum][offset]
	}
	return result
}

// GetObjects 返回指定范围的 JSON 数据
func GetObjects(start, end int) []json.RawMessage {
	muMem.Lock()
	defer muMem.Unlock()

	if start < 0 || end > count || start >= end {
		return []json.RawMessage{}
	}

	result := make([]json.RawMessage, end-start)
	for i := 0; i < end-start; i++ {
		idx := (headIndex - count + start + i + maxObjects) % maxObjects
		segNum := idx / segmentSize
		offset := idx % segmentSize
		result[i] = segments[segNum][offset]
	}
	return result
}

// Log 添加一条日志
func Log(msg string) {
	muLog.Lock()
	defer muLog.Unlock()

	segNum := logHeadIndex / logSegmentSize
	offset := logHeadIndex % logSegmentSize

	// 懒分配
	if segNum >= len(logSegments) {
		logSegments = append(logSegments, make([]string, logSegmentSize))
	}

	logSegments[segNum][offset] = msg

	logHeadIndex++
	if logCount < logMaxObjects {
		logCount++
	} else {
		logHeadIndex = logHeadIndex % logMaxObjects
	}
}

// GetAllLogs 获取所有日志
func GetAllLogs() []string {
	muLog.Lock()
	defer muLog.Unlock()

	if logCount == 0 {
		return []string{}
	}

	result := make([]string, logCount)
	for i := 0; i < logCount; i++ {
		idx := (logHeadIndex - logCount + i + logMaxObjects) % logMaxObjects
		segNum := idx / logSegmentSize
		offset := idx % logSegmentSize
		result[i] = logSegments[segNum][offset]
	}
	return result
}

// GetLogs 获取指定范围日志
func GetLogs(start, end int) []string {
	muLog.Lock()
	defer muLog.Unlock()

	if start < 0 || end > logCount || start >= end {
		return []string{}
	}

	result := make([]string, end-start)
	for i := 0; i < end-start; i++ {
		idx := (logHeadIndex - logCount + start + i + logMaxObjects) % logMaxObjects
		segNum := idx / logSegmentSize
		offset := idx % logSegmentSize
		result[i] = logSegments[segNum][offset]
	}
	return result
}

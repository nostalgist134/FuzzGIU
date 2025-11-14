package rp

import (
	"fmt"
	"sync"
)

var rps = sync.Pool{New: func() any { return newRoutinePool(64) }}

// NewRp 从协程池对象池中取出一个并发数为concurrency的rp
func NewRp(concurrency int) (*RoutinePool, error) {
	if concurrency <= 0 {
		return nil, fmt.Errorf("invalid concurrency %d", concurrency)
	}

	p := rps.Get().(*RoutinePool)
	p.Resize(concurrency)
	return p, nil
}

// putRp 将协程池放入协程池对象池中
func putRp(p *RoutinePool) {
	p.Clear()
	rps.Put(p)
}

package resourcePool

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"sync"
)

// SlicePool 通用切片对象池，支持任意类型切片[]T
type SlicePool[T any] struct {
	pool   sync.Pool // 底层对象池
	defCap int       // 初始创建切片时的默认容量（减少首次分配开销）
}

// NewSlicePool 创建新的切片对象池
// defCap: 池为空时，新建切片的默认容量
func NewSlicePool[T any](defCap int) *SlicePool[T] {
	return &SlicePool[T]{
		pool: sync.Pool{
			New: func() any {
				// 初始创建长度0、容量defCap的切片，提升首次复用概率
				return make([]T, 0, defCap)
			},
		},
		defCap: defCap,
	}
}

// Get 从池中获取长度为length的切片
// length < 0 时返回nil
func (p *SlicePool[T]) Get(length int) []T {
	if length < 0 {
		return nil
	}

	// 从池获取切片并断言类型
	slice := p.pool.Get().([]T)

	// 若容量不足，将原切片放回池（不限制容量，即使小也可能被后续小需求复用），新建切片
	if cap(slice) < length {
		p.Put(slice)
		return make([]T, length)
	}

	// 容量足够时直接截断复用
	return slice[:length]
}

// Put 将使用完毕的切片放回池中（截断为长度0以保留容量）
func (p *SlicePool[T]) Put(slice []T) {
	if slice == nil {
		return // 忽略nil切片
	}
	p.pool.Put(slice[:0])
}

// StringSlices 字符串切片池
var StringSlices = NewSlicePool[string](20)

// AnySlices any切片池
var AnySlices = NewSlicePool[any](20)

// IntSlices int切片池
var IntSlices = NewSlicePool[int](20)

var FieldSlices = NewSlicePool[fuzzTypes.Field](20)

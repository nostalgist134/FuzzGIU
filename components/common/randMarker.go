package common

import (
	"math/rand/v2"
	"strings"
	"sync"
)

var (
	pool = sync.Pool{New: func() any { return new(strings.Builder) }}
	dict = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+="
)

// RandMarker 生成一个长度为16的随机字符串
func RandMarker() string {
	sb := pool.Get().(*strings.Builder)
	defer func() { sb.Reset(); pool.Put(sb) }()
	for i := 0; i < 16; i++ {
		sb.WriteByte(dict[rand.IntN(64)])
	}
	return sb.String()
}

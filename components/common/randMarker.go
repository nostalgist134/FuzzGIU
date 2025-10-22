package common

import (
	"math/rand"
	"strings"
)

// RandMarker 生成一个长度为12为的随机字符串
func RandMarker() string {
	dict := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	sb := strings.Builder{}
	for i := 0; i < 12; i++ {
		sb.WriteByte(dict[rand.Intn(len(dict))])
	}
	return sb.String()
}

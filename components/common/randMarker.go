package common

import (
	"math/rand/v2"
	"strings"
)

// RandMarker 生成一个长度为16的随机字符串
func RandMarker() string {
	dict := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+="
	sb := strings.Builder{}
	for i := 0; i < 16; i++ {
		sb.WriteByte(dict[rand.IntN(len(dict))])
	}
	return sb.String()
}

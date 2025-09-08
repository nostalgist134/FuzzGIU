package common

import "sync"

var StringSlices = sync.Pool{
	New: func() any { return make([]string, 0) },
}

func GetStringSlice(length int) []string {
	if length < 0 {
		return nil
	}
	slice := StringSlices.Get().([]string)
	if cap(slice) < length {
		slice = make([]string, length) // 新建
	} else {
		slice = slice[:length] // 复用
	}
	return slice
}

func PutStringSlice(toPut []string) {
	StringSlices.Put(toPut)
}

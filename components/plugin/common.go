package plugin

import (
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"path/filepath"
	"unsafe"
)

func callError(relPath string, p fuzzTypes.Plugin, err error) {
	output.Logf(common.OutputToWhere, "call %s failed: %v", filepath.Join(relPath, p.Name), err)
}

// bytes2Strings 将动态链接库返回的bytes转化为string切片
func bytes2Strings(ptrBytes uintptr) []string {
	if ptrBytes == uintptr(0) {
		return []string{}
	}
	ret := make([]string, 0)
	j := *(*int32)(unsafe.Pointer(ptrBytes)) // 前4位为切片的长度
	for i := uintptr(0); j > 0; j-- {
		length := *(*int32)(unsafe.Pointer(ptrBytes + 4 + i))
		bytesSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptrBytes+8+i)), length)
		ret = append(ret, string(bytesSlice))
		i += 4 + uintptr(length)
	}
	return ret
}

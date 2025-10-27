package plugin

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"path/filepath"
	"unsafe"
)

const sizeInt = 8

func pluginError(outCtx *output.Ctx, relPath string, p fuzzTypes.Plugin, err error) {
	if outCtx == nil {
		return
	}
	outCtx.LogFmtMsg("call %s failed: %v", filepath.Join(relPath, p.Name), err)
}

// bytes2Strings 将动态链接库返回的bytes转化为string切片
func bytes2Strings(ptrBytes uintptr) []string {
	if ptrBytes == 0 {
		return []string{}
	}

	j := *(*int32)(unsafe.Pointer(ptrBytes)) // 前4位为切片的长度
	ret := resourcePool.StringSlices.Get(0)

	for i := uintptr(0); j > 0; j-- {
		length := *(*int32)(unsafe.Pointer(ptrBytes + 4 + i))
		bytesSlice := unsafe.Slice((*byte)(unsafe.Pointer(ptrBytes+8+i)), length)
		ret = append(ret, string(bytesSlice))
		i += 4 + uintptr(length)
	}

	return ret
}

func bytes2Ints(ptrBytes uintptr, out []int) {
	if ptrBytes == 0 {
		return
	}

	j := *(*int32)(unsafe.Pointer(ptrBytes))

	k := 0
	for i := uintptr(0); j > 0 && k < len(out); j-- {
		out[k] = *(*int)(unsafe.Pointer(ptrBytes + 4 + i*uintptr(sizeInt)))
		k++
		i++
	}
}

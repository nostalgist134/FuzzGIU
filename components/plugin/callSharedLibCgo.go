//go:build windows

package plugin

import (
	"errors"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"math"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

type dllRecord struct {
	dll  *syscall.DLL
	proc *syscall.Proc
}

var records = sync.Map{}
var mu = sync.Mutex{}

// callSharedLib 调用插件的PluginWrapper函数 windows
// 调用约定 pluginFunction(...[jsonData, lenJsonData], 用户指定参数),
func callSharedLib(plugin fuzzTypes.Plugin, relPath string, jsons ...[]byte) uintptr {
	var dll *syscall.DLL
	var proc *syscall.Proc
	var err error
	// 加载dll
	if p, ok := records.Load(relPath + plugin.Name); ok {
		proc = p.(*dllRecord).proc
	} else {
		mu.Lock()
		// 如果多个协程进锁，第一个进锁的协程已经加载了，就不用再加载一遍
		if p, ok = records.Load(relPath + plugin.Name); ok {
			proc = p.(*dllRecord).proc
		} else {
			dll, err = syscall.LoadDLL(filepath.Join(BaseDir, relPath, plugin.Name+binSuffix))
			if err != nil {
				panic(err)
			}
			proc, err = dll.FindProc(pluginEntry)
			if err != nil {
				panic(err)
			}
			records.Store(relPath+plugin.Name, &dllRecord{dll: dll, proc: proc})
		}
		mu.Unlock()
	}
	args := make([]uintptr, 0)
	if len(jsons) > 0 && jsons[0] != nil {
		args = append(args, uintptr(unsafe.Pointer(&jsons[0][0])))
		args = append(args, uintptr(len(jsons[0])))
	}
	if len(jsons) > 1 && jsons[1] != nil {
		args = append(args, uintptr(unsafe.Pointer(&jsons[1][0])))
		args = append(args, uintptr(len(jsons[1])))
	}
	for _, arg := range plugin.Args {
		switch v := arg.(type) {
		case int:
			args = append(args, uintptr(v))
		case float64:
			args = append(args, uintptr(math.Float64bits(v)))
		case bool:
			if arg == false {
				args = append(args, uintptr(0))
			} else {
				args = append(args, uintptr(1))
			}
		case string:
			s := arg.(string)
			args = append(args, uintptr(unsafe.Pointer(&s)))
		}
	}
	r1, _, err := proc.Call(args...)
	var errno syscall.Errno
	if err != nil && (!errors.As(err, &errno) || errno != 0) {
		panic(err)
	}
	return r1
}

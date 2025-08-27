//go:build windows

package plugin

import (
	"encoding/json"
	"errors"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"math"
	"path/filepath"
	"strings"
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
	strCache := make([]string, 0)
	for i := 0; i < len(plugin.Args); i++ {
		switch plugin.Args[i].(type) {
		case int:
			args = append(args, uintptr(plugin.Args[i].(int)))
		case float64:
			args = append(args, uintptr(math.Float64bits(plugin.Args[i].(float64))))
		case bool:
			if plugin.Args[i] == false {
				args = append(args, uintptr(0))
			} else {
				args = append(args, uintptr(1))
			}
		case string:
			// 将字符串存到切片中，每个字符串的地址不同，就不会导致参数污染
			strCache = append(strCache, plugin.Args[i].(string))
			args = append(args, uintptr(unsafe.Pointer(&strCache[len(strCache)-1])))
		}
	}
	r1, _, err := proc.Call(args...)
	var errno syscall.Errno
	if err != nil && (!errors.As(err, &errno) || errno != 0) {
		panic(err)
	}
	return r1
}

func parseJson(ptrJson uintptr) []byte {
	if ptrJson == uintptr(0) {
		return []byte{}
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(ptrJson+4)), *(*int32)(unsafe.Pointer(ptrJson)))
}

func stringFromPtr(strBytes uintptr) string {
	sb := strings.Builder{}
	sb.WriteString(unsafe.String((*byte)(unsafe.Pointer(strBytes+4)), *(*int32)(unsafe.Pointer(strBytes))))
	return sb.String()
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

// Preprocess 返回指向preprocessor处理后新生成的*Fuzz
func Preprocess(p fuzzTypes.Plugin, fuzz1 *fuzzTypes.Fuzz) *fuzzTypes.Fuzz {
	fuzzJson, err := json.Marshal(fuzz1)
	fuzzJson = parseJson(callSharedLib(p, RelPathPreprocessor, fuzzJson))
	newFuzz := new(fuzzTypes.Fuzz)
	err = json.Unmarshal(fuzzJson, newFuzz)
	if err != nil {
		panic(err)
	}
	return newFuzz
}

// PayloadGenerator 返回插件生成的payload切片
func PayloadGenerator(p fuzzTypes.Plugin) []string {
	ptrPayload := callSharedLib(p, RelPathPlGen)
	payloads := bytes2Strings(ptrPayload)
	return payloads
}

// PayloadProcessor 返回处理后的字符串
func PayloadProcessor(p fuzzTypes.Plugin) string {
	return stringFromPtr(callSharedLib(p, RelPathPlProc))
}

// SendRequest 根据sendMeat发送请求，并接收响应
func SendRequest(p fuzzTypes.Plugin, m *fuzzTypes.SendMeta) *fuzzTypes.Resp {
	reqJson, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	respJson := parseJson(callSharedLib(p, RelPathReqSender, reqJson))
	resp := new(fuzzTypes.Resp)
	err = json.Unmarshal(respJson, resp)
	if err != nil {
		panic(err)
	}
	return resp
}

// React 返回*reaction
func React(p fuzzTypes.Plugin, req *fuzzTypes.Req, resp *fuzzTypes.Resp) *fuzzTypes.Reaction {
	reqJson, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	reactionJson := parseJson(callSharedLib(p, RelPathReactor, reqJson, respJson))
	reaction := common.GetNewReaction()
	err = json.Unmarshal(reactionJson, reaction)
	if err != nil {
		panic(err)
	}
	return reaction
}

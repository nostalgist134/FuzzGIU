//go:build darwin || linux

package plugin

import (
	"encoding/json"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	goPlugin "plugin"
	"sync"
	"unsafe"
)

type symRecord struct {
	pluginFile *goPlugin.Plugin
	pluginFun  func(...any) uintptr
}

var symRecords = sync.Map{}
var mu = sync.Mutex{}

func callSharedLib(plugin fuzzTypes.Plugin, relPath string, jsons ...[]byte) uintptr {
	pName := relPath + plugin.Name
	var pw func(...any) uintptr
	// 尝试从缓存中加载插件函数
	if pRecord, ok := symRecords.Load(pName); ok {
		pw = pRecord.(symRecord).pluginFun
	} else { // 若失败则使用open打开
		mu.Lock()
		if pRecord, ok := symRecords.Load(pName); ok {
			pw = pRecord.(symRecord).pluginFun
		} else {
			p, err := goPlugin.Open(BaseDir + pName + binSuffix)
			if err != nil {
				return uintptr(0)
			}
			sym, err := p.Lookup(pluginEntry)
			if err != nil {
				return uintptr(0)
			}
			pw, ok = sym.(func(...any) uintptr)
			if !ok {
				return uintptr(0)
			}
			symRecords.Store(pName, symRecord{p, pw})
		}
		mu.Unlock()
	}
	args := make([]any, 0)
	if len(jsons) > 0 && jsons[0] != nil {
		args = append(args, jsons[0])
	}
	if len(jsons) > 1 && jsons[1] != nil {
		args = append(args, jsons[1])
	}
	args = append(args, plugin.Args...)
	return pw(args...)
}

func parseJson(ptrJson uintptr) []byte {
	if ptrJson == uintptr(0) {
		return []byte{}
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(ptrJson+4)), *(*int32)(unsafe.Pointer(ptrJson)))
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
	pStr := (*string)(unsafe.Pointer(callSharedLib(p, RelPathPlProc)))
	return *pStr
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

func Iterator(p fuzzTypes.Plugin, lengths []int, out []int, ind int) []int {
	return []int{}
}

func IterNum(p fuzzTypes.Plugin) int64 {
	return 0
}

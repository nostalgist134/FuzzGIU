//go:build darwin || linux

package plugin

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	fgpkCommon "github.com/nostalgist134/FuzzGIUPluginKit/cmd/common"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
	"path/filepath"
	goPlugin "plugin"
	"sync"
	"unsafe"
)

// 傻逼nostalgist134注释写这么一点点，自己看看能看得懂吗

type pluginRecord struct {
	pluginSelf *goPlugin.Plugin
	pluginFun  func(...any) ([]byte, error)
	pInfo      *convention.PluginInfo
}

var ErrFuncTypeIncorrect = errors.New("plugin entry incorrect, make sure your plugin is built using fgpk")

var registry = sync.Map{}
var mu = sync.Mutex{}

func callSharedLib(p fuzzTypes.Plugin, relPath string, jsons ...[]byte) ([]byte, error) {
	registryName := filepath.Join(relPath, p.Name)
	pluginPath := filepath.Join(BaseDir, relPath, p.Name+binSuffix)

	var pFun func(...any) ([]byte, error)

	var pRecord *pluginRecord

	// 尝试从缓存中加载插件函数
	if record, ok := registry.Load(registryName); ok {
		pRecord = record.(*pluginRecord)
		pFun = pRecord.pluginFun
	} else { // 若失败则使用open打开
		mu.Lock()
		if record, ok = registry.Load(registryName); ok {
			pRecord = record.(*pluginRecord)
			pFun = pRecord.pluginFun
		} else {
			goPlug, err := goPlugin.Open(pluginPath)
			if err != nil {
				return nil, err
			}

			sym, err := goPlug.Lookup(pluginEntry)
			if err != nil {
				return nil, err
			}

			pFun, ok = sym.(func(...any) ([]byte, error))
			if !ok {
				return nil, ErrFuncTypeIncorrect
			}

			pRecord = &pluginRecord{pluginSelf: goPlug, pluginFun: pFun}

			registry.Store(registryName, pRecord)
		}
		mu.Unlock()
	}

	if pi := pRecord.pInfo; pi != nil && len(pi.Params) != len(jsons)+len(p.Args) {
		return nil, fmt.Errorf("incorrect parameter count, expect %d, got %d", len(pi.Params),
			len(jsons)+len(p.Args))
	}

	// 参数列表
	args := make([]any, 0)
	if len(jsons) > 0 && jsons[0] != nil {
		args = append(args, jsons[0])

		if len(jsons) > 1 && jsons[1] != nil {
			args = append(args, jsons[1])
		}
	}
	args = append(args, p.Args...)

	return pFun(args...)
}

// PreLoad 预加载插件，并尝试获取插件的信息
func PreLoad(p fuzzTypes.Plugin, relPath string) (*convention.PluginInfo, error) {
	registryName := filepath.Join(relPath, p.Name)

	if record, ok := registry.Load(registryName); ok {
		return record.(*pluginRecord).pInfo, nil
	} else {
		mu.Lock()
		defer mu.Unlock()
		if record, ok = registry.Load(registryName); ok {
			return record.(*pluginRecord).pInfo, nil
		}
		pluginPath := filepath.Join(BaseDir, relPath, p.Name+binSuffix)

		goPlug, err := goPlugin.Open(pluginPath)
		if err != nil {
			return nil, err
		}
		sym, err := goPlug.Lookup(pluginEntry)
		if err != nil {
			return nil, err
		}
		pFun, ok := sym.(func(...any) ([]byte, error))
		if !ok {
			return nil, ErrFuncTypeIncorrect
		}

		pi, _ := fgpkCommon.GetPluginInfo(pluginPath)
		registry.Store(registryName, &pluginRecord{pluginSelf: goPlug, pInfo: pi, pluginFun: pFun})

		return pi, nil
	}
}

// Preprocess 返回指向preprocessor处理后新生成的*Fuzz
func Preprocess(p fuzzTypes.Plugin, fuzz1 *fuzzTypes.Fuzz) *fuzzTypes.Fuzz {
	fuzzJson, err := json.Marshal(fuzz1)
	if err != nil {
		return fuzz1
	}

	jsonBytes, err := callSharedLib(p, RelPathPreprocessor, fuzzJson)
	if err != nil {
		return fuzz1
	}

	newFuzz := new(fuzzTypes.Fuzz)

	err = json.Unmarshal(jsonBytes, newFuzz)
	if err != nil {
		return fuzz1
	}

	return newFuzz
}

// PayloadGenerator 返回插件生成的payload切片
func PayloadGenerator(p fuzzTypes.Plugin, outCtx *output.Ctx) []string {
	payloadsBytes, err := callSharedLib(p, RelPathPlGen)
	if err != nil {
		pluginError(outCtx, RelPathPlGen, p, err)
		return []string{}
	}
	return bytes2Strings(uintptr(unsafe.Pointer(&payloadsBytes[0])))
}

// PayloadProcessor 返回处理后的字符串
func PayloadProcessor(p fuzzTypes.Plugin, outCtx *output.Ctx) string {
	payload := p.Args[0].(string)
	strBytes, err := callSharedLib(p, RelPathPlProc)
	if err != nil {
		pluginError(outCtx, RelPathPlProc, p, err)
		return payload
	}
	return unsafe.String(&strBytes[0], len(strBytes))
}

// DoRequest 根据requestCtx发送请求，并接收响应
func DoRequest(p fuzzTypes.Plugin, r *fuzzTypes.RequestCtx) *fuzzTypes.Resp {
	resp := new(fuzzTypes.Resp)

	marshaled, err := json.Marshal(r)
	if err != nil {
		resp.ErrMsg = err.Error()
		return resp
	}

	jsonBytes, err := callSharedLib(p, RelPathReqSender, marshaled)

	err = json.Unmarshal(jsonBytes, resp)
	if err != nil {
		resp.ErrMsg = err.Error()
	}
	return resp
}

// React 返回*reaction
func React(p fuzzTypes.Plugin, req *fuzzTypes.Req, resp *fuzzTypes.Resp) *fuzzTypes.Reaction {
	rct := resourcePool.GetReaction()

	reqJson, err := json.Marshal(req)
	if err != nil {
		rct.Output.Msg = err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
		return rct
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		rct.Output.Msg = err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
		return rct
	}

	jsonBytes, err := callSharedLib(p, RelPathReactor, reqJson, respJson)
	if err != nil {
		rct.Output.Msg = err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
		return rct
	}

	err = json.Unmarshal(jsonBytes, rct)
	if err != nil {
		rct.Output.Msg = err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
	}

	return rct
}

// IterIndex
// 外部调用这个函数时的arg列表如下
// selectorIndex, ind, 用户自定义参数
// windows上可以用jsons参数传递，这个可不行
func IterIndex(p fuzzTypes.Plugin, lengths []int, out []int) {
	old := p.Args

	// lengths参数入arg列表
	p.Args = resourcePool.AnySlices.Get(len(p.Args) + 1)
	defer resourcePool.AnySlices.Put(p.Args)
	p.Args[0] = lengths
	copy(p.Args[1:], old) // 在此之后变成 lengths, selectorIndex, ind, ...，与插件模板指定顺序相同

	intsBytes, err := callSharedLib(p, RelPathIterator)
	if err != nil {
		for i := 0; i < len(out); i++ {
			out[i] = -1
		}
		return
	}

	bytes2Ints(uintptr(unsafe.Pointer(&intsBytes[0])), out)
	return
}

// IterLen
// 外部调用此函数时的参数列表如下
// selectorIterLen, 0, ...(由于插件编译后参数列表不能改变，且一个插件实际上只有一个调用入口，参数被写死了，所以填充一个0)
func IterLen(p fuzzTypes.Plugin, lengths []int) int {
	old := p.Args

	// lengths参数入arg列表
	p.Args = resourcePool.AnySlices.Get(len(p.Args) + 1)
	defer resourcePool.AnySlices.Put(p.Args)
	p.Args[0] = lengths
	copy(p.Args[1:], old) // 在此之后变成 lengths, selectorIterLen, 0, ...，与插件模板指定顺序相同

	iterLenBytes, err := callSharedLib(p, RelPathIterator)
	if err != nil {
		return -1
	}
	return *(*int)(unsafe.Pointer(&iterLenBytes[0]))
}

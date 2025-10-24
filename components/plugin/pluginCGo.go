//go:build windows

package plugin

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	fgpk "github.com/nostalgist134/FuzzGIUPluginKit/cmd/common"
	"github.com/nostalgist134/FuzzGIUPluginKit/convention"
	"github.com/nostalgist134/reusableBytes"
	"math"
	"path/filepath"
	"sync"
	"syscall"
	"unsafe"
)

type pluginRecord struct {
	dll   *syscall.DLL
	proc  *syscall.Proc
	pInfo *convention.PluginInfo
}

const errInteriorMarshal = "error in marshal/unmarshal in plugin, make sure your plugin act correctly"

var registry = sync.Map{}
var mu = sync.Mutex{}
var bp = reusablebytes.BytesPool{}
var uintptrSlices = sync.Pool{New: func() any { return make([]uintptr, 0) }}

func getUintptrSlice(length int) []uintptr {
	if length < 0 {
		return nil
	}
	slice := uintptrSlices.Get().([]uintptr)
	if cap(slice) < length {
		slice = make([]uintptr, length) // 新建
	} else {
		slice = slice[:length] // 复用
	}
	return slice
}

func putUintptrSlice(toPut []uintptr) {
	uintptrSlices.Put(toPut)
}

func init() {
	bp.Init(128, 131072, 256)
}

func getArgCnt(plugin fuzzTypes.Plugin, writeBuffer *reusablebytes.ReusableBytes, jsons ...[]byte) int {
	cnt := 0
	if len(jsons) > 0 && jsons[0] != nil {
		cnt += 2
		if len(jsons) > 1 && jsons[1] != nil {
			cnt += 2
		}
	}
	if writeBuffer != nil {
		cnt += 2
	}
	return len(plugin.Args) + cnt
}

// callSharedLib 调用插件的PluginWrapper函数 windows
func callSharedLib(plugin fuzzTypes.Plugin, relPath string, writeBuffer *reusablebytes.ReusableBytes,
	byteSlices ...[]byte) (int, error) {
	registryName := filepath.Join(relPath, plugin.Name)

	var dll *syscall.DLL
	var proc *syscall.Proc
	var err error

	var pRecord *pluginRecord

	// 加载dll
	if record, ok := registry.Load(registryName); ok {
		pRecord = record.(*pluginRecord)
		proc = pRecord.proc
	} else {
		mu.Lock()
		// 如果多个协程进锁，第一个进锁的协程已经加载了，就不用再加载一遍
		if record, ok = registry.Load(registryName); ok {
			pRecord = record.(*pluginRecord)
			proc = pRecord.proc
		} else {
			dll, err = syscall.LoadDLL(filepath.Join(BaseDir, relPath, plugin.Name+binSuffix))
			if err != nil {
				return 0, err
			}
			proc, err = dll.FindProc(pluginEntry)
			if err != nil {
				return 0, err
			}

			pRecord = &pluginRecord{dll: dll, proc: proc}
			registry.Store(registryName, pRecord)
		}
		mu.Unlock()
	}

	if pi := pRecord.pInfo; pi != nil && len(pi.Params) != len(byteSlices)+len(plugin.Args) {
		return 0, fmt.Errorf("incorrect parameter count, expect %d, got %d", len(pi.Params),
			len(byteSlices)+len(plugin.Args))
	}

	argList := getUintptrSlice(getArgCnt(plugin, writeBuffer, byteSlices...))
	defer putUintptrSlice(argList)

	i := 0

	// 填入缓冲区参数
	if writeBuffer != nil {
		argList[i] = uintptr(writeBuffer.UnsafeBuffer())
		argList[i+1] = uintptr(writeBuffer.Cap())
		i += 2
	}

	// 填入byte切片参数
	if len(byteSlices) > 0 && byteSlices[0] != nil {
		argList[i] = uintptr(unsafe.Pointer(&byteSlices[0][0]))
		argList[i+1] = uintptr(len(byteSlices[0]))
		i += 2

		if len(byteSlices) > 1 && byteSlices[1] != nil {
			argList[i] = uintptr(unsafe.Pointer(&byteSlices[1][0]))
			argList[i+1] = uintptr(len(byteSlices[1]))
			i += 2
		}
	}

	// 分配一个string切片用于存储string类型参数，避免参数污染
	strCnt := 0
	for _, arg := range plugin.Args {
		if _, ok := arg.(string); ok {
			strCnt++
		}
	}
	strCache := resourcePool.StringSlices.Get(strCnt)
	defer resourcePool.StringSlices.Put(strCache)

	// 填入普通参数
	j := 0
	for k := 0; k < len(plugin.Args); k++ {
		switch plugin.Args[k].(type) {
		case int:
			argList[i] = uintptr(plugin.Args[k].(int))
		case float64:
			argList[i] = uintptr(math.Float64bits(plugin.Args[k].(float64)))
		case bool:
			if plugin.Args[k] == false {
				argList[i] = uintptr(0)
			} else {
				argList[i] = uintptr(1)
			}
		case string:
			// 将字符串存到切片中，每个字符串的地址不同，就不会导致参数污染
			strCache[j] = plugin.Args[k].(string)
			argList[i] = uintptr(unsafe.Pointer(&strCache[j]))
			j++
		default: // 不支持的参数类型，填入0（NULL）
			err = errors.Join(err, fmt.Errorf("callSharedLib: unsupported arg %v, to nil", plugin.Args[k]))
			argList[i] = uintptr(0)
		}
		i++
	}

	// 调用dll函数并获取返回值
	r1, _, err2 := proc.Call(argList...)
	var errno syscall.Errno
	if err2 != nil && (!errors.As(err2, &errno) || errno != 0) {
		err = errors.Join(err, err2)
	}

	return int(r1), err
}

func ints2Bytes(ints []int) ([]byte, int32) {
	rb, id := bp.Get()

	lenBuffer := 4 + sizeInt*len(ints)
	rb.Resize(lenBuffer)
	byteSlice := unsafe.Slice((*byte)(rb.UnsafeBuffer()), lenBuffer)

	// 存储int切片的长度
	binary.LittleEndian.PutUint32(byteSlice, uint32(len(ints)))

	for k, i := range ints {
		binary.LittleEndian.PutUint64(byteSlice[4+k*sizeInt:], uint64(i))
	}
	return byteSlice, id
}

/*
插件入口函数PluginWrapper的逻辑：为了在dll与主程序之间安全传递数据，插件入口函数会接收一段主程序的内存空间与长度（使用reusableBytes分配），并
往其中写入数据；所有类型插件的PluginWrapper统一返回写入数据所需的长度（可能多可能少，但是少了就不写入），调用方（下面的那些函数）可通过返回的长度
信息重新调用一遍，从而成功写入。

每种类型的插件其PluginWrapper不同，有且只有一个，且PluginWrapper的实现涉及到底层内存操作，不能交给用户来做，因此才需要有
fgpk(github.com/nostalgist134/FuzzGIUpluginkit)这套工具链，使用fgpk编译一个插件时，仅需编写插件逻辑的实现函数，而PluginWrapper由fgpk
自动生成，无需自行实现。linux/macOS的插件同样有PluginWrapper中间层，不过它不需要涉及到内存管理，因此逻辑简单些

关于多重调用的逻辑：为了性能考量，目前插件的调用设计为最多只会重试一次（不然每次调用都失败，然后就一直再调用，这样下去程序都不要继续执行了），也就是
说当插件第一次调用返回的数据超出缓冲区大小，则会增大到第一次返回的大小1.5倍，然后再调用一次（第二次的大小可以相等可以更小），在此之后就不再调用或者
验证。

这里要搞这么多弯弯绕绕纯粹是因为go的原生plugin库不支持windows，因此只能采用cgo和build-mode=c-shared做中间层，傻逼go语言

原先还设计了另一套注册->调用框架，会把插件函数和参数都注册到registry中，但是我发现这种框架难以描述需要在单次调用过程中多次调用同一个插件的过程，因
此先放在那，目前参数还是调用时动态传入

插件一旦加载就无法卸载，这是因为虽然go提供了syscall.FreeLibrary，但是linux/macOS端使用go原生plugin实现，它们是没办法卸载动态库的，保持一致；
还有就是我发现在go中使用syscall.FreeLibrary必定导致io错乱，io系列包会变得完全不可用，一用就直接退出程序
*/

// PreLoad 预加载插件，并尝试获取插件的信息
func PreLoad(p fuzzTypes.Plugin, relPath string) (*convention.PluginInfo, error) {
	registryName := filepath.Join(relPath, p.Name)

	if loaded, ok := registry.Load(registryName); ok {
		return loaded.(*pluginRecord).pInfo, nil
	} else {
		mu.Lock()
		defer mu.Unlock()
		if loaded, ok = registry.Load(registryName); ok {
			return loaded.(*pluginRecord).pInfo, nil
		}
		pluginPath := filepath.Join(BaseDir, relPath, p.Name+binSuffix)

		dll, err := syscall.LoadDLL(pluginPath)
		if err != nil {
			return nil, err
		}
		proc, err := dll.FindProc(pluginEntry)
		if err != nil {
			return nil, err
		}

		// pluginInfo是可选的，因此加载失败也不返回错误
		pi, _ := fgpk.GetPluginInfo(pluginPath)
		registry.Store(registryName, &pluginRecord{dll: dll, proc: proc, pInfo: pi})
		return pi, nil
	}
}

// PayloadProcessor 返回处理后的字符串
func PayloadProcessor(p fuzzTypes.Plugin, outCtx *output.Ctx) string {
	rb, id := bp.Get()
	defer bp.Put(id)

	var needed int
	var err error
	if needed, err = callSharedLib(p, RelPathPlProc, rb); err != nil {
		callError(outCtx, RelPathPlProc, p, err)
		return ""
	} else if needed > rb.Cap() {
		rb.Resize(needed + needed>>1)
		needed, err = callSharedLib(p, RelPathPlProc, rb)
		if err != nil {
			callError(outCtx, RelPathPlProc, p, err)
			return ""
		}
	}

	rb.Resize(needed)

	return string(rb.Bytes())
}

// PayloadGenerator 返回插件生成的payload切片
func PayloadGenerator(p fuzzTypes.Plugin, outCtx *output.Ctx) []string {
	rb, id := bp.Get()
	rb.Resize(4096)
	defer bp.Put(id)

	if needed, err := callSharedLib(p, RelPathPlGen, rb); err != nil {
		callError(outCtx, RelPathPlGen, p, err)
		return []string{}
	} else if needed > rb.Cap() {
		rb.Resize(needed + needed>>1)
		_, err = callSharedLib(p, RelPathPlGen, rb)
		if err != nil {
			callError(outCtx, RelPathPlGen, p, err)
			return []string{}
		}
	}

	// 这个不需要强制重置游标，因为返回值已经包含了长度相关的信息了
	payloads := bytes2Strings(uintptr(rb.UnsafeBuffer()))
	return payloads
}

// Preprocess 返回指向preprocessor处理后新生成的Fuzz指针
func Preprocess(p fuzzTypes.Plugin, fuzz1 *fuzzTypes.Fuzz, outCtx *output.Ctx) *fuzzTypes.Fuzz {
	marshaled, err := json.Marshal(fuzz1)
	if err != nil {
		return fuzz1
	}

	rb, id := bp.Get()
	defer bp.Put(id)
	// 对于返回值为结构体的插件，由于需要存储json结构体，需要分配较大的空间，从而尽量避免可能的双重调用
	rb.Resize(4096)

	var needed int
	if needed, err = callSharedLib(p, RelPathPreprocessor, rb, marshaled); err != nil {
		callError(outCtx, RelPathPreprocessor, p, err)
		return fuzz1
	} else if needed == -1 {
		return fuzz1
	} else if needed > rb.Cap() { // 长度不够时，调用resize，而后再次调用动态库
		rb.Resize(needed + needed>>1)
		needed, err = callSharedLib(p, RelPathPreprocessor, rb, marshaled)
		if err != nil {
			callError(outCtx, RelPathPreprocessor, p, err)
			return fuzz1
		}
	}

	// 重置游标到needed，避免获取到无效数据
	rb.Resize(needed)

	newFuzz := new(fuzzTypes.Fuzz)
	err = json.Unmarshal(rb.Bytes(), newFuzz)
	if err != nil {
		return fuzz1
	}
	return newFuzz
}

// DoRequest 根据sendMeta发送请求，并接收响应
func DoRequest(p fuzzTypes.Plugin, reqCtx *fuzzTypes.RequestCtx) *fuzzTypes.Resp {
	reqJson, err := json.Marshal(reqCtx)
	if err != nil {
		return &fuzzTypes.Resp{ErrMsg: err.Error()}
	}

	rb, id := bp.Get()
	rb.Resize(4096)
	defer bp.Put(id)

	var needed int
	if needed, err = callSharedLib(p, RelPathReqSender, rb, reqJson); err != nil {
		return &fuzzTypes.Resp{ErrMsg: err.Error()}
	} else if needed == -1 {
		return &fuzzTypes.Resp{ErrMsg: errInteriorMarshal}
	} else if needed > rb.Cap() {
		rb.Resize(needed + needed>>1)
		needed, err = callSharedLib(p, RelPathReqSender, rb, reqJson)
		if err != nil {
			return &fuzzTypes.Resp{ErrMsg: err.Error()}
		}
	}

	rb.Resize(needed)

	resp := new(fuzzTypes.Resp)
	err = json.Unmarshal(rb.Bytes(), resp)
	if err != nil {
		return &fuzzTypes.Resp{ErrMsg: err.Error()}
	}

	return resp
}

// React 返回响应处理结果（reaction）
func React(p fuzzTypes.Plugin, req *fuzzTypes.Req, resp *fuzzTypes.Resp) *fuzzTypes.Reaction {
	rct := resourcePool.GetReaction()

	reqJson, err := json.Marshal(req)
	if err != nil {
		rct.Output.Msg = "error: " + err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
		return rct
	}

	respJson, err := json.Marshal(resp)
	if err != nil {
		rct.Output.Msg = "error: " + err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
		return rct
	}

	rb, id := bp.Get()
	rb.Resize(4096)
	defer bp.Put(id)

	var needed int
	if needed, err = callSharedLib(p, RelPathReactor, rb, reqJson, respJson); err != nil {
		rct.Output.Msg = err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
		return rct
	} else if needed == -1 {
		rct.Output.Msg = errInteriorMarshal
		rct.Flag |= fuzzTypes.ReactOutput
		return rct
	} else if needed > rb.Cap() {
		rb.Resize(needed + needed>>1)
		needed, err = callSharedLib(p, RelPathReactor, rb, reqJson, respJson)
		if err != nil {
			rct.Output.Msg = err.Error()
			rct.Flag |= fuzzTypes.ReactOutput
			return rct
		}
	}

	rb.Resize(needed)

	err = json.Unmarshal(rb.Bytes(), rct)
	if err != nil {
		rct.Output.Msg = "error: " + err.Error()
		rct.Flag |= fuzzTypes.ReactOutput
	}

	return rct
}

// IterIndex
// 接收一个lengths数组，代表每个fuzz关键字对应的payload列表的长度，返回一个与lengths同样大的数
// 组 （实际上通过写入out实现），标记每个关键字使用的payload的下标
// 迭代器调用前会通过下面的 IterLen 确定迭代长度，但是实现上这个函数是可选的，若选择不限迭代长度，也可
// 在迭代器中返回一个元素全为无效元素（全为负数）的数组来标记结束
// 实际上这个函数有2个隐藏的预定义参数，第一个是一个函数选择器，位于p.Args[0]，迭代器插件由于要在一个
// 插件中调用两个导出函数（IterIndex与IterLen），但是一个插件只有一个PluginWrapper，因此在
// PluginWrapper中实际上有一个选择器的逻辑，通过选择器选择IterIndex还是IterLen
// ind，位于p.Args[1]，表示当前的迭代下标，但是这个参数是调用者
// 传入（详见fuzz包中的iterator.go）
// out数组中可以有为负数，如果出现负数，则对应的关键字使用空payload
func IterIndex(p fuzzTypes.Plugin, lengths []int, out []int) {
	lengthsBytes, id := ints2Bytes(lengths)
	defer bp.Put(id)

	rb, id2 := bp.Get()
	defer bp.Put(id2)

	rb.Resize(len(lengthsBytes))

	needed, err := callSharedLib(p, RelPathIterator, rb, lengthsBytes)

	// iterator不使用“先探测，后写入”的逻辑，因为理论上来讲写入的长度是可预测的（长度即为len(lengthsBytes)），因此长度不符合直接报错
	if err != nil || needed > rb.Len() {
		for i, _ := range out {
			out[i] = -1
		}
	}

	bytes2Ints(uintptr(rb.UnsafeBuffer()), out)
	return
}

// IterLen 是迭代器的可选功能，返回迭代长度，若返回-1，则不限迭代长度
func IterLen(p fuzzTypes.Plugin, lengths []int) int {
	lengthsBytes, id := ints2Bytes(lengths)
	defer bp.Put(id)

	iterLen, err := callSharedLib(p, RelPathIterator, nil, lengthsBytes)
	if err != nil {
		iterLen = -1
	}
	return iterLen
}

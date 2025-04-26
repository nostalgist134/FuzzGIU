package plugin

import (
	"C"
	"FuzzGIU/components/fuzzTypes"
	"encoding/json"
	"strings"
	"syscall"
	"unsafe"
)

// callCShared 调用插件的PluginWrapper函数
// 调用约定 pluginFunction([jsonData, lenJsonData], 用户指定参数), json数据可选，如果为nil则不传入
func callCShared(plugin Plugin, relPath string, jsonData []byte, jsonData2 []byte) uintptr {
	dll, err := syscall.LoadDLL(pluginBase + relPath + plugin.Name + suffix)
	if err != nil {
		panic(err)
	}
	proc, err := dll.FindProc(pluginEntry)
	if err != nil {
		panic(err)
	}
	args := make([]uintptr, 0)
	if len(jsonData) > 0 {
		args = append(args, uintptr(unsafe.Pointer(&jsonData[0])))
		args = append(args, uintptr(len(jsonData)))
	}
	if len(jsonData2) > 0 {
		args = append(args, uintptr(unsafe.Pointer(&jsonData2[0])))
		args = append(args, uintptr(len(jsonData2)))
	}
	for _, arg := range plugin.Args {
		switch v := arg.(type) {
		case int:
			args = append(args, uintptr(v))
		case int8:
			args = append(args, uintptr(v))
		case int16:
			args = append(args, uintptr(v))
		case int32:
			args = append(args, uintptr(v))
		case int64:
			args = append(args, uintptr(v))
		case float32:
			args = append(args, uintptr(v))
		case float64:
			args = append(args, uintptr(v))
		case string:
			args = append(args, uintptr(unsafe.Pointer(&v)))
		}
	}
	r1, _, err := proc.Call(args...)
	if err != nil && err.Error() != "The operation completed successfully." {
		panic(err)
	}
	return r1
}

func parseJson(ptrJson uintptr) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(ptrJson+4)), *(*int32)(unsafe.Pointer(ptrJson)))
}

func bytes2strSlice(ptrBytes uintptr) []string {
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

// Call 根据类型调用插件函数，接收plugin类型字符串以及plugin结构体以及jsonData（用于preprocessor和reactor），
// 返回类型为any，在各个插件处理模块中转换
// payloadGenerator: 不使用json []string - 生成的payload切片
// preprocessor: json->*Fuzz类型 *Fuzz - 指向preprocessor处理后新生成的Fuzz结构的指针
// payloadProcessor: 不使用json string - 处理后的字符串
// reactor: json1->*Req json2->*Resp *reaction
// reqSender: json->*SendMeta *Resp
func Call(pluginType string, p Plugin, jsonData []byte, jsonData2 []byte) any {
	switch pluginType {
	case PTypePlGen: // payloadGenerator类型的plugin返回类型为[]string
		ptrPayload := callCShared(p, relPathPlGen, nil, nil)
		payload := bytes2strSlice(ptrPayload)
		return payload
	case PTypePreProc: // preprocessor类型的plugin返回*Fuzz
		fuzzJson := parseJson(callCShared(p, relPathPreprocessor, jsonData, nil))
		newFuzz := new(fuzzTypes.Fuzz)
		err := json.Unmarshal(fuzzJson, newFuzz)
		if err != nil {
			panic(err)
		}
		return newFuzz
	case PTypePlProc: // payloadProcessor类型的plugin返回string
		strBuilder := strings.Builder{}
		ptrPayloads := (*string)(unsafe.Pointer(callCShared(p, relPathPlProc, nil, nil)))
		strBuilder.WriteString(*ptrPayloads)
		return strBuilder.String() // 使用strings.builder将返回的string复制到主程序，避免gc回收问题
	case PTypeReactor: // reactor类型返回*reaction
		reactionJson := parseJson(callCShared(p, relPathReactor, jsonData, jsonData2))
		reaction := new(fuzzTypes.Reaction)
		err := json.Unmarshal(reactionJson, reaction)
		if err != nil {
			panic(err)
		}
		return reaction
	case PTypeReqSender: // requestSender类型返回*resp
		respJson := parseJson(callCShared(p, relPathReqSender, jsonData, nil))
		resp := new(fuzzTypes.Resp)
		err := json.Unmarshal(respJson, resp)
		if err != nil {
			panic(err)
		}
		return resp
	default:
		return nil
	}
}

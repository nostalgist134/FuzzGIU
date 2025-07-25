package plugin

import (
	"encoding/json"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
	"unsafe"
)

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

// Call 根据类型调用插件函数，接收plugin类型字符串以及plugin结构体以及jsonData（用于preprocessor和reactor），
// 返回类型为any，在各个插件处理模块中转换
// reqSender: json->*SendMeta *Resp
/*func Call(pluginType string, p Plugin, jsonData []byte, jsonData2 []byte) any {
	switch pluginType {
	case PTypePlGen:
		ptrPayload := callSharedLibOld(p, RelPathPlGen, nil, nil)
		payloads := bytes2Strings(ptrPayload)
		return payloads
	case PTypePreProc:
		fuzzJson := parseJson(callSharedLibOld(p, RelPathPreprocessor, jsonData, nil))
		newFuzz := new(fuzzTypes.Fuzz)
		err := json.Unmarshal(fuzzJson, newFuzz)
		if err != nil {
			panic(err)
		}
		return newFuzz
	case PTypePlProc:
		sb := strings.Builder{}
		ptrPayloads := (*string)(unsafe.Pointer(callSharedLibOld(p, RelPathPlProc, nil, nil)))
		sb.WriteString(*ptrPayloads)
		// 使用strings.builder将返回的string复制到主程序，避免gc回收问题
		ret := sb.String()
		return ret
	case PTypeReactor:
		reactionJson := parseJson(callSharedLibOld(p, RelPathReactor, jsonData, jsonData2))
		reaction := new(fuzzTypes.Reaction)
		err := json.Unmarshal(reactionJson, reaction)
		if err != nil {
			panic(err)
		}
		return reaction
	case PTypeReqSender: // requestSender类型返回*resp
		respJson := parseJson(callSharedLibOld(p, RelPathReqSender, jsonData, nil))
		resp := new(fuzzTypes.Resp)
		err := json.Unmarshal(respJson, resp)
		if err != nil {
			panic(err)
		}
		return resp
	default:
		return nil
	}
}*/

// PreProcessor 返回指向preprocessor处理后新生成的*Fuzz
func PreProcessor(p Plugin, fuzz1 *fuzzTypes.Fuzz) *fuzzTypes.Fuzz {
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
func PayloadGenerator(p Plugin) []string {
	ptrPayload := callSharedLib(p, RelPathPlGen)
	payloads := bytes2Strings(ptrPayload)
	return payloads
}

// PayloadProcessor 返回处理后的字符串
func PayloadProcessor(p Plugin) string {
	sb := strings.Builder{}
	ptrPayloads := (*string)(unsafe.Pointer(callSharedLib(p, RelPathPlProc)))
	sb.WriteString(*ptrPayloads)
	// 使用strings.builder将返回的string复制到主程序，避免gc回收问题
	ret := sb.String()
	return ret
}

// ReqSender 返回响应
func ReqSender(p Plugin, m *fuzzTypes.SendMeta) *fuzzTypes.Resp {
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

// Reactor 返回*reaction
func Reactor(p Plugin, req *fuzzTypes.Req, resp *fuzzTypes.Resp) *fuzzTypes.Reaction {
	reqJson, err := json.Marshal(req)
	if err != nil {
		panic(err)
	}
	respJson, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}
	reactionJson := parseJson(callSharedLib(p, RelPathReactor, reqJson, respJson))
	reaction := new(fuzzTypes.Reaction)
	err = json.Unmarshal(reactionJson, reaction)
	if err != nil {
		panic(err)
	}
	return reaction
}

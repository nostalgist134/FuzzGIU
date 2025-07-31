package plugin

import (
	"encoding/json"
	"github.com/nostalgist134/FuzzGIU/components/common"
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
	sb := strings.Builder{}
	ptrPayloads := (*string)(unsafe.Pointer(callSharedLib(p, RelPathPlProc)))
	sb.WriteString(*ptrPayloads)
	// 使用strings.builder将返回的string复制到主程序，避免gc回收问题
	ret := sb.String()
	return ret
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

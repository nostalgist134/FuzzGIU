package common

import (
	"FuzzGIU/components/fuzzTypes"
	"strings"
)

// 这个文件用于将req结构中的关键字替换为payload，所有的函数共享一个req结构池，当结构用完后，在react函数中会放回
/* 如果keyword遍历顺序有影响，一律按这个顺序：HttpSpec.Method -> URL -> HttpSpec.Version -> HttpSpec.Headers -> Data */

func replaceAndTrack(s, oldStr, newStr string, maxReplacements int) (string, []int) {
	var endIndices []int
	if oldStr == "" || maxReplacements == 0 {
		endIndices = append(endIndices, 0)
		return s, endIndices
	}
	var builder strings.Builder
	start := 0
	count := 0

	for {
		index := strings.Index(s[start:], oldStr)
		if index == -1 || (maxReplacements > 0 && count >= maxReplacements) {
			break
		}

		absoluteIndex := start + index
		builder.WriteString(s[start:absoluteIndex])    // 添加未匹配部分
		builder.WriteString(newStr)                    // 替换部分
		endIndices = append(endIndices, builder.Len()) // 记录替换后结尾索引
		start = absoluteIndex + len(oldStr)            // 更新起始索引
		count++
	}

	builder.WriteString(s[start:]) // 追加剩余部分
	if len(endIndices) > 0 {       // 最后一位转为负数，标记结束
		endIndices[len(endIndices)-1] *= -1
	} else { // patchLog#6: 没有匹配，返回-len(s)（原先是len(newStr)）
		endIndices = append(endIndices, -len(s))
	}
	return builder.String(), endIndices
}

// GetKeywordNum 获取一个关键字在req结构中出现的次数
func GetKeywordNum(req *fuzzTypes.Req, keyword string) int {
	num := strings.Count(req.HttpSpec.Method, keyword) + strings.Count(req.URL, keyword) +
		strings.Count(req.HttpSpec.Version, keyword) + strings.Count(req.Data, keyword)
	for _, header := range req.HttpSpec.Headers {
		num += strings.Count(header, keyword)
	}
	return num
}

// ReplacePayloadOld 将request中按照关键词列表替换单个payload，返回替换后的request
func ReplacePayloadOld(originalReq *fuzzTypes.Req, keywords []string, payloads []string) *fuzzTypes.Req {
	newReq := GetNewReq(originalReq)
	for i, keyword := range keywords {
		newReq.URL = strings.Replace(newReq.URL, keyword, payloads[i], -1)                         // 替换url中的关键字
		newReq.HttpSpec.Method = strings.Replace(newReq.HttpSpec.Method, keyword, payloads[i], -1) // 替换方法中的关键字

		for j := 0; j < len(newReq.HttpSpec.Headers); j++ { // 替换http头中的关键字
			newReq.HttpSpec.Headers[j] = strings.Replace(newReq.HttpSpec.Headers[j], keyword, payloads[i], -1)
		}
		newReq.Data = strings.Replace(newReq.Data, keyword, payloads[i], -1) // 替换data中的关键字
	}
	return newReq
}

// ReplacePayloadSniperOld 专用于sniper模式的替换函数
func ReplacePayloadSniperOld(originalReq *fuzzTypes.Req, keyword string, payload string, position int) *fuzzTypes.Req {
	if position < 0 || GetKeywordNum(originalReq, keyword) <= position {
		return originalReq
	}
	newReq := GetNewReq(originalReq)
	sb := strings.Builder{}
	splitter := getSplitter() // 其实需要使用getKeywordNum判断splitter是否出现在结构中，不过出错概率只有1/62^8，就不判断了
	// 将结构按照顺序入string builder，并以分隔符隔开
	sb.WriteString(newReq.HttpSpec.Method)
	sb.WriteString(splitter)
	sb.WriteString(newReq.URL)
	sb.WriteString(splitter)
	sb.WriteString(newReq.HttpSpec.Version)
	sb.WriteString(splitter)
	for _, header := range newReq.HttpSpec.Headers {
		sb.WriteString(header)
		sb.WriteString(splitter)
	}
	sb.WriteString(newReq.Data)
	buffer := sb.String()
	// 对buffer按照position进行替换
	if position > 0 {
		buffer = strings.Replace(buffer, keyword, "", position)
	}
	buffer = strings.Replace(buffer, keyword, payload, 1)
	buffer = strings.Replace(buffer, keyword, "", -1)
	// 替换完成后分隔，依次填回
	fields := strings.Split(buffer, splitter)
	newReq.HttpSpec.Method = fields[0]
	newReq.URL = fields[1]
	newReq.HttpSpec.Version = fields[2]
	i := 0
	for ; i < len(newReq.HttpSpec.Headers); i++ {
		newReq.HttpSpec.Headers[i] = fields[3+i]
	}
	newReq.Data = fields[3+i]
	return newReq
}

// ReplacePayloadTrack 将request中按照关键词替换payload，返回新的request以及替换后的结尾下标列表，用于递归模式
func ReplacePayloadTrack(originalReq *fuzzTypes.Req, keyword string, payload string) (*fuzzTypes.Req, []int) {
	newReq := GetNewReq(originalReq)
	var recursionPos []int
	var tmp []int
	// 按顺序替换关键字，并记录替换的尾部下标，在递归时使用
	// Method
	newReq.HttpSpec.Method, recursionPos = replaceAndTrack(newReq.HttpSpec.Method, keyword, payload, -1)
	// URL
	newReq.URL, tmp = replaceAndTrack(newReq.URL, keyword, payload, -1)
	recursionPos = append(recursionPos, tmp...)
	// HttpVersion
	newReq.HttpSpec.Version, tmp = replaceAndTrack(newReq.HttpSpec.Version, keyword, payload, -1)
	recursionPos = append(recursionPos, tmp...)
	// HttpHeaders
	for i := 0; i < len(newReq.HttpSpec.Headers); i++ {
		newReq.HttpSpec.Headers[i], tmp = replaceAndTrack(newReq.HttpSpec.Headers[i], keyword, payload, -1)
		recursionPos = append(recursionPos, tmp...)
	}
	// Data
	newReq.Data, tmp = replaceAndTrack(newReq.Data, keyword, payload, -1)
	recursionPos = append(recursionPos, tmp...)

	return newReq, recursionPos
}

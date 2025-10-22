package tmplReplace

import (
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
)

// req2Str 将req结构转为字符串表示，注意：即使是全为空的结构，也至少会返回一个包含3个分隔符的串
func req2Str(req *fuzzTypes.Req) (stringified string, splitter string) {
	splitter = common.RandMarker()
	sb := strings.Builder{}

	sb.WriteString(req.HttpSpec.Method)
	sb.WriteString(splitter)

	sb.WriteString(req.URL)
	sb.WriteString(splitter)

	sb.WriteString(req.HttpSpec.Version)
	sb.WriteString(splitter)

	for _, header := range req.HttpSpec.Headers {
		sb.WriteString(header)
		sb.WriteString(splitter)
	}

	for _, field := range req.Fields {
		sb.WriteString(field.Name)
		sb.WriteString(splitter)
		sb.WriteString(field.Value)
		sb.WriteString(splitter)
	}

	sb.Write(req.Data)

	stringified = sb.String()
	return
}

// countFields 计算field数，field数恒等于phSplitter数+1
func (t *ReplaceTemplate) countFields() int {
	t.fieldNum = 1

	i := 0
	for ; i < len(t.placeholders); i++ {
		if t.placeholders[i] == phSplitter {
			t.fieldNum++
		}
	}

	return t.fieldNum
}

// parse 将字符串化的请求解析并转化为ReplaceTemplate结构
// fragments与placeholders对应关系如下
// fr[0]->ph[0]->fr[1]->ph[1]->...->fr[n]->ph[n]->fr[n+1]
// 也就是说fragments的长度总是比placeholders的长度多1个
func (t *ReplaceTemplate) parse(s string, splitterAndKeywords []string, headerNum int) {
	t.fragments = make([]string, 0)
	keywordsOccur := getKeywordsOccurrences(s, splitterAndKeywords)
	mergedOccur := mergeK(keywordsOccur)
	i := 0
	sInd := 0 // 字符串的下标

	/*
		if len(keywordsOccur) == 0 {
			...
		}
		正常情况下，只要从包外部使用，是不可能出现keywordsOccur长度等于0的情况，因为将Req转为字符串的过程一定会写入
		splitter，因此此处不做判断
	*/

	// 关键字出现在字符串开头的特殊情况
	if mergedOccur[0].KeywordInReq == 0 {
		t.fragments = append(t.fragments, "") // 此时fragments[0]使用空字符串
		t.placeholders = append(t.placeholders, mergedOccur[0].KeywordInSlice)
		i++
		// 及时将sInd更新，避免关键字出现在开头时模板解析异常的问题
		sInd = len(splitterAndKeywords[mergedOccur[0].KeywordInSlice])
	}

	for ; i < len(mergedOccur); i++ {
		t.fragments = append(t.fragments, s[sInd:mergedOccur[i].KeywordInReq])
		// placeholders数组标记了在关键字对应的下标出现的是哪一个关键字
		t.placeholders = append(t.placeholders, mergedOccur[i].KeywordInSlice)
		sInd = mergedOccur[i].KeywordInReq + len(splitterAndKeywords[mergedOccur[i].KeywordInSlice])
	}

	if sInd < len(s) {
		t.fragments = append(t.fragments, s[sInd:])
	} else {
		// 特殊情况，结尾是关键字，添加一个空fragment，对开头和结尾特殊情况的处理能
		// 使得len(t.fragments)恒等于len(t.placeholders+1)，从而避免越界问题
		t.fragments = append(t.fragments, "")
	}
	t.headerNum = headerNum
	t.countFields()
}

// ParseReqTmpl 解析req模板
func ParseReqTmpl(req *fuzzTypes.Req, keywords []string) *ReplaceTemplate {
	s, splitter := req2Str(req)
	splitterAndKeywords := append([]string{splitter}, keywords...)
	tmpl := new(ReplaceTemplate)
	tmpl.parse(s, splitterAndKeywords, len(req.HttpSpec.Headers))
	return tmpl
}

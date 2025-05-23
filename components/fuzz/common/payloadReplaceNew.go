package common

import (
	"FuzzGIU/components/fuzzTypes"
	"container/heap"
	"math/rand"
	"strings"
)

func findAllIndices(s, substr string) []int {
	var indices []int
	if substr == "" {
		return indices // 空子串不处理
	}

	index := strings.Index(s, substr)
	for index != -1 {
		indices = append(indices, index)
		index = strings.Index(s[index+1:], substr)
		if index != -1 {
			index += indices[len(indices)-1] + 1 // 修正相对位置为原始字符串下标
		}
	}
	return indices
}

func getKeywordsOccurrences(s string, keywords []string) [][]int {
	ret := make([][]int, 0)
	for _, keyword := range keywords {
		ret = append(ret, findAllIndices(s, keyword))
	}
	return ret
}

// Node 表示堆中的元素
type Node struct {
	value    int // 数值大小
	arrayIdx int // 所在数组编号
	idx      int // 在该数组中的索引
}

// PriorityQueue 实现 heap.Interface，并按 value 从小到大排序
type PriorityQueue []Node

func (pq *PriorityQueue) Len() int { return len(*pq) }

func (pq *PriorityQueue) Less(i, j int) bool {
	return (*pq)[i].value < (*pq)[j].value // 小顶堆
}

func (pq *PriorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
}
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(Node))
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	*pq = old[0 : n-1]
	return node
}

// record 结果结构体，包含数值和其来源数组下标
type record struct {
	KeywordIndex int // 关键字在原字符串中的下标
	ArrayIdx     int // 关键字在occur二维数组中的下标
}

// mergeK 返回排序值 + 所属数组下标，all credits to ChatGPT
func mergeK(arrays [][]int) []record {
	pq := &PriorityQueue{}
	heap.Init(pq)
	// 初始化堆，把每个数组的第一个元素加入堆
	for i, arr := range arrays {
		if len(arr) > 0 {
			heap.Push(pq, Node{
				value:    arr[0],
				arrayIdx: i,
				idx:      0,
			})
		}
	}

	var result []record
	for pq.Len() > 0 {
		node := heap.Pop(pq).(Node)
		result = append(result, record{
			KeywordIndex: node.value,
			ArrayIdx:     node.arrayIdx,
		})

		// 如果该数组还有下一个元素，继续压入堆
		nextIdx := node.idx + 1
		if nextIdx < len(arrays[node.arrayIdx]) {
			heap.Push(pq, Node{
				value:    arrays[node.arrayIdx][nextIdx],
				arrayIdx: node.arrayIdx,
				idx:      nextIdx,
			})
		}
	}
	return result
}

type ReplaceTemplate struct {
	fragments    []string
	placeholders []int
}

func (t *ReplaceTemplate) parse(s string, keywords []string) {
	t.fragments = make([]string, 0)
	keywordsOccur := getKeywordsOccurrences(s, keywords)
	mergedOccur := mergeK(keywordsOccur)
	i := 0
	if mergedOccur[0].KeywordIndex == 0 { // 特殊情况：关键字出现在字符串开头，使用空字符串
		t.fragments = append(t.fragments, "")
		t.placeholders = append(t.placeholders, mergedOccur[0].ArrayIdx)
		i++
	}
	sInd := 0
	for ; i < len(mergedOccur); i++ {
		t.fragments = append(t.fragments, s[sInd:mergedOccur[i].KeywordIndex])
		// placeholders数组标记了在关键字对应的下标出现的是哪一个关键字
		t.placeholders = append(t.placeholders, mergedOccur[i].ArrayIdx)
		sInd = mergedOccur[i].KeywordIndex + len(keywords[mergedOccur[i].ArrayIdx])
	}
	if sInd < len(s) {
		t.fragments = append(t.fragments, s[sInd:])
	} else {
		// 特殊情况，结尾是关键字，添加一个空字符串，有了开头特殊和结尾特殊情况的处理，就能稳定保持#t.fragments=#t.placeholders+1
		// 从而避免越界问题
		t.fragments = append(t.fragments, "")
	}
}

// renderNew 约定placeholder值为0为分隔符，返回切片字符串
func (t *ReplaceTemplate) renderNew(payloads []string) []string {
	sb := strings.Builder{}
	fields := make([]string, 0)
	i := 0
	for ; i < len(t.placeholders); i++ {
		sb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			fields = append(fields, sb.String())
			sb.Reset()
			continue
		}
		sb.WriteString(payloads[t.placeholders[i]-1])
	}
	sb.WriteString(t.fragments[i])
	fields = append(fields, sb.String())
	return fields
}

func (t *ReplaceTemplate) render1New(payload string, pos int) []string {
	// 不知道怎么回事，这里的下标是从1开始算的
	if pos < 0 || pos > len(t.placeholders) {
		payload = ""
	}
	fields := make([]string, 0)
	sb := strings.Builder{}
	i := 0
	j := 0
	for ; j < pos && i < len(t.placeholders); j++ {
		sb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			j--
			fields = append(fields, sb.String())
			sb.Reset()
		}
		i++
	}
	sb.WriteString(payload)
	for ; i < len(t.placeholders); i++ {
		sb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			fields = append(fields, sb.String())
			sb.Reset()
			continue
		}
	}
	sb.WriteString(t.fragments[i])
	fields = append(fields, sb.String())
	return fields
}

// getSplitter 使用的分隔符
func getSplitter() string {
	dict := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	splitter := strings.Builder{}
	for i := 0; i < 12; i++ {
		splitter.WriteByte(dict[rand.Intn(len(dict))])
	}
	return splitter.String()
}

func req2Str(req *fuzzTypes.Req) (string, string) {
	splitter := getSplitter()
	sb := strings.Builder{}
	// 将结构按照顺序入string builder，并以分隔符隔开
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
	sb.WriteString(req.Data)
	return sb.String(), splitter
}

func ParseReqTemplate(req *fuzzTypes.Req, keywords []string) *ReplaceTemplate {
	s, splitter := req2Str(req)
	keywords = append([]string{splitter}, keywords...)
	replaceTemp := new(ReplaceTemplate)
	replaceTemp.parse(s, keywords)
	return replaceTemp
}

func ReplacePayloadsByTemplate(t *ReplaceTemplate, payloads []string, sniperPos int) *fuzzTypes.Req {
	var fields []string
	if sniperPos >= 0 {
		fields = t.render1New(payloads[0], sniperPos+1)
	} else {
		fields = t.renderNew(payloads)
	}
	newReq := GetNewReq(nil)
	newReq.HttpSpec.Method = fields[0]
	newReq.URL = fields[1]
	newReq.HttpSpec.Version = fields[2]
	i := 0
	newReq.HttpSpec.Headers = make([]string, 0)
	for ; i < len(fields)-4; i++ {
		newReq.HttpSpec.Headers = append(newReq.HttpSpec.Headers, fields[3+i])
	}
	newReq.Data = fields[3+i]
	return newReq
}

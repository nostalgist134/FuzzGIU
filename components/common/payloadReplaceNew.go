package common

import (
	"container/heap"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"math/rand"
	"strings"
)

// GetKeywordNum 获取一个关键字在req结构中出现的次数
func GetKeywordNum(req *fuzzTypes.Req, keyword string) int {
	num := strings.Count(req.HttpSpec.Method, keyword) + strings.Count(req.URL, keyword) +
		strings.Count(req.HttpSpec.Version, keyword) + strings.Count(req.Data, keyword)
	for _, header := range req.HttpSpec.Headers {
		num += strings.Count(header, keyword)
	}
	return num
}

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
	placeholders []int // placeholders 存储每个片段后关键字在关键字列表的下标列表，特殊情况：下标值为0，代表分隔符
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
		// 特殊情况，结尾是关键字，添加一个空fragment，对开头和结尾特殊情况的处理能
		// 使得len(t.fragments)恒等于len(t.placeholders+1)，从而避免越界问题
		t.fragments = append(t.fragments, "")
	}
}

// renderNew 对模板进行渲染，返回通过分隔符分隔的fields切片
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

// render2 用于替代原先的replacePayloadAndTrack函数
func (t *ReplaceTemplate) render2(payload string) ([]string, []int) {
	fields := make([]string, 0)
	trackPos := make([]int, 0)
	sb := strings.Builder{}
	i := 0
	trackPosInd := -1
	fieldHasPayload := false
	var tmp string
	for ; i < len(t.placeholders); i++ {
		sb.WriteString(t.fragments[i])
		// 分隔符
		if t.placeholders[i] == 0 {
			tmp = sb.String()
			fields = append(fields, tmp)
			sb.Reset()
			if !fieldHasPayload {
				trackPos = append(trackPos, -(len(tmp) + 1))
				trackPosInd++
			} else {
				trackPos[trackPosInd] *= -1
				fieldHasPayload = false
			}
		} else {
			sb.WriteString(payload)
			tmp = sb.String()
			trackPos = append(trackPos, len(tmp))
			fieldHasPayload = true
			trackPosInd++
		}
	}
	sb.WriteString(t.fragments[i])
	fields = append(fields, sb.String())
	if trackPos[trackPosInd] > 0 {
		trackPos[trackPosInd] *= -1
	}
	if t.placeholders[i-1] == 0 {
		trackPos = append(trackPos, -(len(fields[len(fields)-1]) + 1))
	}
	return fields, trackPos
}

func (t *ReplaceTemplate) render3(payload string, pos int) ([]string, []int) {
	if pos < 0 || pos > len(t.placeholders) {
		payload = ""
	}
	var field string
	fields := make([]string, 0)
	trackPos := make([]int, 0)
	sb := strings.Builder{}
	i := 0
	j := 0
	for ; j < pos && i < len(t.placeholders); j++ {
		sb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			j--
			field = sb.String()
			fields = append(fields, field)
			trackPos = append(trackPos, -(len(field) + 1))
			sb.Reset()
		}
		i++
	}
	sb.WriteString(payload)
	field = sb.String()
	trackPos = append(trackPos, -(len(field)))
	sniperFieldEnd := false
	for ; i < len(t.placeholders); i++ {
		sb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			field = sb.String()
			fields = append(fields, field)
			if sniperFieldEnd {
				trackPos = append(trackPos, -(len(field) + 1))
			} else {
				sniperFieldEnd = true
			}
			sb.Reset()
			continue
		}
	}
	sb.WriteString(t.fragments[i])
	field = sb.String()
	fields = append(fields, field)
	if sniperFieldEnd {
		trackPos = append(trackPos, -(len(field) + 1))
	}
	return fields, trackPos
}

// GetRandMarker 生成一个长度为12为的随机字符串
func GetRandMarker() string {
	dict := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	sb := strings.Builder{}
	for i := 0; i < 12; i++ {
		sb.WriteByte(dict[rand.Intn(len(dict))])
	}
	return sb.String()
}

func req2Str(req *fuzzTypes.Req) (string, string) {
	splitter := GetRandMarker()
	sb := strings.Builder{}
	// 将结构按照顺序入string builder，并以分隔符隔开，splitter本来应该是要校验的，但是为了运行速度就不做校验了
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
	newReq := GetNewReq()
	newReq.HttpSpec.Method = fields[0]
	newReq.URL = fields[1]
	newReq.HttpSpec.Version = fields[2]
	i := 0
	// GetNewReq获取的Req结构http头可能是已经分配好的，可以复用
	for ; i < len(fields)-4 && i < len(newReq.HttpSpec.Headers); i++ {
		newReq.HttpSpec.Headers[i] = fields[3+i]
	}
	// 预分配的头如果为nil则新建
	if newReq.HttpSpec.Headers == nil {
		newReq.HttpSpec.Headers = make([]string, 0)
	}
	// 原有的字符串切片不够用了就append
	for ; i < len(fields)-4; i++ {
		newReq.HttpSpec.Headers = append(newReq.HttpSpec.Headers, fields[3+i])
	}
	// 若预分配的头长于模板解析出的头，则截断
	if len(newReq.HttpSpec.Headers) > len(fields)-4 {
		newReq.HttpSpec.Headers = newReq.HttpSpec.Headers[:len(fields)-4]
	}
	newReq.Data = fields[3+i]
	return newReq
}

func ReplacePayloadTrackTemplate(t *ReplaceTemplate, payload string, sniperPos int) (*fuzzTypes.Req, []int) {
	var fields []string
	var track []int

	if sniperPos > 0 {
		fields, track = t.render3(payload, sniperPos+1)
	} else {
		fields, track = t.render2(payload)
	}
	newReq := GetNewReq()
	newReq.HttpSpec.Method = fields[0]
	newReq.URL = fields[1]
	newReq.HttpSpec.Version = fields[2]
	i := 0
	// GetNewReq获取的Req结构http头可能是已经分配好的，可以复用
	for ; i < len(fields)-4 && i < len(newReq.HttpSpec.Headers); i++ {
		newReq.HttpSpec.Headers[i] = fields[3+i]
	}
	// 预分配的头如果为nil则新建
	if newReq.HttpSpec.Headers == nil {
		newReq.HttpSpec.Headers = make([]string, 0)
	}
	// 原有的字符串切片不够用了就append
	for ; i < len(fields)-4; i++ {
		newReq.HttpSpec.Headers = append(newReq.HttpSpec.Headers, fields[3+i])
	}
	// 若预分配的头长于模板解析出的头，则截断
	if len(newReq.HttpSpec.Headers) > len(fields)-4 {
		newReq.HttpSpec.Headers = newReq.HttpSpec.Headers[:len(fields)-4]
	}
	newReq.Data = fields[3+i]
	return newReq, track
}

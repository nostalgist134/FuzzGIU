package common

import (
	"container/heap"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	reusablebytes "github.com/nostalgist134/reusableBytes"
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

// heapNode 表示堆中的元素
type heapNode struct {
	value    int // 数值大小
	arrayIdx int // 所在数组编号
	idx      int // 在该数组中的索引
}

// priorityQueue 实现 heap.Interface，并按 value 从小到大排序
type priorityQueue []heapNode

func (pq *priorityQueue) Len() int { return len(*pq) }

func (pq *priorityQueue) Less(i, j int) bool {
	return (*pq)[i].value < (*pq)[j].value // 小顶堆
}

func (pq *priorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
}
func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(heapNode))
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	node := old[n-1]
	*pq = old[0 : n-1]
	return node
}

// record 结果结构体，包含数值和其来源数组下标
type record struct {
	KeywordInReq   int // 关键字在Req字符串中的下标
	KeywordInSlice int // 关键字在传入的关键字切片中的下标
}

// mergeK 返回排序值 + 所属数组下标，all credits to ChatGPT
func mergeK(arrays [][]int) []record {
	pq := &priorityQueue{}
	heap.Init(pq)
	// 初始化堆，把每个数组的第一个元素加入堆
	for i, arr := range arrays {
		if len(arr) > 0 {
			heap.Push(pq, heapNode{
				value:    arr[0],
				arrayIdx: i,
				idx:      0,
			})
		}
	}

	var result []record
	for pq.Len() > 0 {
		node := heap.Pop(pq).(heapNode)
		result = append(result, record{
			KeywordInReq:   node.value,
			KeywordInSlice: node.arrayIdx,
		})

		// 如果该数组还有下一个元素，继续压入堆
		nextIdx := node.idx + 1
		if nextIdx < len(arrays[node.arrayIdx]) {
			heap.Push(pq, heapNode{
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
	fieldNum     int
}

var bp = new(reusablebytes.BytesPool)

func init() {
	bp.Init(128, 131072, 128)
}

func (t *ReplaceTemplate) getFieldNum() int {
	if t.fieldNum < 2 {
		t.fieldNum = 0
		for _, ph := range t.placeholders {
			if ph == 0 {
				t.fieldNum++
			}
		}
		t.fieldNum++
	}
	return t.fieldNum
}

func (t *ReplaceTemplate) parse(s string, keywords []string) {
	t.fragments = make([]string, 0)
	keywordsOccur := getKeywordsOccurrences(s, keywords)
	mergedOccur := mergeK(keywordsOccur)
	i := 0
	sInd := 0                             // req字符串的下标
	if mergedOccur[0].KeywordInReq == 0 { // 特殊情况：关键字出现在字符串开头，使用空字符串
		t.fragments = append(t.fragments, "")
		t.placeholders = append(t.placeholders, mergedOccur[0].KeywordInSlice)
		i++
		// 及时将sInd更新，修复了关键字出现在开头时模板解析异常的问题
		sInd = len(keywords[mergedOccur[0].KeywordInSlice])
	}
	for ; i < len(mergedOccur); i++ {
		t.fragments = append(t.fragments, s[sInd:mergedOccur[i].KeywordInReq])
		// placeholders数组标记了在关键字对应的下标出现的是哪一个关键字
		t.placeholders = append(t.placeholders, mergedOccur[i].KeywordInSlice)
		sInd = mergedOccur[i].KeywordInReq + len(keywords[mergedOccur[i].KeywordInSlice])
	}
	if sInd < len(s) {
		t.fragments = append(t.fragments, s[sInd:])
	} else {
		// 特殊情况，结尾是关键字，添加一个空fragment，对开头和结尾特殊情况的处理能
		// 使得len(t.fragments)恒等于len(t.placeholders+1)，从而避免越界问题
		t.fragments = append(t.fragments, "")
	}
	t.getFieldNum()
}

// renderNew 对模板进行渲染，返回通过分隔符分隔的fields切片
func (t *ReplaceTemplate) renderNew(payloads []string) ([]string, int32) {
	rb, id := bp.Get()
	fields := make([]string, t.fieldNum)
	i := 0
	indField := 0
	rb.Anchor()
	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			fields[indField] = rb.StringFromAnchor()
			rb.Anchor()
			indField++
			continue
		}
		rb.WriteString(payloads[t.placeholders[i]-1])
	}
	rb.WriteString(t.fragments[i])
	fields[indField] = rb.StringFromAnchor()
	return fields, id
}

// render1New 用于sniper模式的渲染函数
func (t *ReplaceTemplate) render1New(payload string, pos int) ([]string, int32) {
	if pos < 0 || pos > len(t.placeholders) {
		payload = ""
	}
	fields := make([]string, t.fieldNum)
	rb, id := bp.Get()
	i := 0
	j := 0
	fieldInd := 0
	rb.Anchor()
	for ; j <= pos && i < len(t.placeholders); j++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			j--
			fields[fieldInd] = rb.StringFromAnchor()
			fieldInd++
			rb.Anchor()
		}
		i++
	}
	rb.WriteString(payload)
	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			fields[fieldInd] = rb.StringFromAnchor()
			fieldInd++
			rb.Anchor()
		}
	}
	rb.WriteString(t.fragments[i])
	fields[fieldInd] = rb.StringFromAnchor()
	return fields, id
}

// render2 用于替代原先的replacePayloadAndTrack函数
func (t *ReplaceTemplate) render2(payload string) ([]string, []int, int32) {
	fields := make([]string, t.fieldNum)
	trackPos := make([]int, 0)
	rb, id := bp.Get()
	i := 0
	trackPosInd := -1
	fieldHasPayload := false
	fieldInd := 0
	var tmp string
	rb.Anchor()
	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		// 分隔符
		if t.placeholders[i] == 0 {
			tmp = rb.StringFromAnchor()
			fields[fieldInd] = tmp
			rb.Anchor()
			if !fieldHasPayload {
				trackPos = append(trackPos, -(len(tmp) + 1))
				trackPosInd++
			} else {
				trackPos[trackPosInd] *= -1
				fieldHasPayload = false
			}
			fieldInd++
		} else {
			rb.WriteString(payload)
			tmp = rb.StringFromAnchor()
			trackPos = append(trackPos, len(tmp))
			fieldHasPayload = true
			trackPosInd++
		}
	}
	rb.WriteString(t.fragments[i])
	fields[fieldInd] = rb.StringFromAnchor()
	if trackPos[trackPosInd] > 0 {
		trackPos[trackPosInd] *= -1
	}
	if t.placeholders[i-1] == 0 {
		trackPos = append(trackPos, -(len(fields[len(fields)-1]) + 1))
	}
	return fields, trackPos, id
}

func (t *ReplaceTemplate) render3(payload string, pos int) ([]string, []int, int32) {
	if pos < 0 || pos > len(t.placeholders) {
		payload = ""
	}
	var field string
	fields := make([]string, t.fieldNum)
	trackPos := make([]int, 0)
	rb, id := bp.Get()
	i := 0
	j := 0
	fieldInd := 0
	sniperFieldEnd := false
	rb.Anchor()
	for ; j <= pos && i < len(t.placeholders); j++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			j--
			field = rb.StringFromAnchor()
			fields[fieldInd] = field
			trackPos = append(trackPos, -(len(field) + 1))
			rb.Anchor()
			fieldInd++
		}
		i++
	}
	rb.WriteString(payload)
	field = rb.StringFromAnchor()
	trackPos = append(trackPos, -(len(field)))
	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == 0 {
			field = rb.StringFromAnchor()
			fields[fieldInd] = field
			if sniperFieldEnd {
				trackPos = append(trackPos, -(len(field) + 1))
			} else {
				sniperFieldEnd = true
			}
			rb.Anchor()
			fieldInd++
		}
	}
	rb.WriteString(t.fragments[i])
	field = rb.StringFromAnchor()
	fields[fieldInd] = field
	if sniperFieldEnd {
		trackPos = append(trackPos, -(len(field) + 1))
	}
	return fields, trackPos, id
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

func ReplacePayloadsByTemplate(t *ReplaceTemplate, payloads []string, sniperPos int) (*fuzzTypes.Req, int32) {
	var fields []string
	var cacheId int32
	if sniperPos >= 0 {
		fields, cacheId = t.render1New(payloads[0], sniperPos)
	} else {
		fields, cacheId = t.renderNew(payloads)
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
	return newReq, cacheId
}

func ReplacePayloadTrackTemplate(t *ReplaceTemplate, payload string, sniperPos int) (*fuzzTypes.Req, []int, int32) {
	var fields []string
	var track []int
	var id int32

	if sniperPos >= 0 {
		fields, track, id = t.render3(payload, sniperPos)
	} else {
		fields, track, id = t.render2(payload)
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
	return newReq, track, id
}

func ReleaseReqCache(id int32) {
	bp.Put(id)
}

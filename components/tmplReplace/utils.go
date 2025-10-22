package tmplReplace

import (
	"container/heap"
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

package tmplReplace

import (
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	reusablebytes "github.com/nostalgist134/reusableBytes"
)

// todo: 为这个包添加一些注释

var lazyPool = resourcePool.SlicePool[reusablebytes.Lazy]{}

// render 对模板进行渲染，返回通过分隔符分隔的fields切片
func (t *ReplaceTemplate) render(payloads []string) ([]reusablebytes.Lazy, int32) {
	rb, id := bp.Get()
	lazyFields := lazyPool.Get(t.fieldNum)
	i := 0
	indField := 0

	rb.Anchor()
	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == phSplitter {
			lazyFields[indField] = rb.LazyFromAnchor()
			rb.Anchor()
			indField++
			continue
		}
		rb.WriteString(payloads[t.placeholders[i]-1])
	}
	rb.WriteString(t.fragments[i])
	lazyFields[indField] = rb.LazyFromAnchor()
	return lazyFields, id
}

// renderSniper 用于sniper模式的渲染函数
func (t *ReplaceTemplate) renderSniper(payload string, pos int) ([]reusablebytes.Lazy, int32) {
	if pos < 0 || pos > len(t.placeholders) {
		payload = ""
	}
	lazyFields := lazyPool.Get(t.fieldNum)
	rb, id := bp.Get()
	i := 0
	j := 0
	fieldInd := 0

	rb.Anchor()
	for ; j <= pos && i < len(t.placeholders); j++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == phSplitter {
			j--
			lazyFields[fieldInd] = rb.LazyFromAnchor()
			fieldInd++
			rb.Anchor()
		}
		i++
	}

	rb.WriteString(payload)

	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == phSplitter {
			lazyFields[fieldInd] = rb.LazyFromAnchor()
			fieldInd++
			rb.Anchor()
		}
	}
	rb.WriteString(t.fragments[i])
	lazyFields[fieldInd] = rb.LazyFromAnchor()
	return lazyFields, id
}

// renderTrack 他妈的之前写嗨了，现在自己都看不懂，等有时间了再读一下然后写注释吧
func (t *ReplaceTemplate) renderTrack(payload string) ([]reusablebytes.Lazy, []int, int32) {
	lazyFields := lazyPool.Get(t.fieldNum)
	trackPos := make([]int, 0)
	rb, id := bp.Get()
	i := 0
	trackPosInd := -1
	fieldHasPayload := false
	fieldInd := 0
	var tmp reusablebytes.Lazy

	rb.Anchor()
	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == phSplitter { // 分隔符
			tmp = rb.LazyFromAnchor()
			lazyFields[fieldInd] = tmp
			rb.Anchor()

			if !fieldHasPayload {
				trackPos = append(trackPos, -(tmp.Len() + 1))
				trackPosInd++
			} else {
				trackPos[trackPosInd] *= -1
				fieldHasPayload = false
			}
			fieldInd++
		} else {
			rb.WriteString(payload)
			tmp = rb.LazyFromAnchor()
			trackPos = append(trackPos, tmp.Len())
			fieldHasPayload = true
			trackPosInd++
		}
	}
	rb.WriteString(t.fragments[i])
	lazyFields[fieldInd] = rb.LazyFromAnchor()
	if trackPos[trackPosInd] > 0 {
		trackPos[trackPosInd] *= -1
	}
	if t.placeholders[i-1] == phSplitter {
		trackPos = append(trackPos, -(lazyFields[len(lazyFields)-1].Len() + 1))
	}
	return lazyFields, trackPos, id
}

// renderTrackSniper 同时带有跟踪功能和sniper下标功能，现在写了一些注释，应该会好理解些
func (t *ReplaceTemplate) renderTrackSniper(payload string, pos int) ([]reusablebytes.Lazy, []int, int32) {
	if pos < 0 || pos > len(t.placeholders) {
		payload = ""
	}
	var lazyField reusablebytes.Lazy
	lazyFields := lazyPool.Get(t.fieldNum)
	trackPos := resourcePool.IntSlices.Get(0)
	rb, id := bp.Get()
	i := 0
	j := 0
	fieldInd := 0

	rb.Anchor()
	// 将sniper下标前的所有关键字占位符替换为空
	for ; j <= pos && i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == phSplitter { // 遇到分隔符，单个字段结束
			lazyField = rb.LazyFromAnchor()
			lazyFields[fieldInd] = lazyField
			trackPos = append(trackPos, -(lazyField.Len() + 1))
			rb.Anchor() // 重置锚点为当前缓冲区的尾部，也就是下一个字段的起始点）
			fieldInd++
		} else { // 若不为分隔符，则sniper下标增加
			j++
		}
	}

	// 上面的循环如果结束，说明：
	// 	1.placeholders走完了还没到sniper下标，这种情况通常不可能发生，因为用户实际上并不能控制sniper下标，是程序自动控制
	//	的，不过就算发生了也能处理，将sniper下标视为“最后出现的位置”，也就是直接在最后写入payload就行了
	//	2.到达sniper下标，在下标处写入payload
	rb.WriteString(payload)
	lazyField = rb.LazyFromAnchor()
	// 注意：由于在sniper模式下，一个请求模板中必然只有一个关键字被替换为payload，而一个关键字又不可能同时存在两个字段中，因此
	// 这里trackPos设置为负的lazyField的长度，这是因为“绝对值不超过当前字段的长度的负数“代表“当前字段有插入一个payload，且
	// 是最后一个”
	trackPos = append(trackPos, -(lazyField.Len()))

	// 标记了包含sniper下标的字段是否已经结束
	sniperFieldEnd := false
	for ; i < len(t.placeholders); i++ {
		rb.WriteString(t.fragments[i])
		if t.placeholders[i] == phSplitter {
			lazyField = rb.LazyFromAnchor()
			lazyFields[fieldInd] = lazyField
			// 若sniper下标字段已经结束，之后就一律填入“-每个字段的长度+1”，表明这些字段中并不包含payload插入点
			if sniperFieldEnd {
				trackPos = append(trackPos, -(lazyField.Len() + 1))
			} else { // 若遇到分隔符，则代表含有sniper下标的字段已经结束了
				sniperFieldEnd = true
			}
			rb.Anchor()
			fieldInd++
		}
	}
	rb.WriteString(t.fragments[i]) // 写入最后一个fragment（别忘了fragment数总是比placeholder数多1）
	lazyField = rb.LazyFromAnchor()
	lazyFields[fieldInd] = lazyField

	// 处理最后一个字段，有两种可能的情况：
	//	1.sniperFieldEnd为假，这说明含有sniper下标的字段是最后一个字段，这里就不需要append，因为这个字段对应的
	//	跟踪下标已经在上面两个循环之间的代码中处理过了
	// 	2.sniperFieldEnd为真，说明含有sniper下标的字段不是最后一个字段，由于fragment数总是比placeholder数多
	//	1，因此这里需要手动append最后一个字段的跟踪下标
	if sniperFieldEnd {
		trackPos = append(trackPos, -(lazyField.Len() + 1))
	}
	return lazyFields, trackPos, id
}

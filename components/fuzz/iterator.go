package fuzz

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
)

func iterLenClusterbomb(lengths []int) int {
	ret := 1
	for _, l := range lengths {
		ret *= l
	}
	return ret
}

func iterLenPitchfork(lengths []int, cycle bool) int {
	if len(lengths) == 0 {
		return 0
	}
	ret := lengths[0]
	if cycle {
		for _, l := range lengths {
			if l > ret {
				ret = l
			}
		}
	} else {
		for _, l := range lengths {
			if l < ret {
				ret = l
			}
		}
	}
	return ret
}

func iterLen(p fuzzTypes.Plugin, lengths []int) int {
	switch p.Name {
	case "clusterbomb":
		return iterLenClusterbomb(lengths)
	case "pitchfork":
		return iterLenPitchfork(lengths, false)
	case "pitchfork-cycle":
		return iterLenPitchfork(lengths, true)
	case "":
		return 0
	default:
		tmp := fuzzTypes.Plugin{Name: p.Name, Args: resourcePool.AnySlices.Get(len(p.Args) + 2)}
		defer resourcePool.AnySlices.Put(tmp.Args)

		tmp.Args[0] = plugin.SelectIterLen
		tmp.Args[1] = 0 // 插件一旦编译后参数数量就无法改变了，这里做填充作用
		copy(tmp.Args[2:], p.Args)

		return plugin.IterLen(tmp, lengths)
	}
}

func iterIndexClusterbomb(lengths []int, ind int, out []int) {
	for i := len(lengths) - 1; i >= 0; i-- {
		out[i] = ind % lengths[i]
		ind /= lengths[i]
	}
}

func iterIndexPitchfork(lengths []int, ind int, out []int) {
	for i, _ := range out {
		// 如果出现某个列表长度小于等于下标，说明已经过了pitchfork的边界，此时标记结束
		if ind >= lengths[i] {
			for j, _ := range out {
				out[j] = -1
			}
			return
		}
		out[i] = ind
	}
	return
}

func iterIndexPitchforkCycle(lengths []int, ind int, out []int) {
	for i := 0; i < len(out); i++ {
		out[i] = ind % lengths[i]
	}
}

func iterIndex(lengths []int, ind int, out []int, p fuzzTypes.Plugin) {
	if len(lengths) != len(out) {
		return
	}

	switch p.Name {
	case "clusterbomb":
		iterIndexClusterbomb(lengths, ind, out)
	case "pitchfork":
		iterIndexPitchfork(lengths, ind, out)
	case "pitchfork-cycle":
		iterIndexPitchforkCycle(lengths, ind, out)
	case "":
		return
	default:
		tmp := fuzzTypes.Plugin{Name: p.Name, Args: resourcePool.AnySlices.Get(len(p.Args) + 2)}
		defer resourcePool.AnySlices.Put(tmp.Args)

		tmp.Args[0] = plugin.SelectIterIndex
		tmp.Args[1] = ind
		copy(tmp.Args[2:], p.Args)

		plugin.IterIndex(tmp, lengths, out)
	}
}

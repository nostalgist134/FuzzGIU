package stagePreprocess

// iterator.go 迭代器，新增功能，还未集成到逻辑中

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"slices"
)

func iterNumClusterbomb(lengths []int) int64 {
	if len(lengths) == 0 {
		return 0
	}
	num := int64(lengths[0])
	for _, l := range lengths[1:] {
		num *= int64(l)
	}
	return num
}

func iterNumPitchfork(lengths []int, cycle bool) int64 {
	if len(lengths) == 0 {
		return 0
	}
	if cycle {
		return int64(slices.Max(lengths))
	} else {
		return int64(slices.Min(lengths))
	}
}

func iteratorClusterbomb(lengths []int, res []int, ind int) []int {
	if len(lengths) == 0 {
		return res
	}
	if len(lengths) > len(res) {
		res = make([]int, len(lengths))
	}
	for i := len(lengths) - 1; i >= 0; i-- {
		if lengths[i] == 0 {
			res[i] = 0
			continue
		}
		res[i] = ind % lengths[i]
		ind /= lengths[i]
	}
	return res
}

func iteratorPitchfork(lengths []int, res []int, ind int, cycle bool) []int {
	if len(lengths) == 0 {
		return res
	}
	if len(lengths) > len(res) {
		res = make([]int, len(lengths))
	}
	for i, l := range lengths {
		if l == 0 {
			res[i] = 0
			continue
		}
		if cycle {
			res[i] = ind % l
		} else {
			res[i] = ind
		}
	}
	return res
}

func IterNum(iterPlug fuzzTypes.Plugin, lengths []int) int64 {
	switch iterPlug.Name {
	case "clusterbomb":
		return iterNumClusterbomb(lengths)
	case "pitchfork":
		return iterNumPitchfork(lengths, false)
	case "pitchfork-cycle":
		return iterNumPitchfork(lengths, true)
	default:
		return plugin.IterNum(iterPlug)
	}
}

// Iterator 迭代器
func Iterator(iterPlug fuzzTypes.Plugin, lengths []int, out []int, ind int) []int {
	switch iterPlug.Name {
	case "clusterbomb":
		return iteratorClusterbomb(lengths, out, ind)
	case "pitchfork":
		return iteratorPitchfork(lengths, out, ind, false)
	case "pitchfork-cycle":
		return iteratorPitchfork(lengths, out, ind, true)
	default:
		return plugin.Iterator(iterPlug, lengths, out, ind)
	}
}

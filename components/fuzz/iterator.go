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
	default:
		tmp := fuzzTypes.Plugin{Name: p.Name, Args: resourcePool.AnySlices.Get(len(p.Args) + 1)}
		defer resourcePool.AnySlices.Put(tmp.Args)

		tmp.Args[0] = plugin.SelectIterLen
		copy(tmp.Args[1:], p.Args)

		return plugin.IterLen(p, lengths)
	}
}

func iterIndexClusterbomb(lengths []int, ind int, out []int) {
	for i := len(lengths) - 1; i >= 0; i-- {
		r := ind % lengths[i]
		ind /= lengths[i]
		out[i] = r
	}
}

func iterIndexPitchfork(lengths []int, ind int, out []int) {
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
	case "pitchfork", "pitchfork-cycle":
		iterIndexPitchfork(lengths, ind, out)
	default:
		tmp := fuzzTypes.Plugin{Name: p.Name, Args: resourcePool.AnySlices.Get(len(p.Args) + 2)}
		defer resourcePool.AnySlices.Put(tmp.Args)

		tmp.Args[0] = plugin.SelectIterIndex
		tmp.Args[1] = ind
		copy(tmp.Args[2:], p.Args)

		plugin.IterIndex(tmp, lengths, out)
	}
}

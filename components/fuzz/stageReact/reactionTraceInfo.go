package stageReact

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
)

// GetReactTraceInfo 获取reaction结构中的追溯信息
func GetReactTraceInfo(reaction *fuzzTypes.Reaction) ([]string, []string) {
	markerInd := strings.Index(reaction.Output.Msg, infoMarker)
	if markerInd == -1 {
		return nil, nil
	}
	k := make([]string, 0)
	p := make([]string, 0)
	if len(reaction.Output.Msg[markerInd:]) == len(infoMarker) {
		return nil, nil
	}
	for _, kpPair := range strings.Split(reaction.Output.Msg[markerInd+len(infoMarker):], infoMarker) {
		if kpPair != "" {
			if key, payload, ok := strings.Cut(kpPair, ":"); ok {
				k = append(k, key)
				p = append(p, payload)
			}
		}
	}
	return k, p
}

func AppendReactTraceInfo(reaction *fuzzTypes.Reaction, keywords, payloads []string) {
	sb := strings.Builder{}
	sb.WriteByte('\n')
	// 写入infoMarker，避免与原先的信息冲突，InfoMarker是随机生成的12位长字符串
	sb.WriteString(infoMarker)
	for i, k := range keywords {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(payloads[i])
		if i != len(k)-1 {
			sb.WriteString(infoMarker)
		}
	}
	reaction.Output.Msg += sb.String()
}

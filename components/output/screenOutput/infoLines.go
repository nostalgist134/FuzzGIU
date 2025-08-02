package output

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strconv"
	"strings"
)

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

func ranges2String(ranges []fuzzTypes.Range) string {
	if len(ranges) == 0 {
		return "[]"
	}
	sb := strings.Builder{}
	sb.WriteByte('[')
	for i, r := range ranges {
		if r.Upper > r.Lower {
			sb.WriteString(fmt.Sprintf("%d-%d", r.Lower, r.Upper))
		} else if r.Upper == r.Lower {
			sb.WriteString(strconv.Itoa(r.Upper))
		}
		if i != len(ranges)-1 {
			sb.WriteByte(' ')
		}
	}
	sb.WriteByte(']')
	return sb.String()
}

// match2Lines 将fuzzTypes.Match结构转化为行
func match2Lines(m *fuzzTypes.Match) []string {
	ret := []string{
		"CODE  : " + ranges2String(m.Code),
		"LINES : " + ranges2String(m.Lines),
		"WORDS : " + ranges2String(m.Words),
		"SIZE  : " + ranges2String(m.Size),
		"REGEX : " + m.Regex,
	}
	if m.Time.Lower.Milliseconds() != m.Time.Upper.Milliseconds() {
		ret = append(ret,
			"TIME  : "+fmt.Sprintf("%d-%d(ms)", m.Time.Lower.Milliseconds(), m.Time.Upper.Milliseconds()))
	} else {
		ret = append(ret, "TIME  : -")
	}
	ret = append(ret, "MODE  : "+m.Mode)
	return ret
}

// recCtrl2Lines 将递归设置转化为string切片
func recCtrl2Lines(recCtrl *struct {
	RecursionDepth    int               `json:"recursion_depth,omitempty"`     // 当前递归深度
	MaxRecursionDepth int               `json:"max_recursion_depth,omitempty"` // 最大递归深度
	Keyword           string            `json:"keyword,omitempty"`
	StatCodes         []fuzzTypes.Range `json:"stat_codes,omitempty"`
	Regex             string            `json:"regex,omitempty"`
	Splitter          string            `json:"splitter,omitempty"`
}) []string {
	return []string{
		"CUR_DEPTH : " + strconv.Itoa(recCtrl.RecursionDepth),
		"MAX_DEPTH : " + strconv.Itoa(recCtrl.MaxRecursionDepth),
		"KEYWORD   : " + recCtrl.Keyword,
		"CODES     : " + ranges2String(recCtrl.StatCodes),
		"REGEX     : " + recCtrl.Regex}
}

// genInfoLines 将Fuzz结构转化为字符串切片
func genInfoLines(globInfo *fuzzTypes.Fuzz) []string {
	infoLines := []string{
		globInfo.Send.Request.URL,
		globInfo.Send.Request.Data,
		strconv.Itoa(globInfo.Misc.PoolSize),
		strconv.Itoa(globInfo.Misc.Delay),
		strconv.Itoa(globInfo.Send.Timeout),
		strconv.Itoa(globInfo.React.OutSettings.Verbosity),
		globInfo.React.OutSettings.OutputFile,
		globInfo.React.OutSettings.OutputFormat,
		fuzzTypes.Plugins2Expr(globInfo.Preprocess.Preprocessors),
		globInfo.React.Reactor,
		"FUZZ_KEYWORDS >"}

	// globInfo部分每一个单行使用的标题
	var lineTitles = []string{
		"URL",
		"SEND_DATA",
		"RP_SIZE",
		"DELAY",
		"TIMEOUT",
		"VERBOSITY",
		"OUT_FILE",
		"OUT_FORMAT",
		"PREPROCESSORS",
		"REACTORS"}

	for i := 0; i < len(lineTitles); i++ {
		infoLines[i] = fmt.Sprintf("%-13s : %s", lineTitles[i], infoLines[i])
	}

	addInfoLines := func(s []string, prefix string) {
		if prefix != "" {
			for _, str := range s {
				infoLines = append(infoLines, prefix+str)
			}
			return
		}
		infoLines = append(infoLines, s...)
	}

	addInfoLines(buildKeywordsLines(globInfo.Preprocess.PlTemp), "    ")
	addInfoLines([]string{"PROXIES >"}, "")
	addInfoLines(globInfo.Send.Proxies, "    ")
	addInfoLines([]string{"MATCHER >"}, "")
	addInfoLines(match2Lines(&globInfo.React.Matcher), "    ")
	addInfoLines([]string{"FILTER >"}, "")
	addInfoLines(match2Lines(&globInfo.React.Filter), "    ")
	addInfoLines([]string{"RECURSION >"}, "")
	addInfoLines(recCtrl2Lines(&globInfo.React.RecursionControl), "    ")
	return infoLines
}

// truncateLines 从切片中按照下标取出一个指定长度的片段，按照宽度截断后填入另一切片
func truncateLines(dst []string, src []string, ind int, maxLines int, width int) {
	for i := ind; i < len(src) && i-ind < maxLines && i-ind < len(dst); i++ {
		if len(src[i]) > width && width >= 3 {
			dst[i-ind] = src[i][:width-3] + "..."
			continue
		}
		dst[i-ind] = src[i]
	}
}

// lines2Text 将行转化为单个字符串，如果最大行数设置为-1则全部输出
func lines2Text(lines []string) string {
	sb := strings.Builder{}
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// centeredLines 将字符串切片中所有行按照基准行和宽度居中
func centeredLines(lines []string, width int) {
	baseLine := lines[(len(lines)-1)/2]
	if len(baseLine) >= width {
		return
	}
	prefixSpNum := 0
	for i := 0; i < len(baseLine); i++ {
		if baseLine[i] == ' ' {
			prefixSpNum++
		} else {
			break
		}
	}
	paddingNum := (width-len(baseLine))/2 - 1 - prefixSpNum
	if paddingNum < 0 {
		return
	}
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.Repeat(" ", paddingNum) + lines[i]
	}
}

// buildKeywordsLines 将fuzz关键字信息转化为格式化行
func buildKeywordsLines(plTmp map[string]fuzzTypes.PayloadTemp) []string {
	ret := make([]string, 0)
	for keyword, pt := range plTmp {
		ret = append(ret, fmt.Sprintf("%-7s :: Gen:[%s] Proc:[%s]", keyword,
			fuzzTypes.Plugins2Expr(pt.Generators.Gen), fuzzTypes.Plugins2Expr(pt.Processors)))
	}
	return ret
}

package tviewOutput

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"strings"
	"time"
)

const (
	colorSp   = "[-]"
	nilWColor = "[[#70aeff]nil[-]]"
)

var colors = []string{
	"[#76bdff]", // 尖头标题，最外层
	"[#00ffc0]", // 尖头标题，中间层
	"[#00ff00]", // 尖头标题，最内层
	"[#fffa7f]", // 冒号标题
}

func getColorByType(a any) string {
	switch a.(type) {
	case bool:
		return "[#ca772e]"
	case float64, int, time.Duration:
		return "[#27abb5]"
	case string:
		return "[#69a963]"
	}
	return "[-]"
}

func coloredValueStr(a any) string {
	return fmt.Sprintf("%s%v%s", getColorByType(a), a, colorSp)
}

// coloredPlugin2Expr 单个plugin转为字符串表达式
func coloredPlugin2Expr(p fuzzTypes.Plugin) string {
	if len(p.Args) == 0 && p.Name == "" {
		return nilWColor
	}
	sb := strings.Builder{}
	sb.WriteString("[#4f9fee]")
	sb.WriteString(p.Name)
	sb.WriteString(colorSp)
	// 参数列表
	if len(p.Args) != 0 {
		sb.WriteByte('(')
		for j, a := range p.Args {
			switch a.(type) {
			case string:
				sb.WriteString(fmt.Sprintf("%s\"%s\"%s", getColorByType(a), a, colorSp))
			default:
				sb.WriteString(fmt.Sprintf("%s", coloredValueStr(a)))
			}
			if j != len(p.Args)-1 {
				sb.WriteString(",")
			}
		}
		sb.WriteByte(')')
	}
	return sb.String()
}

func coloredPluginsExpr(ps []fuzzTypes.Plugin) string {
	if len(ps) == 0 {
		return nilWColor
	}
	sb := strings.Builder{}
	for i, p := range ps {
		if len(p.Args) == 0 && p.Name == "" {
			continue
		}
		sb.WriteString(coloredPlugin2Expr(p))
		if i != len(ps)-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func stringifyPlMeta(m map[string]*fuzzTypes.PayloadMeta) string {
	stringified := strings.Builder{}
	for k, p := range m {
		if p.Generators.Type == "plugin" {
			stringified.WriteString(fmt.Sprintf("\t%s%s%s : %s", "[#2dffff]", k, colorSp,
				coloredPluginsExpr(p.Generators.Gen)))
		} else {
			stringified.WriteString(fmt.Sprintf("\t%s%s%s : %s", "[#2dffff]", k, colorSp,
				p.Generators.Gen[0].Name))
		}
		if len(p.Processors) != 0 {
			stringified.WriteString(fmt.Sprintf(" <- %s", coloredPluginsExpr(p.Processors)))
		}
		stringified.WriteByte('\n')
	}
	return stringified.String()
}

func ifEmptyStr(s string) string {
	if s == "" {
		return nilWColor
	}
	return s
}

func stringifyStringSlice(slic []string, level int) string {
	prefix := strings.Repeat("\t", level)
	if len(slic) == 0 {
		return fmt.Sprintf("%s%s\n", prefix, nilWColor)
	}
	stringified := strings.Builder{}
	for _, s := range slic {
		stringified.WriteString(fmt.Sprintf("%s%s\n", prefix, s))
	}
	return stringified.String()
}

func stringifyRequest(req *fuzzTypes.Req) string {
	if req == nil {
		return "\t" + nilWColor
	}
	stringified := strings.Builder{}
	stringified.WriteString(fmt.Sprintf("\t%sURL%s : %s\n", colors[3], colorSp, req.URL))

	stringified.WriteString(fmt.Sprintf("\t%sHTTP_OPTIONS%s>\n", colors[1], colorSp))
	stringified.WriteString(fmt.Sprintf("\t\t%sHTTP_METHOD%s : %s\n", colors[3], colorSp, req.HttpSpec.Method))
	stringified.WriteString(fmt.Sprintf("\t\t%sHTTP_HEADERS%s>", colors[2], colorSp))
	if len(req.HttpSpec.Headers) == 0 {
		stringified.WriteString(" " + nilWColor + "\n")
	} else {
		stringified.WriteByte('\n')
		stringified.WriteString(stringifyStringSlice(req.HttpSpec.Headers, 3))
	}
	stringified.WriteString(fmt.Sprintf("\t\t%sHTTP_PROTO%s   : %s\n", colors[3], colorSp,
		ifEmptyStr(req.HttpSpec.Proto)))
	stringified.WriteString(fmt.Sprintf("\t\t%sFORCE_HTTPS%s  : %s\n", colors[3], colorSp,
		coloredValueStr(req.HttpSpec.ForceHttps)))
	stringified.WriteString(fmt.Sprintf("\t\t%sRANDOM_AGENT%s : %v\n", colors[3], colorSp,
		coloredValueStr(req.HttpSpec.RandomAgent)))

	stringified.WriteString(fmt.Sprintf("\t%sFIELDS%s>", colors[1], colorSp))
	if len(req.Fields) == 0 {
		stringified.WriteString(" " + nilWColor + "\n")
	} else {
		stringified.WriteByte('\n')
		for _, f := range req.Fields {
			stringified.WriteString(fmt.Sprintf("\t\t%s : %s\n", f.Name, f.Value))
		}
	}
	stringified.WriteString(fmt.Sprintf("\t%sDATA%s : ", colors[3], colorSp))
	if len(req.Data) == 0 {
		stringified.WriteString(nilWColor)
	} else {
		stringified.Write(req.Data)
	}
	return stringified.String()
}

func stringifyRanges(ranges []fuzzTypes.Range) string {
	stringified := strings.Builder{}
	if len(ranges) == 0 {
		return nilWColor
	}
	for i, r := range ranges {
		if r.Upper == r.Lower {
			stringified.WriteString(coloredValueStr(r.Lower))
		} else {
			stringified.WriteString(fmt.Sprintf("%s-%s", coloredValueStr(r.Lower), coloredValueStr(r.Upper)))
		}
		if i != len(ranges)-1 {
			stringified.WriteByte(',')
		}
	}
	return stringified.String()
}

func stringifyMatch(match *fuzzTypes.Match) string {
	if match == nil {
		return "\t" + nilWColor
	}
	stringified := strings.Builder{}
	if len(match.Lines) != 0 {
		stringified.WriteString(fmt.Sprintf("\t%sLINES%s      : %s\n", colors[3], colorSp,
			stringifyRanges(match.Lines)))
	}
	if len(match.Words) != 0 {
		stringified.WriteString(fmt.Sprintf("\t%sWORDS%s      : %s\n", colors[3], colorSp,
			stringifyRanges(match.Words)))
	}
	if len(match.Size) != 0 {
		stringified.WriteString(fmt.Sprintf("\t%sSIZE%s       : %s\n", colors[3], colorSp,
			stringifyRanges(match.Size)))
	}
	if len(match.Code) != 0 {
		stringified.WriteString(fmt.Sprintf("\t%sSTAT_CODES%s : %s\n", colors[3], colorSp,
			stringifyRanges(match.Code)))
	}
	if match.Regex != "" {
		stringified.WriteString(fmt.Sprintf("\t%sREGEX%s      : %s\n", colors[3], colorSp,
			ifEmptyStr(match.Regex)))
	}
	if !match.Time.Valid() {
		stringified.WriteString(fmt.Sprintf("\t%sTIME%s       : %s\n", colors[3], colorSp, nilWColor))
	} else {
		stringified.WriteString(fmt.Sprintf("\t%sTIME%s       : (%s, %s]\n", colors[3], colorSp,
			match.Time.Lower, match.Time.Upper))
	}
	stringified.WriteString(fmt.Sprintf("\t%sMODE%s       : %s", colors[3], colorSp, ifEmptyStr(match.Mode)))
	return stringified.String()
}

func stringifyIteration(iteration *fuzzTypes.Iteration) string {
	stringified := strings.Builder{}
	stringified.WriteString(fmt.Sprintf("\t%sITERATOR%s    : %s\n", colors[3], colorSp,
		coloredPlugin2Expr(iteration.Iterator)))
	stringified.WriteString(fmt.Sprintf("\t%sSTART_INDEX%s : %d", colors[3], colorSp, iteration.Start))
	return stringified.String()
}

func stringifyOutputSetting(outSetting *fuzzTypes.OutputSetting) string {
	if outSetting == nil {
		return "\t" + nilWColor
	}
	stringified := strings.Builder{}
	stringified.WriteString(fmt.Sprintf("\t%sOUTPUT_FLAG%s : ", colors[3], colorSp))
	if outSetting.ToWhere&outputFlag.OutToTview != 0 {
		stringified.WriteString("tview")
	}
	if outSetting.ToWhere&outputFlag.OutToFile != 0 {
		stringified.WriteString(" | file")
	}
	if outSetting.ToWhere&outputFlag.OutToStdout != 0 {
		stringified.WriteString(" | native_stdout")
	}
	if outSetting.ToWhere&outputFlag.OutToHttp != 0 {
		stringified.WriteString(" | http")
	}
	if outSetting.ToWhere&outputFlag.OutToChan != 0 {
		stringified.WriteString(" | channel")
	}
	stringified.WriteByte('\n')
	stringified.WriteString(fmt.Sprintf("\t%sVERBOSITY%s   : %s\n", colors[3], colorSp,
		coloredValueStr(outSetting.Verbosity)))
	stringified.WriteString(fmt.Sprintf("\t%sFORMAT%s      : %s\n", colors[3], colorSp,
		ifEmptyStr(outSetting.OutputFormat)))
	stringified.WriteString(fmt.Sprintf("\t%sFILE%s        : %s\n", colors[3], colorSp,
		ifEmptyStr(outSetting.OutputFile)))
	stringified.WriteString(fmt.Sprintf("\t%sHTTP_URL%s    : %s", colors[3], colorSp,
		ifEmptyStr(outSetting.HttpURL)))
	return stringified.String()
}

func stringifyJobInfo(jobInfo *fuzzTypes.Fuzz) string {
	return fmt.Sprintf(`%sKEYWORDS%s>
%s
%sREQUEST%s>
%s

%sMATCHER%s>
%s

%sFILTER%s>
%s

%sREQUEST_SETTINGS%s>
	%sFOLLOW_REDIRECTS%s  : %s
	%sMAX_RETRY%s         : %s
	%sRETRY_CODE%s        : %s
	%sRETRY_REGEX%s       : %s
	%sTIMEOUT%s           : %ss

%sRECURSION_CONTROL%s>
	%sRECURSION_DEPTH%s   : %s
	%sCURRENT_RECURSION%s : %s
	%sRECURSION_REGEX%s   : %s
	%sRECURSION_CODES%s   : %s

%sPROXIES%s      : %v
%sREACTOR%s      : %s
%sPREPROCESSOR%s : %s
%sCONCURRENCY%s  : %s
%sDELAY%s        : %v

%sITERATION%s>
%s

%sOUTPUT_SETTINGS%s>
%s`,
		colors[0], colorSp,
		stringifyPlMeta(jobInfo.Preprocess.PlMeta),
		colors[0], colorSp,
		stringifyRequest(&jobInfo.Preprocess.ReqTemplate),
		colors[0], colorSp,
		stringifyMatch(&jobInfo.React.Matcher),
		colors[0], colorSp,
		stringifyMatch(&jobInfo.React.Filter),
		colors[0], colorSp,
		colors[3], colorSp, coloredValueStr(jobInfo.Request.HttpFollowRedirects),
		colors[3], colorSp, coloredValueStr(jobInfo.Request.Retry),
		colors[3], colorSp, stringifyRanges(jobInfo.Request.RetryCodes),
		colors[3], colorSp, ifEmptyStr(jobInfo.Request.RetryRegex),
		colors[3], colorSp, coloredValueStr(jobInfo.Request.Timeout),
		colors[0], colorSp,
		colors[3], colorSp, coloredValueStr(jobInfo.React.RecursionControl.MaxRecursionDepth),
		colors[3], colorSp, coloredValueStr(jobInfo.React.RecursionControl.RecursionDepth),
		colors[3], colorSp, ifEmptyStr(jobInfo.React.RecursionControl.Regex),
		colors[3], colorSp, stringifyRanges(jobInfo.React.RecursionControl.StatCodes),
		colors[3], colorSp, jobInfo.Request.Proxies,
		colors[3], colorSp, coloredPlugin2Expr(jobInfo.React.Reactor),
		colors[3], colorSp, coloredPluginsExpr(jobInfo.Preprocess.Preprocessors),
		colors[3], colorSp, coloredValueStr(jobInfo.Control.PoolSize),
		colors[3], colorSp, jobInfo.Control.Delay,
		colors[0], colorSp,
		stringifyIteration(&jobInfo.Control.IterCtrl),
		colors[0], colorSp,
		stringifyOutputSetting(&jobInfo.Control.OutSetting))
}

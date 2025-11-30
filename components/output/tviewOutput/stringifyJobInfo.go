package tviewOutput

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"strings"
)

func stringifyPlMeta(m map[string]*fuzzTypes.PayloadMeta) string {
	stringified := strings.Builder{}
	for k, p := range m {
		stringified.WriteString(fmt.Sprintf("\t%s : %s|%s <- %s\n", k,
			fuzzTypes.Plugins2Expr(p.Generators.Gen), p.Generators.Type,
			fuzzTypes.Plugins2Expr(p.Processors)))
	}
	return stringified.String()
}

func stringifyStringSlice(slic []string, level int) string {
	prefix := strings.Repeat("\t", level)
	if slic == nil || len(slic) == 0 {
		return fmt.Sprintf("%s[nil]\n", prefix)
	}
	stringified := strings.Builder{}
	for _, s := range slic {
		stringified.WriteString(fmt.Sprintf("%s%s\n", prefix, s))
	}
	return stringified.String()
}

func stringifyRequest(req *fuzzTypes.Req) string {
	if req == nil {
		return "\t[nil]"
	}
	stringified := strings.Builder{}
	stringified.WriteString(fmt.Sprintf("\tURL : %s\n", req.URL))

	stringified.WriteString("\tHTTP_OPTIONS>\n")
	stringified.WriteString(fmt.Sprintf("\t\tHTTP_METHOD : %s\n", req.HttpSpec.Method))
	stringified.WriteString("\t\tHTTP_HEADERS>\n")
	stringified.WriteString(stringifyStringSlice(req.HttpSpec.Headers, 3))
	stringified.WriteString(fmt.Sprintf("\t\tHTTP_PROTO   : %s\n", req.HttpSpec.Proto))
	stringified.WriteString(fmt.Sprintf("\t\tFORCE_HTTPS  : %v\n", req.HttpSpec.ForceHttps))
	stringified.WriteString(fmt.Sprintf("\t\tRANDOM_AGENT : %v\n", req.HttpSpec.RandomAgent))

	stringified.WriteString("\tFIELDS>\n")
	if req.Fields == nil {
		stringified.WriteString("\t\t[nil]\n")
	} else {
		for _, f := range req.Fields {
			stringified.WriteString(fmt.Sprintf("\t\t%s : %s\n", f.Name, f.Value))
		}
	}
	stringified.WriteString("\tDATA : ")
	stringified.Write(req.Data)
	return stringified.String()
}

func stringifyRanges(ranges []fuzzTypes.Range) string {
	stringified := strings.Builder{}
	for _, r := range ranges {
		stringified.WriteString(fmt.Sprintf("[%d, %d] ", r.Lower, r.Upper))
	}
	return stringified.String()
}

func stringifyMatch(match *fuzzTypes.Match) string {
	if match == nil {
		return "\t[nil]"
	}
	stringified := strings.Builder{}
	if len(match.Lines) != 0 {
		stringified.WriteString(fmt.Sprintf("\tLINES      : %s\n", stringifyRanges(match.Lines)))
	}
	if len(match.Words) != 0 {
		stringified.WriteString(fmt.Sprintf("\tWORDS      : %s\n", stringifyRanges(match.Words)))
	}
	if len(match.Size) != 0 {
		stringified.WriteString(fmt.Sprintf("\tSIZE       : %s\n", stringifyRanges(match.Size)))
	}
	if len(match.Code) != 0 {
		stringified.WriteString(fmt.Sprintf("\tSTAT_CODES : %s\n", stringifyRanges(match.Code)))
	}
	if match.Regex != "" {
		stringified.WriteString(fmt.Sprintf("\tREGEX      : %s\n", match.Regex))
	}
	stringified.WriteString(fmt.Sprintf("\tTIME       : (%s, %s]\n", match.Time.Lower, match.Time.Upper))
	stringified.WriteString(fmt.Sprintf("\tMODE       : %s", match.Mode))
	return stringified.String()
}

func stringifyIteration(iteration *fuzzTypes.Iteration) string {
	stringified := strings.Builder{}
	stringified.WriteString(fmt.Sprintf("\tITERATOR    : %s\n",
		fuzzTypes.Plugin2Expr(iteration.Iterator)))
	stringified.WriteString(fmt.Sprintf("\tSTART_INDEX : %d", iteration.Start))
	return stringified.String()
}

func stringifyOutputSetting(outSetting *fuzzTypes.OutputSetting) string {
	if outSetting == nil {
		return "\t[nil]"
	}
	stringified := strings.Builder{}
	stringified.WriteString("\tOUTPUT_FLAG : ")
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
	stringified.WriteString(fmt.Sprintf("\tVERBOSITY : %d\n", outSetting.Verbosity))
	stringified.WriteString(fmt.Sprintf("\tFORMAT    : %s\n", outSetting.OutputFormat))
	stringified.WriteString(fmt.Sprintf("\tFILE      : %s\n", outSetting.OutputFile))
	stringified.WriteString(fmt.Sprintf("\tHTTP_URL  : %s", outSetting.HttpURL))
	return stringified.String()
}

func stringifyJobInfo(jobInfo *fuzzTypes.Fuzz) string {
	return fmt.Sprintf(`KEYWORDS>
%s
REQUEST>
%s

MATCHER>
%s

FILTER>
%s

REQUEST_SETTINGS>
	FOLLOW_REDIRECTS  : %v
	MAX_RETRIES       : %d
	RETRY_CODE        : %s
	RETRY_REGEX       : %s
	TIMEOUT           : %ds

RECURSION_CONTROL>
	RECURSION_DEPTH   : %d
	CURRENT_RECURSION : %d
	RECURSION_REGEX   : %s
	RECURSION_CODES   : %s

PROXIES      : %v
REACTOR      : %s
PREPROCESSOR : %s
CONCURRENCY  : %d
DELAY        : %v

ITERATION>
%s

OUTPUT_SETTINGS>
%s`,
		stringifyPlMeta(jobInfo.Preprocess.PlMeta),
		stringifyRequest(&jobInfo.Preprocess.ReqTemplate),
		stringifyMatch(&jobInfo.React.Matcher),
		stringifyMatch(&jobInfo.React.Filter),
		jobInfo.Request.HttpFollowRedirects,
		jobInfo.Request.Retry,
		jobInfo.Request.RetryCodes,
		jobInfo.Request.RetryRegex,
		jobInfo.Request.Timeout,
		jobInfo.React.RecursionControl.MaxRecursionDepth,
		jobInfo.React.RecursionControl.RecursionDepth,
		jobInfo.React.RecursionControl.Regex,
		stringifyRanges(jobInfo.React.RecursionControl.StatCodes),
		jobInfo.Request.Proxies,
		fuzzTypes.Plugin2Expr(jobInfo.React.Reactor),
		fuzzTypes.Plugins2Expr(jobInfo.Preprocess.Preprocessors),
		jobInfo.Control.PoolSize,
		jobInfo.Control.Delay,
		stringifyIteration(&jobInfo.Control.IterCtrl),
		stringifyOutputSetting(&jobInfo.Control.OutSetting))
}

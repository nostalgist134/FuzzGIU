package stagePreprocess

import (
	"FuzzGIU/components/plugin"
	"encoding/base64"
	"net/url"
	"regexp"
	"strings"
)

var re = regexp.MustCompile("/+")

func urlencode(s string) string {
	return url.QueryEscape(s)
}

func base64encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func addslashes(s string) string {
	return strings.Replace(s, "\"", "\\\"", -1)
}

func stripslashes(s string) string {
	ret := re.ReplaceAllString(s, "/")
	if ret[0] == '/' {
		ret = ret[1:]
	}
	return ret
}

// PayloadProcessor 对单个payload进行处理，默认的处理模块有urlencode、addslashes、base64以及给payload加后缀
func PayloadProcessor(payload string, plugins []plugin.Plugin) string {
	processedPayload := payload
	for _, p := range plugins { // 与preprocessor类似的循环
		switch p.Name {
		case "urlencode":
			processedPayload = urlencode(processedPayload)
		case "addslashes":
			processedPayload = addslashes(processedPayload)
		case "stripslashes":
			processedPayload = stripslashes(processedPayload)
		case "base64":
			processedPayload = base64encode(processedPayload)
		case "suffix":
			suffix := (p.Args[0]).(string)
			processedPayload += suffix
		case "repeat":
			processedPayload = strings.Repeat(processedPayload, p.Args[0].(int))
		default:
			p.Args = append([]any{processedPayload}, p.Args...) // payloadProcessor类型的插件中，第一个为待处理的payload
			processedPayload = plugin.PayloadProcessor(p)
		}
	}
	return processedPayload
}

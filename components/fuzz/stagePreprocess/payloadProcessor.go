package stagePreprocess

import (
	plugin2 "FuzzGIU/components/plugin"
	"encoding/base64"
	"net/url"
	"strings"
)

func urlencode(s string) string {
	return url.QueryEscape(s)
}

func base64encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func addslashes(s string) string {
	return strings.Replace(s, "\"", "\\\"", -1)
}

// PayloadProcessor 对单个payload进行处理，默认的处理模块有urlencode、addslashes、base64以及给payload加后缀
func PayloadProcessor(payload string, plugins []plugin2.Plugin) string {
	processedPayload := payload
	for _, p := range plugins { // 与preprocessor类似的循环
		switch p.Name {
		case "urlencode":
			processedPayload = urlencode(processedPayload)
		case "addslashes":
			processedPayload = addslashes(processedPayload)
		case "base64":
			processedPayload = base64encode(processedPayload)
		case "suffix":
			suffix := (p.Args[0]).(string)
			processedPayload += suffix
		default:
			p.Args = append([]any{processedPayload}, p.Args...) // payloadProcessor类型的插件中，第一个为待处理的payload
			ret := plugin2.Call(plugin2.PTypePlProc, p, nil, nil)
			retString := ret.(string)
			processedPayload = retString
		}
	}
	return processedPayload
}

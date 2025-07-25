package stageReact

import (
	"FuzzGIU/components/fuzzTypes"
	"bytes"
	"net/url"
	"strings"
)

func reactDns(req *fuzzTypes.Req, resp *fuzzTypes.Resp) *fuzzTypes.Reaction { // 对dns请求专用的react函数
	reaction := new(fuzzTypes.Reaction)
	reaction.Flag |= fuzzTypes.ReactFlagMatch
	reaction.Output.Overwrite = false
	sb := strings.Builder{}
	URL, _ := url.Parse(req.URL)
	sb.WriteString(URL.Host)
	ips := bytes.Split(resp.RawResponse, []byte("\n"))
	for _, ip := range ips {
		sb.WriteString("    ")
		sb.Write(ip)
		sb.WriteByte('\n')
	}
	reaction.Output.Msg = sb.String()
	return reaction
}

package outputable

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
	"unsafe"
)

const (
	strNil      = "[nil]"
	nilWColor   = "[[#70aeff]nil[-]]"
	empty       = "[empty]"
	emptyWColor = "[[#c5a97a]empty[-]]"
)

func nilIfColor(color bool) string {
	if color {
		return nilWColor
	}
	return strNil
}

func emptyIfColor(color bool) string {
	if color {
		return emptyWColor
	}
	return empty
}

func outObj2Json(obj *OutObj) []byte {
	outObjJson, err := json.Marshal(obj)
	if err != nil {
		errMsg, _ := json.Marshal(err.Error())
		outObjJson = []byte(fmt.Sprintf(`{"error":"json marshal failed - %s"}`, string(errMsg)))
	}
	return outObjJson
}

func outObj2Xml(obj *OutObj) []byte {
	outObjXml, err := xml.Marshal(obj)
	if err != nil {
		errMsg, _ := xml.Marshal(err.Error())
		outObjXml = []byte(fmt.Sprintf(`<error>xml marshal failed - %s</error>`, string(errMsg)))
	}
	return outObjXml
}

func getBytesFirstLine(b []byte) []byte {
	if b == nil {
		return nil
	}
	i := -1
	if i = bytes.Index(b, []byte("\r\n")); i != -1 {
		return b[:i]
	}
	if i = bytes.IndexByte(b, '\n'); i != -1 {
		return b[:i]
	}
	return b
}

func clearColors(colors []string, useColor bool) {
	if useColor {
		return
	}
	for i, _ := range colors {
		colors[i] = ""
	}
}

func getColorSplitter(color bool) string {
	if color {
		return "[-]"
	}
	return ""
}

func kwPlPair(keywords, payloads []string, color bool, level int) string {
	sb := strings.Builder{}
	sp := getColorSplitter(color)
	indent := strings.Repeat("\t", level)
	colors := []string{"[#3af4f1]"}
	clearColors(colors, color)
	for i := 0; i < len(keywords) && i < len(payloads); i++ {
		sb.WriteString(indent)
		sb.WriteString(colors[0])
		sb.WriteString(keywords[i])
		sb.WriteString(sp)
		sb.WriteString(" : ")
		sb.WriteString(payloads[i])
		sb.WriteByte('\n')
	}
	if sb.Len() == 0 {
		sb.WriteString(indent)
		sb.WriteString(nilIfColor(color))
		sb.WriteByte('\n')
	}
	return sb.String()
}

func resp2FmtNative(resp *fuzzTypes.Resp, color bool, verbosity int) string {
	sb := strings.Builder{}

	defer sb.WriteByte('\n')

	if resp == nil {
		sb.WriteString("{nil response}\n")
		return sb.String()
	}

	colors := []string{"[#46f758]", "[blue]", "[#3af4f1]", "[orange]", "[#ff5bee]", "[red]"}
	colorSp := getColorSplitter(color)
	clearColors(colors, color)

	if resp.ErrMsg != "" {
		sb.WriteString(fmt.Sprintf("%sERROR%s : ", colors[5], colorSp))
		sb.WriteString(resp.ErrMsg)
		sb.WriteByte('\n')
	}

	httpStat := 0
	if resp.HttpResponse != nil {
		httpStat = resp.HttpResponse.StatusCode
	}
	sb.WriteString(
		fmt.Sprintf(
			"%sRESPONSE%s : [%sSIZE%s: %d|%sLINES%s: %d|%sWORDS%s: %d|%sTIME%s: %v|%sHTTP_STATUS%s: %d]\n",
			colors[4], colorSp, colors[2], colorSp, resp.Size, colors[2], colorSp, resp.Lines, colors[2], colorSp,
			resp.Words, colors[2], colorSp, resp.ResponseTime, colors[2], colorSp, httpStat))

	if resp.HttpRedirectChain != "" {
		sb.WriteString(fmt.Sprintf("%sHTTP_REDIRECT%s : ", colors[3], colorSp))
		sb.WriteString(resp.HttpRedirectChain)
		sb.WriteByte('\n')
	}

	rawRespToWrite := resp.RawResponse
	if verbosity >= 3 {
		if len(rawRespToWrite) == 0 {
			sb.WriteString(fmt.Sprintf("%sRAW_RESPONSE%s: ", colors[0], colorSp))
		} else {
			sb.WriteString(fmt.Sprintf("%sRAW_RESPONSE%s↓\n", colors[0], colorSp))
		}
		if rawRespToWrite == nil {
			rawRespToWrite = []byte(nilIfColor(color))
		} else if len(rawRespToWrite) == 0 {
			rawRespToWrite = []byte(emptyIfColor(color))
		}
		sb.Write(rawRespToWrite)
	} else {
		sb.WriteString(fmt.Sprintf("%s└>%s ", colors[0], colorSp))
		rawRespToWrite = getBytesFirstLine(rawRespToWrite)
		if rawRespToWrite == nil {
			rawRespToWrite = []byte(nilIfColor(color))
			sb.Write(rawRespToWrite)
		} else if len(rawRespToWrite) == 0 {
			rawRespToWrite = []byte(emptyIfColor(color))
			sb.Write(rawRespToWrite)
		} else {
			sb.WriteString(colors[1])
			sb.Write(rawRespToWrite)
			sb.WriteString(colorSp)
		}
	}
	sb.WriteByte('\n')
	return sb.String()
}

func req2FmtNative(req *fuzzTypes.Req, color bool, verbosity int) string {
	if verbosity == 1 {
		return ""
	}
	colors := []string{"[orange]", "[#7589e4]", "[#efc894]", "[#46f758]"}
	colorSp := getColorSplitter(color)
	clearColors(colors, color)

	sb := strings.Builder{}

	writeStringColor := func(s string) {
		sb.WriteString(colors[0])
		sb.WriteString(s)
		sb.WriteString(colorSp)
	}

	writeWTitle := func(content, title string, level int) {
		sb.Write(bytes.Repeat([]byte{'\t'}, level))
		writeStringColor(title)
		sb.WriteString(" : ")
		if content == "" {
			sb.WriteString(nilIfColor(color))
		} else {
			sb.WriteString(content)
		}
		sb.WriteByte('\n')
	}

	switch verbosity {
	case 2:
		url := ""
		if req != nil {
			url = req.URL
		}
		writeWTitle(url, "URL", 0)
		return sb.String()
	case 3:
		sb.WriteString(fmt.Sprintf("%sREQUEST%s>\n", colors[1], colorSp))
		if req == nil {
			sb.WriteByte('\t')
			sb.WriteString(nilIfColor(color))
			sb.WriteByte('\n')
			return sb.String()
		}

		writeWTitle(req.HttpSpec.Method, "METHOD", 1)
		writeWTitle(req.URL, "URL", 1)
		writeWTitle(req.HttpSpec.Proto, "PROTO", 1)

		sb.WriteString(fmt.Sprintf("\t%sHTTP_HEADERS%s>\n", colors[3], colorSp))
		if len(req.HttpSpec.Headers) == 0 {
			sb.WriteString("\t\t")
			sb.WriteString(nilIfColor(color))
		} else {
			for i, h := range req.HttpSpec.Headers {
				sb.WriteString("\t\t")
				sb.WriteString(h)
				if i != len(req.HttpSpec.Headers)-1 {
					sb.WriteByte('\n')
				}
			}
		}
		sb.WriteByte('\n')

		sb.WriteString(fmt.Sprintf("\t%sFIELDS%s>\n", colors[3], colorSp))
		if len(req.Fields) == 0 {
			sb.WriteString("\t\t")
			sb.WriteString(nilIfColor(color))
			sb.WriteByte('\n')
		} else {
			for _, f := range req.Fields {
				writeWTitle(f.Value, f.Name, 2)
			}
		}

		sb.WriteString(fmt.Sprintf("\t%sDATA%s : ", colors[1], colorSp))
		dataToWrite := []byte(nilIfColor(color))
		if len(req.Data) != 0 {
			dataToWrite = req.Data
			sb.WriteByte('`')
		}
		sb.Write(dataToWrite)
		if len(req.Data) != 0 {
			sb.WriteByte('`')
		}
	}
	sb.WriteByte('\n')
	return sb.String()
}

func nativeOutObj(obj *OutObj, color bool, verbosity int) []byte {
	if obj == nil {
		return []byte(nilIfColor(color))
	}
	colors := []string{"[#3af4f1]", "[#7589e4]", "[red]"}
	colorSp := getColorSplitter(color)
	clearColors(colors, color)
	buf := bytes.Buffer{}
	tagWColor := func(tag string) {
		buf.WriteString(fmt.Sprintf("%s%s%s>\n", colors[1], tag, colorSp))
	}
	tagWColor("PAYLOADS")
	buf.WriteString(kwPlPair(obj.Keywords, obj.Payloads, color, 1))
	buf.WriteString(req2FmtNative(obj.Request, color, verbosity))
	if obj.Msg != "" {
		buf.WriteString(fmt.Sprintf("%sMSG%s : ", colors[2], colorSp))
		buf.WriteString(obj.Msg)
		buf.WriteByte('\n')
	}
	buf.WriteString(resp2FmtNative(obj.Response, color, verbosity))
	return buf.Bytes()
}

// ToFormatBytes 将OutObj转化为特定表示格式的字节流
func (o *OutObj) ToFormatBytes(format string, color bool, verbosity int) []byte {
	switch format {
	case "xml":
		return outObj2Xml(o)
	case "json", "json-line":
		return outObj2Json(o)
	case "native":
		return nativeOutObj(o, color, verbosity)
	}

	return nil
}

func (o *OutObj) ToFormatStr(format string, color bool, verbosity int) string {
	fmtBytes := o.ToFormatBytes(format, color, verbosity)
	return unsafe.String(&fmtBytes[0], len(fmtBytes))
}

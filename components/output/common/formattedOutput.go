package common

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
)

var coloredSplit = "[----------------------------------------------------------------------------------------------------](fg:red)"
var split = "----------------------------------------------------------------------------------------------------"

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

func nativeOutputMsg(obj *OutObj, ignoreError bool, verbosity int) []byte {
	bb := bytes.Buffer{}
	respFirstLine := bytes.Split(obj.Response.RawResponse, []byte{'\n'})[0]
	writeFmtStr := func(title string, val string) {
		if val == "" {
			return
		}
		bb.WriteString(fmt.Sprintf("%-8s : %s\n", title, val))
	}
	// 输出fuzz关键字和payload
	if len(obj.Keywords) != 0 {
		bb.WriteString("PAYLOAD :\n")
		if len(obj.Keywords) == 1 {
			bb.WriteString(fmt.Sprintf("    %s", obj.Keywords[0]))
			bb.WriteByte('\n')
		} else {
			for i, k := range obj.Keywords {

				bb.WriteString(fmt.Sprintf("    %-8s : %s", k, obj.Payloads[i]))
				bb.WriteByte('\n')
			}
		}
	}
	// 输出响应相关数据
	resp := obj.Response
	bb.WriteString(fmt.Sprintf("RESPONSE : [SIZE = %d|LINES = %d|WORDS = %d|TIME = %dms", resp.Size, resp.Lines,
		resp.Words, resp.ResponseTime.Milliseconds()))
	if resp.HttpResponse != nil {
		bb.WriteString(fmt.Sprintf("|HTTP_CODE = %d", resp.HttpResponse.StatusCode))
	}
	bb.Write([]byte{']', '\n'})
	// 输出Reaction自定义消息以及错误信息
	writeFmtStr("MESSAGE", obj.Msg)
	if !ignoreError {
		writeFmtStr("ERROR", resp.ErrMsg)
	}
	// 根据输出详细程度输出其它信息
	switch verbosity {
	case 1:
		bb.WriteString(fmt.Sprintf(" └> %s\n", string(respFirstLine)))
	case 2:
		writeFmtStr("URL", obj.Request.URL)
		writeFmtStr("REQ_DATA", obj.Request.Data)
		bb.WriteString(fmt.Sprintf(" └> %s\n", string(respFirstLine)))
	case 3:
		j, _ := json.Marshal(obj.Request)
		bb.Write(j)
		bb.Write([]byte("\n    |\n    V\n"))
		bb.Write(obj.Response.RawResponse)
		bb.WriteByte('\n')
	}
	bb.WriteString(split)
	return bb.Bytes()
}

func coloredNativeOutputMsg(obj *OutObj, ignoreError bool, verbosity int) []byte {
	bb := bytes.Buffer{}
	respFirstLine := bytes.Split(obj.Response.RawResponse, []byte{'\n'})[0]
	if len(respFirstLine) == 0 {
		respFirstLine = []byte{'[', 'n', 'i', 'l', ']'}
	}
	writeFmtStr := func(title string, val string) {
		if val == "" {
			return
		}
		bb.WriteString(fmt.Sprintf("[%-8s](fg:yellow) : %s\n", title, val))
	}
	// 输出fuzz关键字和payload
	if len(obj.Keywords) != 0 {
		bb.WriteString("[PAYLOAD](fg:yellow) :\n")
		if len(obj.Keywords) == 1 {
			bb.WriteString(fmt.Sprintf("    %s", obj.Payloads[0]))
			bb.WriteByte('\n')
		} else {
			for i, k := range obj.Keywords {
				bb.WriteString(fmt.Sprintf("    [%-8s](fg:blue) : %s", k, obj.Payloads[i]))
				bb.WriteByte('\n')
			}
		}
	}
	// 输出响应相关数据
	resp := obj.Response
	bb.WriteString(fmt.Sprintf("[RESPONSE](fg:yellow) : {SIZE = %d|LINES = %d|WORDS = %d|TIME = %dms",
		resp.Size, resp.Lines, resp.Words, resp.ResponseTime.Milliseconds()))
	if resp.HttpResponse != nil {
		bb.WriteString(fmt.Sprintf("|HTTP_CODE = %d", resp.HttpResponse.StatusCode))
	}
	bb.Write([]byte{'}', '\n'})
	if resp.HttpRedirectChain != "" {
		writeFmtStr("HTTP_RDR", resp.HttpRedirectChain)
	}
	// 输出Reaction自定义消息以及错误信息
	writeFmtStr("MESSAGE", obj.Msg)
	if !ignoreError {
		writeFmtStr("ERROR", resp.ErrMsg)
	}
	// 根据输出详细程度输出其它信息
	switch verbosity {
	case 1:
		bb.WriteString(fmt.Sprintf(" [└>](fg:green) [%s](fg:cyan)\n", string(respFirstLine)))
	case 2:
		writeFmtStr("URL", obj.Request.URL)
		writeFmtStr("REQ_DATA", obj.Request.Data)
		bb.WriteString(fmt.Sprintf(" [└>](fg:green) [%s](fg:cyan)\n", string(respFirstLine)))
	case 3:
		j, _ := json.Marshal(obj.Request)
		bb.Write(j)
		bb.WriteString("\n    |\n    V\n")
		bb.Write(obj.Response.RawResponse)
		bb.WriteByte('\n')
	}
	bb.WriteString(coloredSplit)
	return bb.Bytes()
}

// FormatObjOutput 根据格式将OutObj输出
func FormatObjOutput(obj *OutObj, format string, color bool) []byte {
	formatOutput := []byte("")
	switch format {
	case "xml":
		formatOutput = outObj2Xml(obj)
	case "json":
		formatOutput = outObj2Json(obj)
	case "native":
		if color {
			formatOutput = coloredNativeOutputMsg(obj, GlobOutSettings.IgnoreError, GlobOutSettings.Verbosity)
		} else {
			formatOutput = nativeOutputMsg(obj, GlobOutSettings.IgnoreError, GlobOutSettings.Verbosity)
		}
	}
	return formatOutput
}

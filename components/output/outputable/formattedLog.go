package outputable

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"time"
	"unsafe"
)

func log2Json(log *Log) []byte {
	marshaled, err := json.Marshal(log)
	if err != nil {
		errMsg, _ := json.Marshal(err.Error())
		marshaled = []byte(fmt.Sprintf(`{"error":"json marshal failed - %s"}`, string(errMsg)))
	}
	return marshaled
}

func log2Xml(log *Log) []byte {
	marshaled, err := xml.Marshal(log)
	if err != nil {
		errMsg, _ := xml.Marshal(err.Error())
		marshaled = []byte(fmt.Sprintf(`<error>xml marshal failed - %s</error>`, string(errMsg)))
	}
	return marshaled
}

func log2NativeFmt(log *Log) []byte {
	return []byte(fmt.Sprintf("[LOG @ JOB#%d %s] %s", log.Jid, log.Time.Format("02/01/2006 15:04:05"),
		log.Msg))
}

// ToFormatBytes 将log转化为指定格式的字节流表示
func (log *Log) ToFormatBytes(format string) []byte {
	log2 := *log
	if log2.Time.IsZero() {
		log2.Time = time.Now()
	}
	switch format {
	case "xml":
		return log2Xml(&log2)
	case "json", "json-line":
		return log2Json(&log2)
	case "native":
		return log2NativeFmt(&log2)
	default:
		return nil
	}
}

func (log *Log) ToFormatStr(format string) string {
	fmtBytes := log.ToFormatBytes(format)
	return unsafe.String(&fmtBytes[0], len(fmtBytes))
}

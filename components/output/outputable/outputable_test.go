package outputable

import (
	"fmt"
	"testing"
	"time"
)

func TestLog_ToFormatBytes(t *testing.T) {
	l := &Log{
		Jid:  0,
		Msg:  "MIALGOU",
		Time: time.Now(),
	}
	b := l.ToFormatStr("json-line")
	fmt.Println(b)
}

func TestOutObj_ToFormatStr(t *testing.T) {
	o := &OutObj{
		Jid:      9,
		Keywords: []string{"NISHIGIU", "WOSHIGIU", "MILAOGIU"},
		Payloads: []string{"WOSHIGIU", "MILAOGIU", "NISHIGIU"},
		Request:  nil,
		Response: nil,
		Msg:      "MILAOGIU",
		Time:     time.Now(),
	}
	fmt.Println(o.ToFormatStr("json-line", false, 0))
}

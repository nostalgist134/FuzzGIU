package outputable

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
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
		Jid: 9,
		Request: &fuzzTypes.Req{URL: "https://nishigiu.com", Data: []byte("NISHI=GIU&WOSHI=GIU\nMILOGIU"),
			HttpSpec: fuzzTypes.HTTPSpec{Method: "GET", Headers: []string{"NISHI: GIU", "WOSHI: GIU", "MILAO: GIU"},
				Proto: "2.99", ForceHttps: true, RandomAgent: true}, Fields: []fuzzTypes.Field{
				{"NISHIGIU", "WOSHIGIU"}, {"NISHIGIU", "MILAOGIU"},
			}},
		Response: &fuzzTypes.Resp{RawResponse: []byte(""),
			ResponseTime: 834194 * time.Microsecond},
		Msg:  "MILAOGIU",
		Time: time.Now(),
	}
	f := o.ToFormatStr("native", true, 2)
	fmt.Println(f)
}

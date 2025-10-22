package tmplReplace

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
	"testing"
)

func TestCountFields(t *testing.T) {

}

func nop(_ any) {}

func TestReq2Str(t *testing.T) {
	milaogiu := &fuzzTypes.Req{
		URL: "https://milaogiu.com/woshigiu",
		HttpSpec: fuzzTypes.HTTPSpec{
			Method:      "MILAOGIU",
			Headers:     []string{"NISHI:GIU", "WOSHI:GIU", "MILAO:GIU"},
			Version:     "2.999",
			ForceHttps:  false,
			RandomAgent: false,
		},
		Fields: []fuzzTypes.Field{
			{
				Name:  "GIU",
				Value: "TEST",
			},
			{
				Name:  "TESTGIU",
				Value: "MILAOGIU",
			},
		},
		Data: []byte("NISHIGIUWOSHIGIUMILAOGIU"),
	}
	least := &fuzzTypes.Req{
		URL:      "http://milaogiu.com",
		HttpSpec: fuzzTypes.HTTPSpec{Method: "GET", Version: "2.9"},
		Data:     []byte("MILAOGIU"),
	} // least是一个“最小非空”的req实例，即使req结构为空，req2str也会解析least所包含的这4个字段
	nop(milaogiu)
	stringified, splitter := req2Str(least)
	fmt.Println(stringified)
	fmt.Println(splitter)
	splitted := strings.Split(stringified, splitter)
	for _, sp := range splitted {
		fmt.Println(sp)
	}
	fmt.Println(len(splitted))
}

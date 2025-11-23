package tmplReplace

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
	"testing"
)

var milaogiu = &fuzzTypes.Req{
	URL: "https://mi_FUZZ1_laogiu.com/woshigiu",
	HttpSpec: fuzzTypes.HTTPSpec{
		Method:      "FUZZ1MIL_FUZZ2_AOGIU",
		Headers:     []string{"NIS_FUZZ1_HI:GIU", "WOSHI:GIU", "MILAO:GIU"},
		Proto:       "2.99_FUZZ1_9",
		ForceHttps:  false,
		RandomAgent: false,
	},
	Fields: []fuzzTypes.Field{
		{
			Name:  "GIU",
			Value: "T_FUZZ3__FUZZ2_E_FUZZ1_ST",
		},
		{
			Name:  "TEST_FUZZ2_GIU",
			Value: "MILA_FUZZ1_OGIU",
		},
	},
	Data: []byte("NISHIG_FUZZ2_IUWOSH_FUZZ1_IGIU_FUZZ3_MILAOGIU"),
}

var milaogiu2 = &fuzzTypes.Req{
	URL: "_FUZZ1__FUZZ1__FUZZ1_",
	HttpSpec: fuzzTypes.HTTPSpec{
		Method:      "FUZZ1MIL_FUZZ1_AOGIU",
		Headers:     []string{"NIS_FUZZ1_HI:GIU", "WOSHI:GIU", "MILAO:GIU"},
		Proto:       "2.99_FUZZ1",
		ForceHttps:  false,
		RandomAgent: false,
	},
	Fields: []fuzzTypes.Field{},
	Data:   []byte("NISHIG_FUZZ2_IUWOSH_FUZZ1_IGIU_FUZZ1_MILAOGIU"),
}

var milaogiu3 = &fuzzTypes.Req{URL: "https://www.baidu.com/FUZZ", HttpSpec: fuzzTypes.HTTPSpec{Method: "GET"}}

func TestCountFields(t *testing.T) {

}

func TestReq2Str(t *testing.T) {
	least := &fuzzTypes.Req{
		URL:      "http://milaogiu.com",
		HttpSpec: fuzzTypes.HTTPSpec{Method: "GET", Proto: "2.9"},
		Data:     []byte("MILAOGIU"),
	} // least是一个“最小非空”的req实例，即使req结构为空，req2str也会解析least所包含的这4个字段
	_ = milaogiu
	stringified, splitter := req2Str(least)
	fmt.Println(stringified)
	fmt.Println(splitter)
	splitted := strings.Split(stringified, splitter)
	for _, sp := range splitted {
		fmt.Println(sp)
	}
	fmt.Println(len(splitted))
}

func TestLazy(t *testing.T) {
	a := lazyPool.Get(10)
	fmt.Println(a)
}

func TestParseReqTmpl(t *testing.T) {
	/*tmpl := ParseReqTmpl(milaogiu, []string{"FUZZ1", "FUZZ2", "FUZZ3"})
	r, id := tmpl.Replace([]string{"AAA", "BBB", "CCC"}, -1)
	b, _ := json.MarshalIndent(r, "", "  ")
	fmt.Println(string(b))
	fmt.Println(id)*/
	/*tmpl2 := ParseReqTmpl(milaogiu, []string{"FUZZ1"})
	for i := 0; i < tmpl2.KeywordCount(0); i++ {
		r2, _ := tmpl2.Replace([]string{"AAA"}, i)
		fmt.Println(r2)
	}*/
	tmpl3 := ParseReqTmpl(milaogiu2, []string{"FUZZ1"})
	r3, tr, _ := tmpl3.ReplaceTrack("woshigiu", -1)
	fmt.Println(r3)
	fmt.Println(tr)
	fmt.Println(len(r3.HttpSpec.Proto))
	/*for i := 0; i < tmpl3.KeywordCount(0); i++ {
		r2, _ := tmpl3.Replace([]string{"AAA"}, i)
		fmt.Println(r2)
	}*/
}

func Test4(t *testing.T) {
	tmpl1 := ParseReqTmpl(milaogiu3, []string{"FUZZ"})
	r0, err := tmpl1.Replace([]string{"FCKeditor/editor/filemanager/browser/default/browser.html?Type=Image&Connector=connectors/jsp/connector"}, -1)
	fmt.Println(r0)
	fmt.Println(err)
}

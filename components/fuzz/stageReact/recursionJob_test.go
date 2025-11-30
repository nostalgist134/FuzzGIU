package stageReact

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/tmplReplace"
	"testing"
)

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
	Data:   []byte("NISHIG_FUZZ2_IUWOSH_FUZZ1_IGIU_FUZZ1"),
}

func TestRecursive1(t *testing.T) {
	tmpl := tmplReplace.ParseReqTmpl(milaogiu2, []string{"FUZZ1"})
	r, track, _ := tmpl.ReplaceTrack("MILAOGIU", 0)
	fmt.Println(track)
	newJob := deriveRecursionJob(&fuzzTypes.Fuzz{React: fuzzTypes.FuzzStageReact{
		RecursionControl: fuzzTypes.ReactRecursionControl{
			Keyword:  "FUZZ",
			Splitter: "/",
		},
	}}, r, track)
	fmt.Println(newJob.Preprocess.ReqTemplate)
	fmt.Println(string(newJob.Preprocess.ReqTemplate.Data))
}

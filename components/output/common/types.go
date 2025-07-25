package common

import (
	"FuzzGIU/components/fuzzTypes"
	"encoding/xml"
)

// OutObj 用于输出的结构体
type OutObj struct {
	XMLName  xml.Name        `json:"-" xml:"output"`
	Keywords []string        `json:"keywords" xml:"keywords>keyword"`
	Payloads []string        `json:"payloads" xml:"payloads>payload"`
	Request  *fuzzTypes.Req  `json:"request"  xml:"request"`
	Response *fuzzTypes.Resp `json:"response" xml:"response"`
	Msg      string          `json:"msg,omitempty" xml:"msg,omitempty"`
}

type counter struct {
	count int64
	total int64
}

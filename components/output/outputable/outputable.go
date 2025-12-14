package outputable

import (
	"encoding/xml"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"time"
)

// OutObj 用于输出的结构体
type OutObj struct {
	Id       uint            `json:"-" xml:"-" gorm:"primarykey"`
	XMLName  xml.Name        `json:"-" xml:"output"`
	Keywords []string        `json:"keywords" xml:"keywords>keyword"`
	Payloads []string        `json:"payloads" xml:"payloads>payload"`
	Request  *fuzzTypes.Req  `json:"request"  xml:"request"`
	Response *fuzzTypes.Resp `json:"response" xml:"response"`
	Msg      string          `json:"msg,omitempty" xml:"msg,omitempty"`
	Time     time.Time       `json:"time,omitempty" xml:"time,omitempty"`
}

type Log struct {
	XMLName xml.Name  `json:"-" xml:"log"`
	Msg     string    `json:"msg" xml:"msg"`
	Time    time.Time `json:"time" xml:"time"`
}

package fileOutput

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"github.com/xyproto/randomstring"
	"sync"
	"testing"
	"time"
)

func TestCtxOutputAndLog(t *testing.T) {
	oSetting := &fuzzTypes.OutputSetting{
		Verbosity:    0,
		OutputFile:   "C:\\Users\\patrick\\Desktop\\test123351.json",
		OutputFormat: "json-line",
		HttpURL:      "",
		ChanSize:     0,
		ToWhere:      0,
	}
	c, err := NewFileOutputCtx(oSetting, 0)
	if err != nil {
		t.Fatal(err)
	}
	outObj := &outputable.OutObj{
		Keywords: nil,
		Payloads: nil,
		Request:  nil,
		Response: nil,
		Msg:      "",
		Time:     time.Now(),
	}
	l := &outputable.Log{
		Msg:  "MILAOGIU",
		Time: time.Now(),
	}
	err = c.Output(outObj)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Log(l)
	if err != nil {
		t.Fatal(err)
	}
	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAsnycOutputAndLog(t *testing.T) {
	oSetting := &fuzzTypes.OutputSetting{
		Verbosity:    0,
		OutputFile:   "C:\\Users\\patrick\\Desktop\\test1299.txt",
		OutputFormat: "xml",
		HttpURL:      "",
		ChanSize:     0,
		ToWhere:      0,
	}
	c, err := NewFileOutputCtx(oSetting, 0)
	if err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	async := func(msg string) {
		wg.Add(1)
		defer wg.Done()
		outObj := &outputable.OutObj{
			Keywords: nil,
			Payloads: nil,
			Request:  nil,
			Response: nil,
			Msg:      msg,
			Time:     time.Now(),
		}
		log := &outputable.Log{
			Time: time.Now(),
			Msg:  msg,
		}
		err = c.Output(outObj)
		if err != nil {
			return
		}
		err = c.Log(log)
	}
	for i := 0; i < 50; i++ {
		msg := randomstring.String(5)
		go async(msg)
	}
	wg.Wait()
	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}
}

package stdoutOutput

import (
	"encoding/xml"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/counter"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	oSetting := &fuzzTypes.OutputSetting{
		OutputFormat: "json",
	}
	cntr := &counter.Counter{
		StartTime:    time.Time{},
		TaskRate:     0,
		Errors:       0,
		TaskProgress: counter.Progress{},
	}
	cntr.Set(counter.CntrTask, counter.FieldTotal, 999)
	cntr.Set(counter.CntrErrors, counter.FieldTotal, 999)
	c, err := NewStdoutCtx(oSetting, 3)
	if c == nil {
		t.Fatal("c is nil")
	}
	err = c.RegisterCounter(cntr)
	if err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		cntr.StartRecordTaskRate()
		for i := 0; i < 5000; i++ {
			cntr.Complete(counter.CntrTask)
			if rand.Int()%100 == 1 {
				o := &outputable.OutObj{
					Id:       0,
					XMLName:  xml.Name{},
					Jid:      rand.Int() % 1000,
					Keywords: nil,
					Payloads: nil,
					Request:  nil,
					Response: nil,
					Msg:      "milaogiu",
					Time:     time.Now(),
				}
				if err := c.Output(o); err != nil {
					return
				}
			}
			if rand.Int()%1000 == 5 {
				l := outputable.Log{
					Jid:  5,
					Msg:  "nishigiuwoshigiumilaogiu",
					Time: time.Now(),
				}
				if err := c.Log(&l); err != nil {
					return
				}
			}
			time.Sleep(3 * time.Millisecond)
		}
		wg.Done()
	}()
	time.Sleep(5 * time.Second)
	wg.Wait()
	err = c.Close()
	if err != nil {
		t.Fatal(err)
	}
}

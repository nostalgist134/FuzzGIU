package libfgiu

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputFlag"
	"testing"
)

func TestFuzzer(t *testing.T) {
	testF := new(fuzzTypes.Fuzz)
	testF.Preprocess.PlMeta = make(map[string]*fuzzTypes.PayloadMeta)
	testF.Preprocess.PlMeta["FUZZ"] = &fuzzTypes.PayloadMeta{
		Generators: fuzzTypes.PlGen{},
		Processors: nil,
	}
	testF.Preprocess.ReqTemplate = fuzzTypes.Req{
		URL: "https://www.baidu.com/FUZZ",
		HttpSpec: fuzzTypes.HTTPSpec{
			Proto: "HTTP/2.0",
		},
	}
	testF.Control.IterCtrl.Iterator = fuzzTypes.Plugin{Name: "clusterbomb"}
	testF.Control.PoolSize = 1
	testF.Control.OutSetting = fuzzTypes.OutputSetting{
		Verbosity:    3,
		OutputFile:   "",
		OutputFormat: "json",
		HttpURL:      "",
		ChanSize:     0,
		ToWhere:      outputFlag.OutToStdout,
	}
	f, err := NewFuzzer(10)
	if err != nil {
		t.Fatal(err)
	}
	f.Start()
	fmt.Println(f.Submit(testF))
	f.Wait()
}

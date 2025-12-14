package libfgiu

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"testing"
)

func TestCutGenProcArg(t *testing.T) {
	f := &fuzzTypes.Fuzz{}
	wordlists := []string{
		"test.txt,H:\\test.txt,C:\\nishigiu\\woshigiu.txt::FUZZ",
		"test2.txt,H:\\MILAOGIU.txt,home/test3.txt,test4.txt::FUZZ",
		"test2.txt,H:\\MILAOGIU.txt,home/test3.txt,test5txt.txt::MILAOGGIU",
		"test2.txt,H:\\MILAOGIU.txt,home/test3.txt,test4txt.txt::MILAOGGIU",
	}
	gens := []string{
		"permuteex('hello',1,2),int(1,100),int(10,50)::GIUF",
		"permuteex('hello',1,2),int(1,100),int(10,70)::GIUF",
		"int(9,10,100),woshigiu(1,2,3)::GIU2",
		"int(9,10,100),woshigiu(1,2,3)::GIU3",
		"int(9,10,100),woshigiu(1,2,3)",
	}
	proc := []string{
		"suffix('.txt'),base64,aes",
		"suffix('.txt'),base64,aes::FUZZ",
		"aes,rsa,3des::FUZZ",
		"suffix('.php')::GIUF",
	}
	err := appendPlGen(f, wordlists, "wordlist")
	if err != nil {
		t.Fatal(err)
	}
	err = appendPlGen(f, gens, "plugin")
	if err != nil {
		t.Fatal(err)
	}
	err = appendPlProc(f, proc)
	if err != nil {
		t.Fatal(err)
	}
	for k, m := range f.Preprocess.PlMeta {
		fmt.Println(k)
		fmt.Println(m.Processors)
	}
}

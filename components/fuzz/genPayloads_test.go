package fuzz

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"testing"
)

func TestGenPayloads(t *testing.T) {
	jc := &fuzzCtx.JobCtx{Job: &fuzzTypes.Fuzz{}}
	job := jc.Job
	job.Preprocess.PlMeta = make(map[string]*fuzzTypes.PayloadMeta)
	job.Preprocess.PlMeta["GIU"] = &fuzzTypes.PayloadMeta{
		Generators: fuzzTypes.PlGen{
			Wordlists: []string{"C:/Users/patrick/Desktop/uf.txt", "C:/Users/patrick/Desktop/df.txt"},
			Plugins: []fuzzTypes.Plugin{
				{"permuteex", []any{"abcdefghijklmnopqrstuvwxyz", 2, 3}},
				{"int", []any{1, 100, 10, 3}},
			},
		}}
	genPayloads(jc)
	for _, p := range job.Preprocess.PlMeta["GIU"].PlList {
		fmt.Println(p)
	}
}

package fuzz

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
)

func deduplicatePayloads(payloads []string) []string {
	m := make(map[string]struct{})
	deduplicated := make([]string, 0)
	for _, p := range payloads {
		if _, ok := m[p]; ok {
			continue
		}
		m[p] = struct{}{}
		deduplicated = append(deduplicated, p)
	}
	return deduplicated
}

func genPayloads(jobCtx *fuzzCtx.JobCtx) {
	job := jobCtx.Job
	outCtx := jobCtx.OutputCtx
	// 生成payload
	for keyword, pm := range job.Preprocess.PlMeta {
		if pm == nil {
			pm = &fuzzTypes.PayloadMeta{}
			job.Preprocess.PlMeta[keyword] = pm
		}
		plList := job.Preprocess.PlMeta[keyword].PlList
		if len(plList) == 0 { // 仅当列表为空时生成，若已有数据则跳过
			pm.PlList = stagePreprocess.PayloadGenerator(job.Preprocess.PlMeta[keyword].Generators, outCtx)
		}
		if job.Preprocess.PlDeduplicate {
			pm.PlList = deduplicatePayloads(pm.PlList)
		}
	}
}

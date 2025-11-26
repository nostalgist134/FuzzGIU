package fuzz

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
)

func genPayloads(jobCtx *fuzzCtx.JobCtx) {
	job := jobCtx.Job
	outCtx := jobCtx.OutputCtx
	// 生成payload
	for keyword, pMeta := range job.Preprocess.PlMeta {
		plList := job.Preprocess.PlMeta[keyword].PlList
		if len(plList) == 0 { // 仅当列表为空时生成，若已有数据则跳过
			pMeta.PlList = stagePreprocess.PayloadGenerator(job.Preprocess.PlMeta[keyword].Generators, outCtx)
		}
	}
}

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
		if len(plList) == 0 { // 若列表已经有数据（比如通过插件手动添加的任务，可手动添加plList），则用原来的数据
			pMeta.PlList = stagePreprocess.PayloadGenerator(job.Preprocess.PlMeta[keyword].Generators, outCtx)
		}
	}
}

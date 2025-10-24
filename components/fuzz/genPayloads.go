package fuzz

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCtx"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/stagePreprocess"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
)

func genPayloads(jobCtx *fuzzCtx.JobCtx) {
	job := jobCtx.Job
	outCtx := jobCtx.OutputCtx
	// 生成payload
	for keyword, _ := range job.Preprocess.PlTemp {
		// 修改生成payload的逻辑：若列表已经有数据（比如通过插件手动添加的任务，可手动添加plList），则用原来的数据
		plList := job.Preprocess.PlTemp[keyword].PlList
		if len(plList) == 0 {
			plList = stagePreprocess.PayloadGenerator(job.Preprocess.PlTemp[keyword].Generators, outCtx)
		}
		job.Preprocess.PlTemp[keyword] = fuzzTypes.PayloadTemp{
			Generators: job.Preprocess.PlTemp[keyword].Generators,
			Processors: job.Preprocess.PlTemp[keyword].Processors,
			PlList:     plList,
		}
	}
}

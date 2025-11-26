package stagePreprocess

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
)

// Preprocess 预处理函数，用来对fuzz任务进行预处理与调整，可自定义
func Preprocess(f *fuzzTypes.Fuzz, outCtx *output.Ctx, priorGen bool) *fuzzTypes.Fuzz {
	preprocessors := f.Preprocess.Preprocessors
	if priorGen {
		preprocessors = f.Preprocess.PreprocPriorGen
	}
	if len(preprocessors) == 0 {
		return f
	}
	// 遍历预处理器链
	for _, p := range preprocessors {
		f = plugin.Preprocess(p, f, outCtx)
	}
	return f
}

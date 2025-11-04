package stagePreprocess

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
)

// Preprocess 预处理函数，用来对fuzz任务进行预处理与调整，可自定义
func Preprocess(fuzz *fuzzTypes.Fuzz, outCtx *output.Ctx) *fuzzTypes.Fuzz {
	newFuzz := fuzz // newFuzz作为预处理器返回的新Fuzz结构
	preprocessors := fuzz.Preprocess.Preprocessors
	if len(preprocessors) > 0 {
		// 遍历预处理器链
		for _, p := range preprocessors {
			newFuzz = plugin.Preprocess(p, newFuzz, outCtx)
		}
	}
	return newFuzz
}

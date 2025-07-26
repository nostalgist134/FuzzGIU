package stagePreprocess

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	plugin2 "github.com/nostalgist134/FuzzGIU/components/plugin"
)

// Preprocess 预处理函数，用来对fuzz模板进行预处理，可自定义
// 除了自定义的逻辑之外，默认的预处理包括生成payload、处理payload、处理json data（如果有）等
func Preprocess(fuzz *fuzzTypes.Fuzz, preprocessors []fuzzTypes.Plugin) *fuzzTypes.Fuzz {
	newFuzz := fuzz // newFuzz作为预处理器返回的新Fuzz结构
	if len(preprocessors) > 0 {
		// 遍历预处理器链
		for _, p := range preprocessors {
			newFuzz = plugin2.PreProcessor(p, fuzz)
		}
	}
	// 生成payload
	for keyword, _ := range newFuzz.Preprocess.PlTemp {
		newFuzz.Preprocess.PlTemp[keyword] = fuzzTypes.PayloadTemp{
			Generators: newFuzz.Preprocess.PlTemp[keyword].Generators,
			Processors: newFuzz.Preprocess.PlTemp[keyword].Processors,
			PlList:     GeneratePayloads(newFuzz.Preprocess.PlTemp[keyword].Generators)}
	}
	return newFuzz
}

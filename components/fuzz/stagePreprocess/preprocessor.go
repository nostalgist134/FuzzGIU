package stagePreprocess

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	plugin2 "github.com/nostalgist134/FuzzGIU/components/plugin"
)

// Preprocess 预处理函数，用来对fuzz任务进行预处理与调整，可自定义
// 除了自定义的逻辑之外，默认的预处理包括生成payload、处理payload、处理json data（如果有）等
func Preprocess(fuzz *fuzzTypes.Fuzz, preprocessors []fuzzTypes.Plugin) *fuzzTypes.Fuzz {
	newFuzz := fuzz // newFuzz作为预处理器返回的新Fuzz结构
	if len(preprocessors) > 0 {
		// 遍历预处理器链
		for _, p := range preprocessors {
			newFuzz = plugin2.Preprocess(p, newFuzz)
		}
	}
	// 生成payload
	for keyword, _ := range newFuzz.Preprocess.PlTemp {
		// 修改生成payload的逻辑：若列表已经有数据（比如通过插件手动添加的任务，可手动添加plList），则用原来的数据
		plList := newFuzz.Preprocess.PlTemp[keyword].PlList
		if len(plList) == 0 {
			plList = PayloadGenerator(newFuzz.Preprocess.PlTemp[keyword].Generators)
		}
		newFuzz.Preprocess.PlTemp[keyword] = fuzzTypes.PayloadTemp{
			Generators: newFuzz.Preprocess.PlTemp[keyword].Generators,
			Processors: newFuzz.Preprocess.PlTemp[keyword].Processors,
			PlList:     plList,
		}
	}
	return newFuzz
}

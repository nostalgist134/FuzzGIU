package stagePreprocess

import (
	"FuzzGIU/components/fuzzTypes"
	plugin2 "FuzzGIU/components/plugin"
)

// Preprocess 预处理函数，用来对fuzz模板进行预处理，可自定义
// 除了自定义的逻辑之外，默认的预处理包括生成payload、处理payload、处理json data（如果有）等
func Preprocess(fuzz *fuzzTypes.Fuzz, preprocessors string) *fuzzTypes.Fuzz {
	newFuzz := fuzz // newFuzz作为预处理器返回的新Fuzz结构
	if preprocessors != "" {
		plugins := plugin2.ParsePluginsStr(preprocessors)
		// 遍历预处理器链
		for _, p := range plugins {
			/*fuzzJson, err := json.Marshal(newFuzz) // 将fuzz类序列化后传入
			if err != nil {
				panic(err)
			}
			ret := plugin2.Call(plugin2.PTypePreProc, p, fuzzJson, nil)*/
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

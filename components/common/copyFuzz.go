package common

import "github.com/nostalgist134/FuzzGIU/components/fuzzTypes"

// CopyFuzz 复制fuzz结构（半深复制，PlList不复制），目前只有react函数中的递归功能使用此函数
func CopyFuzz(f *fuzzTypes.Fuzz) *fuzzTypes.Fuzz {
	if f == nil {
		return nil
	}

	newFuzz := new(fuzzTypes.Fuzz)

	// 拷贝 Preprocess
	newFuzz.Preprocess.Preprocessors = f.Preprocess.Preprocessors
	newFuzz.Preprocess.Mode = f.Preprocess.Mode
	newFuzz.Preprocess.PlTemp = make(map[string]fuzzTypes.PayloadTemp)
	for k, v := range f.Preprocess.PlTemp {
		newFuzz.Preprocess.PlTemp[k] = fuzzTypes.PayloadTemp{
			Generators: v.Generators,
			Processors: v.Processors,
			PlList:     nil, // PlList可以不复制，因为执行doFuzz会重新走一遍生成
		}
	}

	// 拷贝 Send
	newFuzz.Send.Request = f.Send.Request
	newFuzz.Send.Request.HttpSpec.Headers = append([]string{}, f.Send.Request.HttpSpec.Headers...)
	newFuzz.Send.Proxies = append([]string{}, f.Send.Proxies...)
	newFuzz.Send.Retry = f.Send.Retry

	// 拷贝 React
	newFuzz.React.Reactor = f.React.Reactor
	newFuzz.React.Filter = f.React.Filter
	newFuzz.React.Filter.Words = append([]fuzzTypes.Range{}, f.React.Filter.Words...)
	newFuzz.React.Filter.Size = append([]fuzzTypes.Range{}, f.React.Filter.Size...)
	newFuzz.React.Filter.Lines = append([]fuzzTypes.Range{}, f.React.Filter.Lines...)
	newFuzz.React.Filter.Code = append([]fuzzTypes.Range{}, f.React.Filter.Code...)
	newFuzz.React.Matcher = f.React.Matcher
	newFuzz.React.Matcher.Words = append([]fuzzTypes.Range{}, f.React.Matcher.Words...)
	newFuzz.React.Matcher.Size = append([]fuzzTypes.Range{}, f.React.Matcher.Size...)
	newFuzz.React.Matcher.Lines = append([]fuzzTypes.Range{}, f.React.Matcher.Lines...)
	newFuzz.React.Matcher.Code = append([]fuzzTypes.Range{}, f.React.Matcher.Code...)
	newFuzz.React.RecursionControl = f.React.RecursionControl
	newFuzz.React.RecursionControl.StatCodes = append([]fuzzTypes.Range{}, f.React.RecursionControl.StatCodes...)
	newFuzz.React.OutSettings = f.React.OutSettings

	// 拷贝 Misc
	newFuzz.Misc = f.Misc

	return newFuzz
}

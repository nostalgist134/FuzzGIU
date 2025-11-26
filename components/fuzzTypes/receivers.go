package fuzzTypes

import (
	"bytes"
	"time"
)

func cloneSlice[T any](src []T) []T {
	if src == nil {
		return nil
	}
	newSlice := make([]T, len(src))
	copy(newSlice, src)
	return newSlice
}

func clonePlugin(src Plugin) Plugin {
	return Plugin{src.Name, cloneSlice(src.Args)}
}

func clonePlugins(src []Plugin) []Plugin {
	if src == nil {
		return nil
	}
	copied := make([]Plugin, len(src))
	for i, p := range src {
		copied[i] = clonePlugin(p)
	}
	return copied
}

// LiteralClone 克隆Match结构的字面值（会新建Range切片）
func (m Match) LiteralClone() Match {
	m1 := m
	m1.Code = cloneSlice(m.Code)
	m1.Lines = cloneSlice(m.Lines)
	m1.Words = cloneSlice(m.Words)
	m1.Size = cloneSlice(m.Size)
	return m1
}

// Clone 将当前的Fuzz结构克隆一份（但是不克隆payload列表）
func (f *Fuzz) Clone() *Fuzz {
	newFuzz := new(Fuzz)

	// 拷贝 Preprocess
	newFuzz.Preprocess.Preprocessors = clonePlugins(f.Preprocess.Preprocessors)
	newFuzz.Preprocess.PreprocPriorGen = clonePlugins(f.Preprocess.PreprocPriorGen)
	newFuzz.Preprocess.PlMeta = make(map[string]*PayloadMeta)
	for k, v := range f.Preprocess.PlMeta {
		newFuzz.Preprocess.PlMeta[k] = &PayloadMeta{
			Generators: PlGen{v.Generators.Type, clonePlugins(v.Generators.Gen)},
			Processors: clonePlugins(v.Processors),
			// PlList不复制，因为这个部分通常比较大，如果用户确实有需要可自行复制
		}
	}
	newFuzz.Preprocess.ReqTemplate = f.Preprocess.ReqTemplate.LiteralClone()

	// 拷贝 Request
	newFuzz.Request = f.Request
	newFuzz.Request.Proxies = cloneSlice(f.Request.Proxies)

	// 拷贝 React
	newFuzz.React.Reactor = clonePlugin(f.React.Reactor)
	newFuzz.React.Filter = f.React.Filter.LiteralClone()
	newFuzz.React.Matcher = f.React.Matcher.LiteralClone()
	newFuzz.React.RecursionControl = f.React.RecursionControl
	newFuzz.React.RecursionControl.StatCodes = cloneSlice(f.React.RecursionControl.StatCodes)
	newFuzz.Control.OutSetting = f.Control.OutSetting

	// 拷贝 Control
	newFuzz.Control = f.Control
	newFuzz.Control.IterCtrl.Iterator = clonePlugin(f.Control.IterCtrl.Iterator)

	return newFuzz
}

// Clone 克隆Req结构，返回新结构的指针
func (req *Req) Clone() *Req {
	newReq := &Req{}
	*newReq = *req
	newReq.HttpSpec.Headers = cloneSlice(req.HttpSpec.Headers)
	newReq.Fields = cloneSlice(req.Fields)
	newReq.Data = cloneSlice(req.Data)
	return newReq
}

// LiteralClone 克隆Req结构的字面值（重新分配切片）
func (req *Req) LiteralClone() Req {
	literal := *req
	literal.HttpSpec.Headers = cloneSlice(req.HttpSpec.Headers)
	literal.Fields = cloneSlice(req.Fields)
	literal.Data = cloneSlice(req.Data)
	return literal
}

// Clone 克隆RequestCtx结构
func (rc *RequestCtx) Clone() *RequestCtx {
	newReqCtx := new(RequestCtx)
	*newReqCtx = *rc
	if rc.Request != nil {
		newReqCtx.Request = rc.Request.Clone()
	}
	return newReqCtx
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	line := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		line++
	}
	return line
}

func countWords(data []byte) int {
	count := 0
	inWord := false
	for _, b := range data {
		if b == ' ' || b == '\n' || b == '\t' || b == '\r' || b == '\f' || b == '\v' {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

// Statistic 根据rawResponse计算返回包的统计数据（词数、返回包大小、行数）
func (resp *Resp) Statistic() {
	if len(resp.RawResponse) == 0 {
		resp.Lines = 0
		resp.Words = 0
		resp.Size = 0
		return
	}
	resp.Lines = countLines(resp.RawResponse)
	resp.Words = countWords(resp.RawResponse)
	resp.Size = len(resp.RawResponse)
}

// Contains 判断一个值是否在当前Range内
func (r Range) Contains(v int) bool {
	return r.Upper >= v && v >= r.Lower
}

func (ranges Ranges) Contains(v int) bool {
	for _, r1 := range ranges {
		if r1.Contains(v) {
			return true
		}
	}
	return false
}

// Contains 判断一个时间是否在范围内
func (timeBound TimeBound) Contains(t time.Duration) bool {
	return timeBound.Upper > t && t >= timeBound.Lower
}

// Valid 判断时间范围是否有效
func (timeBound TimeBound) Valid() bool {
	return timeBound.Upper > timeBound.Lower
}

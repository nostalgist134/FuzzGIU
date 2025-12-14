package fuzzTypes

import (
	"bytes"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"strings"
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

// MatchResponse 匹配响应包
func (m Match) MatchResponse(resp *Resp) bool {
	if len(m.Size) == 0 && len(m.Words) == 0 && len(m.Code) == 0 && len(m.Lines) == 0 &&
		len(m.Regex) == 0 && m.Time.Upper == m.Time.Lower {
		return false
	}
	whenToRet := false
	if m.Mode == "or" {
		whenToRet = true
	}
	if len(m.Size) != 0 && m.Size.Contains(resp.Size) == whenToRet {
		return whenToRet
	}
	if len(m.Words) != 0 && m.Words.Contains(resp.Words) == whenToRet {
		return whenToRet
	}
	if len(m.Code) != 0 && m.Code.Contains(resp.StatCode) == whenToRet {
		return whenToRet
	}
	if len(m.Lines) != 0 && m.Lines.Contains(resp.Lines) == whenToRet {
		return whenToRet
	}
	if len(m.Regex) != 0 && common.RegexMatch(resp.RawResponse, m.Regex) == whenToRet {
		return whenToRet
	}
	if m.Time.Valid() && m.Time.Contains(resp.ResponseTime) == whenToRet {
		return whenToRet
	}
	return !whenToRet
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
			Generators: PlGen{cloneSlice(v.Generators.Wordlists), clonePlugins(v.Generators.Plugins)},
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

func (f *Fuzz) setPlMetaIfNil() {
	if f.Preprocess.PlMeta == nil {
		f.Preprocess.PlMeta = make(map[string]*PayloadMeta)
	}
}

// WithMinimalExecutable 最小可运行的Fuzz任务，仅fuzz url，单个关键字，状态码匹配200
func (f *Fuzz) WithMinimalExecutable(url, keyword string, payloads []string, sniper bool) *Fuzz {
	f.Preprocess.ReqTemplate.URL = url
	f.setPlMetaIfNil()
	f.Preprocess.PlMeta[keyword] = &PayloadMeta{PlList: payloads}
	if sniper {
		f.Control.IterCtrl.Iterator.Name = "sniper"
	} else {
		f.Control.IterCtrl.Iterator.Name = "clusterbomb"
	}
	return f
}

// WithTemplate 设置任务使用的模板
func (f *Fuzz) WithTemplate(tmpl Req) *Fuzz {
	f.Preprocess.ReqTemplate = tmpl
	return f
}

// WithMatcher 设置匹配器
func (f *Fuzz) WithMatcher(m Match) *Fuzz {
	f.React.Matcher = m.LiteralClone()
	return f
}

// WithFilter 设置过滤器
func (f *Fuzz) WithFilter(filt Match) *Fuzz {
	f.React.Filter = filt.LiteralClone()
	return f
}

// AddKeyword 尝试添加一个fuzz关键字及其payload信息，若已经存在，则返回原有的payload信息与false，不覆盖
// 否则返回添加后的信息与true
func (f *Fuzz) AddKeyword(keyword string, m PayloadMeta) (*PayloadMeta, bool) {
	f.setPlMetaIfNil()
	m1, ok := f.Preprocess.PlMeta[keyword]
	if ok {
		return m1, false
	}
	f.Preprocess.PlMeta[keyword] = &m
	return &m, true
}

// AddKeywordWordlists 添加一个带字典的关键字，或者为已存在的关键字添加字典
func (f *Fuzz) AddKeywordWordlists(keyword string, wordlists []string) (*PayloadMeta, bool) {
	f.setPlMetaIfNil()
	pm, ok := f.Preprocess.PlMeta[keyword]
	if !ok {
		pm = &PayloadMeta{}
		f.Preprocess.PlMeta[keyword] = pm
	}
	pm.Generators.AddWordlists(wordlists)
	return pm, true
}

// AddKeywordPlGenPlugins 添加一个带插件生成器的关键字，或者为已存在的关键字添加字典
func (f *Fuzz) AddKeywordPlGenPlugins(keyword string, plugins []Plugin) (*PayloadMeta, bool) {
	f.setPlMetaIfNil()
	pm, ok := f.Preprocess.PlMeta[keyword]
	if !ok {
		pm = &PayloadMeta{}
		f.Preprocess.PlMeta[keyword] = pm
	}
	pm.Generators.AddPlGenPlugins(plugins)
	return pm, true
}

// AddKeywordPlProc 为关键字添加payload处理器，如果关键字不存在，则返回false
func (f *Fuzz) AddKeywordPlProc(keyword string, proc []Plugin) bool {
	if f.Preprocess.PlMeta == nil {
		return false
	}
	pm, ok := f.Preprocess.PlMeta[keyword]
	if !ok || pm == nil {
		return false
	}
	pm.Processors = append(pm.Processors, proc...)
	return true
}

// AddProxies 添加代理
func (f *Fuzz) AddProxies(proxies []string) {
	f.Request.Proxies = append(f.Request.Proxies, proxies...)
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

// SetRaw 设置rawResponse，并自动计算统计数据
func (resp *Resp) SetRaw(raw []byte) {
	resp.RawResponse = raw
	resp.Statistic()
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

// AddWordlists 为PlGen添加字典生成源
// wordlists为string切片，每个元素可为一个字典或以','分隔的多个字典
func (g *PlGen) AddWordlists(wordlists []string) {
	for _, w1 := range wordlists {
		if strings.Contains(w1, ",") {
			for _, w2 := range strings.Split(w1, ",") {
				g.Wordlists = append(g.Wordlists, strings.TrimSpace(w2))
			}
		} else {
			g.Wordlists = append(g.Wordlists, strings.TrimSpace(w1))
		}
	}
}

// AddPlGenPlugins 为PlGen添加payloadGenerator插件生成源
func (g *PlGen) AddPlGenPlugins(plugins []Plugin) {
	g.Plugins = append(g.Plugins, plugins...)
}

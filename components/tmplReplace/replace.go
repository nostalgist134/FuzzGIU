package tmplReplace

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/resourcePool"
	"github.com/nostalgist134/reusableBytes"
	"unsafe"
)

// todo:
//	Req结构新增了一个Fields字段，为Field struct {
//		Name  string `json:"name"`  // 字段名
//		Value string `json:"value"` // 字段值
//	}的切片，将Fields的解析与替换添加到模板的解析中（已完成，实际上render函
//	数不用改，只要改req2str和最后将fields转为req结构的逻辑就行）
//	修改render系列函数采用写入时分配Lazy结构，写入全部完成后才分配字符串，这
//	是为了解决writeString后底层内存移动，旧字符串指向的底层内存与rb底层不符
//	的问题（已完成）
//	将replace系列函数转为ReplaceTmpl的receiver（已完成）
//	这个文件太大了，既然现在已经单独独立一个包，就分多几个文件（已完成）

type ReplaceTemplate struct {
	fragments    []string
	placeholders []int // placeholders 存储每个片段后关键字在关键字列表的下标列表，特殊情况：下标值为0，代表分隔符
	fieldNum     int
	headerNum    int
}

var bp = new(reusablebytes.BytesPool).Init(128, 131072, 128)

const (
	phSplitter    = 0
	minimumFields = 4
)

func toBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// strings2Req 将切片转化为req结构
func strings2Req(req *fuzzTypes.Req, fields []string, headerNum int) {
	req.HttpSpec.Method = fields[0]
	req.URL = fields[1]
	req.HttpSpec.Proto = fields[2]
	i := 3

	if headerNum != 0 {
		req.HttpSpec.Headers = resourcePool.StringSlices.Get(headerNum)
		for ; i-3 < headerNum; i++ {
			req.HttpSpec.Headers[i-3] = fields[i]
		}
	}

	if reqFieldInFields := len(fields) - headerNum - minimumFields; reqFieldInFields&1 == 0 {
		req.Fields = resourcePool.FieldSlices.Get(reqFieldInFields / 2)
		for j := 0; i < len(fields)-2; i += 2 {
			req.Fields[j].Name = fields[i]
			req.Fields[j].Value = fields[i+1]
			j++
		}
	}
	req.Data = toBytes(fields[len(fields)-1]) // req.Data恒为fields的最后一个项
}

// loadLazyFields 将lazy结构体加载为字符串，同时将lazy切片放回池
func loadLazyFields(fields []string, lazyFields []reusablebytes.Lazy) {
	if len(fields) != len(lazyFields) {
		return
	}
	for i := 0; i < len(lazyFields); i++ {
		fields[i] = lazyFields[i].String()
	}
	lazyPool.Put(lazyFields)
}

// Replace 将模板中的关键字替换为payload列表
func (t *ReplaceTemplate) Replace(payloads []string, sniperPos int) (req *fuzzTypes.Req, cacheId int32) {
	var lazyFields []reusablebytes.Lazy
	if sniperPos >= 0 {
		lazyFields, cacheId = t.renderSniper(payloads[0], sniperPos)
	} else {
		lazyFields, cacheId = t.render(payloads)
	}
	req = resourcePool.GetReq()
	stringFields := resourcePool.StringSlices.Get(len(lazyFields))
	loadLazyFields(stringFields, lazyFields)
	strings2Req(req, stringFields, t.headerNum)
	return
}

// ReplaceTrack 将模板中的单个关键字替换为payload，并返回替换位置信息
func (t *ReplaceTemplate) ReplaceTrack(payload string, sniperPos int) (req *fuzzTypes.Req, track []int,
	cacheId int32) {
	var lazyFields []reusablebytes.Lazy

	if sniperPos >= 0 {
		lazyFields, track, cacheId = t.renderTrackSniper(payload, sniperPos)
	} else {
		lazyFields, track, cacheId = t.renderTrack(payload)
	}
	req = resourcePool.GetReq()
	stringFields := resourcePool.StringSlices.Get(len(lazyFields))
	loadLazyFields(stringFields, lazyFields)
	strings2Req(req, stringFields, t.headerNum)
	return
}

// KeywordCount 根据解析时传入的关键字列表的下标来计算一个关键字在模板中出现的次数
func (t *ReplaceTemplate) KeywordCount(keywordInd int) int {
	cnt := 0
	for _, ph := range t.placeholders {
		if ph == keywordInd+1 { // placeholder的下标0为分隔符，因此要偏移1位
			cnt++
		}
	}
	return cnt
}

func ReleaseReqCache(id int32) {
	bp.Put(id)
}

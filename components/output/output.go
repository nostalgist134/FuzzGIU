package output

import (
	"FuzzGIU/components/fuzzTypes"
	"FuzzGIU/components/output/common"
	fo "FuzzGIU/components/output/fileOutput"
	native "FuzzGIU/components/output/nativeOutput"
	so "FuzzGIU/components/output/screenOutput"
	"encoding/xml"
	"sync"
)

type ObjectOutput struct {
	XMLName  xml.Name        `json:"-" xml:"output"`
	Keywords []string        `json:"keywords" xml:"keywords>keyword"`
	Payloads []string        `json:"payloads" xml:"payloads>payload"`
	Request  *fuzzTypes.Req  `json:"request"  xml:"request"`
	Response *fuzzTypes.Resp `json:"response" xml:"response"`
	Msg      string          `json:"msg,omitempty" xml:"msg,omitempty"`
}

var distinctLogs = sync.Map{}

// InitOutput 初始化输出。此函数不能在多协程调用
func InitOutput(globInfo *fuzzTypes.Fuzz, toWhere int32) {
	// 将输出设置调整为传入的输出设置
	common.GlobOutSettings = &globInfo.React.OutSettings
	if toWhere&OutToFile != 0 {
		fo.InitOutputFile()
	}
	if toWhere&OutToScreen != 0 {
		so.InitOutputScreen(globInfo)
	}
	if toWhere&OutToNativeStdout != 0 {
		native.InitOutputStdout()
	}
}

// ObjOutput 输出单个OutObj
func ObjOutput(obj *ObjectOutput, toWhere int32) {
	objOut := common.OutObj{
		XMLName:  obj.XMLName,
		Keywords: obj.Keywords,
		Payloads: obj.Payloads,
		Request:  obj.Request,
		Response: obj.Response,
		Msg:      obj.Msg,
	}
	if toWhere&OutToFile != 0 {
		fo.FileOutputObj(&objOut)
	}
	if toWhere&OutToScreen != 0 {
		so.ScreenObjOutput(&objOut)
	}
	if toWhere&OutToNativeStdout != 0 {
		native.NativeStdOutput(&objOut)
	}
}

// FinishOutput 在写完文件（当前任务写完之后下一个任务不使用和当前一样的文件）之后调用
func FinishOutput(toWhere int32) {
	if toWhere&OutToFile != 0 {
		fo.FinishOutputFile()
	}
	if toWhere&OutToScreen != 0 {
		so.FinishOutputScreen()
	}
	if toWhere&OutToNativeStdout != 0 {
		native.FinishOutputStdout()
	}
}

func SetTaskCounter(total int64) {
	common.SetTaskCounter(total)
}

func SetJobCounter(total int64) {
	common.SetJobCounter(total)
}

func AddJobCounter() {
	common.AddJobCounter()
}

func AddTaskCounter() {
	common.AddTaskCounter()
}

func ClearTaskCounter() {
	common.ClearTaskCounter()
}

func Log(log string, toWhere int32) {
	if toWhere&OutToNativeStdout != 0 {
		native.Log(log)
	}
	if toWhere&OutToScreen != 0 {
		so.Log(log)
	}
}

func LogOnce(log string, toWhere int32) {
	if _, ok := distinctLogs.LoadOrStore(log, 1); !ok {
		Log(log, toWhere)
	}
	return
}

func GetCounterSingle(which int8) int64 {
	return common.GetCounterSingle(which)
}

func WaitForScreenQuit() {
	so.WaitForScreenQuit()
}

func ScreenClose() {
	so.ScreenClose()
}

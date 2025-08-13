package output

import (
	"encoding/xml"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/common"
	fo "github.com/nostalgist134/FuzzGIU/components/output/fileOutput"
	native "github.com/nostalgist134/FuzzGIU/components/output/nativeOutput"
	so "github.com/nostalgist134/FuzzGIU/components/output/screenOutput"
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

var pendingLogs = make([]string, 0)

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
	if len(pendingLogs) > 0 {
		for _, log := range pendingLogs {
			Log(toWhere, log)
		}
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

// SetTaskCounter 设置task总数
func SetTaskCounter(total int64) {
	common.SetTaskCounter(total)
}

// SetJobCounter 设置job总数
func SetJobCounter(total int64) {
	common.SetJobCounter(total)
}

// AddJobCounter job数加1
func AddJobCounter() {
	common.AddJobCounter()
}

// AddTaskCounter task数加1
func AddTaskCounter() {
	common.AddTaskCounter()
}

// ClearTaskCounter task数清0（不是总数）
func ClearTaskCounter() {
	common.ClearTaskCounter()
}

// Log 记录日志，支持标准输出或termui，暂不支持文件
func Log(toWhere int32, log string) {
	if toWhere&OutToNativeStdout != 0 {
		native.Log(log)
	}
	if toWhere&OutToScreen != 0 {
		so.Log(log)
	}
}

// Logf 格式化输出日志
func Logf(toWhere int32, fmtLog string, a ...any) {
	Log(toWhere, fmt.Sprintf(fmtLog, a...))
}

// PendLog 在init函数调用前记录日志，在init完成后一次性推送，若已经调用了init函数，则不能再使用此函数输出日志
func PendLog(log string) {
	if log != "" {
		pendingLogs = append(pendingLogs, log)
	}
}

// LogOnce 只记录一次相同的日志
func LogOnce(log string, toWhere int32) {
	if _, ok := distinctLogs.LoadOrStore(log, 1); !ok {
		Log(toWhere, log)
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

func UpdateScreenInfoPage(newInfo *fuzzTypes.Fuzz) {
	so.UpdateGlobInfo(newInfo)
}

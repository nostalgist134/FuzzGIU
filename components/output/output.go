package output

import (
	"encoding/json"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/common"
	fo "github.com/nostalgist134/FuzzGIU/components/output/fileOutput"
	mem "github.com/nostalgist134/FuzzGIU/components/output/memOutput"
	native "github.com/nostalgist134/FuzzGIU/components/output/nativeOutput"
	so "github.com/nostalgist134/FuzzGIU/components/output/screenOutput"
	"sync"
)

/*
output 包用于处理FuzzGIU的输出结果，主要提供3类函数：InitOutput、Output和Finish。
	InitOutput函数需要在单协程环境，且output函数不被调用时中调用，不会自动检查协程安全问题
	Output类函数用于将结果输出到指定文件或输出流，需要在多协程的环境下调用，因此内部检查协程安全
	Finish类函数用于结束输出，具体结束的结果取决于内部实现，这个函数同样需要在单协程环境，且output函数不被调用时中调用
*/

type OutObj = common.OutObj

var distinctLogs = sync.Map{}

var pendingLogs = make([]string, 0)

// InitOutput 初始化输出。此函数不能在多协程调用
func InitOutput(globInfo *fuzzTypes.Fuzz, toWhere int32) {
	// 将输出设置调整为传入的输出设置
	common.GlobOutSettings = &globInfo.React.OutSettings
	if toWhere&OutToFile != 0 {
		fo.InitOutput()
	}
	if toWhere&OutToScreen != 0 {
		so.InitOutput(globInfo)
	}
	if toWhere&OutToNativeStdout != 0 {
		native.InitOutput()
	}
	if toWhere&OutToMem != 0 {
		mem.InitOutput()
	}
	if len(pendingLogs) > 0 {
		for _, log := range pendingLogs {
			Log(toWhere, log)
		}
	}
}

// Output 输出单个OutObj
func Output(obj *OutObj, toWhere int32) {
	if toWhere&OutToFile != 0 {
		fo.Output(obj)
	}
	if toWhere&OutToScreen != 0 {
		so.Output(obj)
	}
	if toWhere&OutToNativeStdout != 0 {
		native.Output(obj)
	}
	if toWhere&OutToMem != 0 {
		mem.Output(obj)
	}
}

// FinishOutput 完成输出，重置状态
func FinishOutput(toWhere int32) {
	if toWhere&OutToFile != 0 {
		fo.FinishOutput()
	}
	if toWhere&OutToScreen != 0 {
		so.FinishOutput()
	}
	if toWhere&OutToNativeStdout != 0 {
		native.FinishOutput()
	}
}

// SetTaskTotal 设置task总数
func SetTaskTotal(total int64) {
	common.SetTaskTotal(total)
}

// SetJobTotal 设置job总数
func SetJobTotal(total int64) {
	common.SetJobTotal(total)
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

// PendLog 在init函数调用前记录日志，在init完成后一次性推送，若已经调用了init函数，则不能再使用此函数输出日志，且此函数不能多协程调用
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

func GetCounterValue(which int8) int64 {
	return common.GetCounterValue(which)
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

func GetMemOutObjects(start int, end int) []json.RawMessage {
	return mem.GetObjects(start, end)
}

func GetAllMemOutObjects() []json.RawMessage {
	return mem.GetAllObjects()
}

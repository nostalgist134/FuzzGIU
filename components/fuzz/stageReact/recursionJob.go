package stageReact

import (
	"bytes"
	"github.com/nostalgist134/FuzzGIU/components/common"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
)

// insertRecursionMarker 往请求中的指定位置插入递归关键字，便于之后递归中使用
// 递归关键字需要插入的位置recursionPos在模板渲染时获取，recursionPos按照如下逻辑解析：
// 一个recursionPos中可能含有正数或者负数，标记了一个字段中需要插入递归关键字的位置或字段的结束。
// 若recursionPos[i]为正数，则说明这个是要插入payload的下标，并且当前段还没结束；
// 若recursionPos[i]为负数，但是绝对值<=len(field)，则其正数代表要插入的下标，并且负号代表当前段结束
// 若recursionPos[i]为负数，且绝对值大于len(field)，则说明当前字段没有要插入递归关键字的位置
func insertRecursionMarker(recKeyword string, splitter string,
	field string, recursionPos []int, currentPos int) (string, int) {
	sb := strings.Builder{}
	ind := 0
	for ; recursionPos[currentPos] > 0; currentPos++ {
		sb.WriteString(field[ind:recursionPos[currentPos]])
		sb.WriteString(splitter)
		sb.WriteString(recKeyword)
		ind = recursionPos[currentPos]
	}
	if -recursionPos[currentPos] <= len(field) {
		sb.WriteString(field[ind:-recursionPos[currentPos]])
		sb.WriteString(splitter)
		sb.WriteString(recKeyword)

		if ind = -recursionPos[currentPos]; ind < len(field) {
			sb.WriteString(field[ind:])
		}
	} else {
		return field, currentPos + 1
	}
	return sb.String(), currentPos + 1
}

func insertLastRecursionMarker(recKeyword string, splitter string, dataField []byte, recursionPos []int,
	currentPos int) []byte {
	buf := bytes.Buffer{}
	ind := 0
	for ; recursionPos[currentPos] > 0; currentPos++ {
		buf.Write(dataField[ind:recursionPos[currentPos]])
		buf.WriteString(splitter)
		buf.WriteString(recKeyword)
	}
	if -recursionPos[currentPos] <= len(dataField) {
		buf.Write(dataField[ind:-recursionPos[currentPos]])
		buf.WriteString(splitter)
		buf.WriteString(recKeyword)
		if ind = -recursionPos[currentPos]; ind < len(dataField) {
			buf.Write(dataField[ind:])
		}
	} else {
		return dataField
	}
	return buf.Bytes()
}

/*
METHOD		URL		HTTP_VERSION
HEADERS
FIELDS

DATA
*/

// deriveRecursionJob 生成递归任务
func deriveRecursionJob(job *fuzzTypes.Fuzz, reqSend *fuzzTypes.Req, recPos []int) *fuzzTypes.Fuzz {
	recCtrl := &job.React.RecursionControl
	recKw := recCtrl.Keyword
	recSp := recCtrl.Splitter

	recJob := common.CopyFuzz(job)
	recJob.Control.IterCtrl.Iterator.Name = ""

	recJob.React.RecursionControl.RecursionDepth++
	recJob.Preprocess.ReqTemplate = *reqSend

	derived := &recJob.Preprocess.ReqTemplate
	curPos := 0

	// METHOD
	derived.HttpSpec.Method, curPos = insertRecursionMarker(recKw, recSp,
		derived.HttpSpec.Method, recPos, curPos)

	// URL
	derived.URL, curPos = insertRecursionMarker(recKw, recSp,
		derived.URL, recPos, curPos)

	// VERSION
	derived.HttpSpec.Version, curPos = insertRecursionMarker(recKw, recSp,
		derived.HttpSpec.Version, recPos, curPos)

	// HEADERS
	for i := 0; i < len(derived.HttpSpec.Headers); i++ {
		derived.HttpSpec.Headers[i], curPos = insertRecursionMarker(recKw, recSp,
			derived.HttpSpec.Headers[i], recPos, curPos)
	}

	// FIELDS
	for i := 0; i < len(derived.Fields); i++ {
		derived.Fields[i].Name, curPos = insertRecursionMarker(recKw, recSp,
			derived.Fields[i].Name, recPos, curPos)
		derived.Fields[i].Value, curPos = insertRecursionMarker(recKw, recSp,
			derived.Fields[i].Value, recPos, curPos)
	}

	// DATA
	derived.Data = insertLastRecursionMarker(recKw, recSp, derived.Data, recPos, curPos)

	return recJob
}

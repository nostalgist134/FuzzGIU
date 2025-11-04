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
	/*
		首先给出一些trackPos的示例，假设我们有一个字符串
		abcdefFUZZghiFUZZjklmn
		然后在使用ReplaceTrack替换，将FUZZ关键字替换为PAYLOAD时，它会变成
		abcdefPAYLOADghiPAYLOADjklmn
		对应的trackPos段为：[..., 13, -23 ,...]
		我们可以看到，13实际上是ghi段g的下标，然后-23的绝对值23是jklmn中j的下标，这个函数采用的是“前插法”，也就是
		它会在trackPos对应的下标前面插入，那么假设递归分隔符为'/'，递归关键字为'FUZZ'，它插入递归关键字后的字符串
		就会变为：
		abcdefPAYLOAD/FUZZghiPAYLOAD/FUZZjklmn
		同时也会自动把下标更新（在这个示例中是加2）

		边界情况：trackPos的绝对值==字符串/字节的长度，这样的字符串其结尾有payload插入，比如abcdefPAYLOAD，由于
		切片允许使用切片的最大长度作为切割下标，因此这种情况也不会panic，而是会在len(s)之前插入，也是正常的，而模板
		替换的trackPos产生过程无法被用户干预，因此也不会产生trackPos==字符串/字节长度的情况
		若trackPos的绝对值为字符串长度加1，且为负数，则代表当前字段没有插入点
	*/
	for ; recursionPos[currentPos] > 0; currentPos++ { // 遍历直到到达最后一个插入点
		// 使用“前插法”
		sb.WriteString(field[ind:recursionPos[currentPos]])
		sb.WriteString(splitter)
		sb.WriteString(recKeyword)

		// 结合上面的示例，这个就是payload插入点的后一位，也就是...PAYLOADghi...的g的下标，从这里开始写
		ind = recursionPos[currentPos]
	}

	// 出现负数，这里是最后一个插入点/当前字段没有插入点
	if -recursionPos[currentPos] <= len(field) {
		sb.WriteString(field[ind:-recursionPos[currentPos]])
		sb.WriteString(splitter)
		sb.WriteString(recKeyword)

		// 不要忘了将当前字段剩余的部分也写入
		if ind = -recursionPos[currentPos]; ind < len(field) {
			sb.WriteString(field[ind:])
		}
	} else { // 没有插入点，直接返回
		return field, currentPos + 1
	}
	return sb.String(), currentPos + 1
}

// insertRecursionMarkerBytes 这个函数实际上和上面的一样，只不过把strings.Builder换成bytes.Buffer而已，由于Data是最后一个字段
// 也就不需要更新currentPos了
func insertRecursionMarkerBytes(recKeyword string, splitter string, dataField []byte, recursionPos []int,
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
	derived.Data = insertRecursionMarkerBytes(recKw, recSp, derived.Data, recPos, curPos)

	return recJob
}

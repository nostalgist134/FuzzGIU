package output

import (
	"github.com/nostalgist134/FuzzGIU/components/output/common"
	"io"
	"os"
)

// closureFileOutput 在格式化输出到文件的前后加上“文件头”与“文件闭合”，确保格式能够正常解析
func closureFileOutput(format string, end bool) {
	xmlClosure := "<outputs>"
	jsonClosure := `[`
	if end {
		xmlClosure = "</outputs>"
		jsonClosure = "]"
	}
	if currentFile != nil {
		switch format {
		case "json":
			// 如果有json输出，则将json加入的多余逗号去掉（通过将文件指针回移一位）
			if !outputObjectEmpty {
				_, err := currentFile.Seek(-1, io.SeekCurrent)
				if err != nil {
					panic(err)
				}
			}
			currentFile.WriteString(jsonClosure)
		case "xml":
			currentFile.WriteString(xmlClosure)
		}
	}
}

func InitOutput() {
	outputHasInit = true
	// 如果新的fuzz任务使用的和旧的是同一个文件名，则初始化时不再次更新文件指针和文件名，从而可以继续追加写入，直到调用FinishOutput
	if currentFileName == "" || currentFileName != common.GlobOutSettings.OutputFile {
		var err error
		currentFile, err = os.OpenFile(common.GlobOutSettings.OutputFile,
			os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
		if err != nil {
			panic(err)
		}
		currentFileName = common.GlobOutSettings.OutputFile
		closureFileOutput(common.GlobOutSettings.OutputFormat, false)
	}
}

func Output(obj *common.OutObj) {
	if !outputHasInit {
		return
	}
	muFile.Lock()
	defer muFile.Unlock()
	objOut := common.FormatObjOutput(obj, common.GlobOutSettings.OutputFormat, false)
	currentFile.Write(objOut)
	switch common.GlobOutSettings.OutputFormat {
	case "json":
		// 如果是json格式，写入逗号（由于这个函数会在多协程中调用，所以没办法知道是不是最后一个，因此也无法在此处判断是否应该写入逗号）
		currentFile.Write([]byte{','})
		// 写入逗号后标记为当前输出体有写入，方便结束输出时删除多余的逗号
		outputObjectEmpty = false
	case "json-lines": // json-lines仅需写入换行
		currentFile.Write([]byte{'\n'})
	}
}

func FinishOutput() {
	if !outputHasInit {
		return
	}
	closureFileOutput(common.GlobOutSettings.OutputFormat, true)
	outputObjectEmpty = true
	currentFile.Close()
	currentFileName = ""
}

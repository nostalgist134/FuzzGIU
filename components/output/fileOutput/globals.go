package output

import (
	"os"
	"sync"
)

var muFile = sync.Mutex{}

var outputObjectEmpty = true

// currentFileName 当前写入文件的文件名
var currentFileName = ""

// currentFile 当前写入文件的文件指针
var currentFile *os.File

// outputHasInit 标记是否已经初始化
var outputHasInit = false

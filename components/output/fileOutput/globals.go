package output

import (
	"os"
	"sync"
)

var muFile = sync.Mutex{}

var outputObjectEmpty = true

// currentOutputFileName 当前写入文件的文件名
var currentOutputFileName = ""

// currentFileOutput 当前写入文件的文件指针
var currentFileOutput *os.File

// outputHasInit 标记是否已经初始化
var outputHasInit = false

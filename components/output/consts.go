package output

const (
	// OutToFile 输出到文件
	OutToFile = 1
	// OutToScreen 输出到屏幕（termui）
	OutToScreen = 2
	// OutToNativeStdout 直接输出到stdout
	OutToNativeStdout = 4

	// CntTask 获取task个数
	CntTask = 0
	// TotalTask 获取task总数
	TotalTask = 1
	// CntJob 获取job个数
	CntJob = 2
	// TotalJob 获取job总数
	TotalJob = 3
)

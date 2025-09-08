package output

const (
	OutToFile         = 1 // OutToFile 输出到文件
	OutToScreen       = 2 // OutToScreen 输出到屏幕（termui）
	OutToNativeStdout = 4 // OutToNativeStdout 直接输出到stdout
	OutToDB           = 8 // OutToDB 输出到db文件中（这个暂时不做了，因为我发现gorm库打包完编译出来的文件直接会到30m以上）

	// CntTask 获取task个数
	CntTask = 0
	// TotalTask 获取task总数
	TotalTask = 1
	// CntJob 获取job个数
	CntJob = 2
	// TotalJob 获取job总数
	TotalJob = 3
)

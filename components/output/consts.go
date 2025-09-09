package output

const (
	OutToFile         = 0x1  // OutToFile 输出到文件
	OutToScreen       = 0x2  // OutToScreen 输出到屏幕（termui）
	OutToNativeStdout = 0x4  // OutToNativeStdout 直接输出到stdout
	OutToDB           = 0x8  // OutToDB 输出到db文件中（这个暂时不做了，因为我发现gorm库打包完编译出来的文件直接会到30m以上）
	OutToMem          = 0x10 // OutToMem 输出到内存缓冲区中（这个标志位不直接通过命令行参数或job成员设置，仅在被动模式下提供给api接口使用）

	// CntTask 获取task个数
	CntTask = 0
	// TotalTask 获取task总数
	TotalTask = 1
	// CntJob 获取job个数
	CntJob = 2
	// TotalJob 获取job总数
	TotalJob = 3
)

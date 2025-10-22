package outputFlag

const (
	OutToFile   = 1 << iota // OutToFile 输出到文件
	OutToTview              // OutToTview 输出到屏幕（termui）
	OutToStdout             // OutToStdout 直接输出到stdout
	OutToDB                 // OutToDB 输出到db文件中（这个暂时不做了，因为我发现gorm库打包完编译出来的文件直接会到30m以上）
	OutToChan               // OutToChan 输出到内存缓冲区中（这个标志位不直接通过命令行参数或job成员设置，仅在被动模式下提供给api接口使用）
)

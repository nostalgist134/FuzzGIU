package output

import "github.com/nostalgist134/FuzzGIU/components/output/outputFlag"

const (
	OutToFile   = outputFlag.OutToFile   // OutToFile 输出到文件
	OutToTview  = outputFlag.OutToTview  // OutToTview 输出到tview
	OutToStdout = outputFlag.OutToStdout // OutToStdout 直接输出到stdout
	OutToDB     = outputFlag.OutToDB     // OutToDB 输出到db文件中（这个暂时不做了，因为我发现gorm库打包完编译出来的文件直接会到30m以上）
	OutToChan   = outputFlag.OutToChan   // OutToChan 输出到管道缓冲区中（这个标志位不直接通过命令行参数或job成员设置，仅在被动模式下提供给api接口使用）
)

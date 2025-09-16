package main

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz"
	"github.com/nostalgist134/FuzzGIU/components/options"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"os"
)

// initEnv 初始化函数，目前功能仅有创建插件目录
func initEnv() {
	dirs := []string{plugin.BaseDir, plugin.BaseDir + plugin.RelPathPlGen, plugin.BaseDir + plugin.RelPathPlProc,
		plugin.BaseDir + plugin.RelPathPreprocessor, plugin.BaseDir + plugin.RelPathReqSender,
		plugin.BaseDir + plugin.RelPathReactor}
	for _, dir := range dirs {
		fmt.Printf("Checking directory %s......", dir)
		// 如果目录不存在，则尝试创建
		if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
			err = os.Mkdir(dir, 0755)
			if err != nil {
				fmt.Printf("We have a problem with creating directory %s: %s\n", dir, err.Error())
			}
			fmt.Println("Created.")
			continue
		}
		fmt.Println("exist.")
	}
}

func RunDirect(opt *options.Opt) {
	fuzz1 := opt2fuzz(opt)
	fuzz.JQ.AddJob(fuzz1)
	fuzz.DoJobs()
}

func main() {
	opt := options.ParseOptCmdline()
	if len(os.Args) == 1 {
		fmt.Println("Checking/initializing environment...")
		initEnv()
		fmt.Println("Done.")
		fmt.Println("For help, use -h flag")
		return
	}
	if opt.General.Passive {
		RunPassive(opt)
		return
	}
	RunDirect(opt)
}

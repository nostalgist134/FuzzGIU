package main

import (
	"FuzzGIU/components/fuzz"
	"FuzzGIU/components/options"
	"FuzzGIU/components/plugin"
	"fmt"
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

func main() {
	opts := options.ParseOptCmdline()
	if len(os.Args) == 1 {
		fmt.Println("Checking/initializing environment...")
		initEnv()
		fmt.Println("Done.")
		fmt.Println("For help, use -h flag")
		return
	}
	fuzz1 := opt2fuzz(opts)
	debug()
	/*fuzz.Debug(fuzz1)
	return*/
	fuzz.JQ.AddJob(fuzz1)
	fuzz.DoJobs()
}

func debug() {
	pluginStr := "pay(14214,\"wa     2223\",false,3.6,0x114514),zwa(0xff,true,'6663ffa'),qw"
	p := plugin.ParsePluginsStr(pluginStr)
	fmt.Printf("%v\n", p)
	fmt.Printf("%v\n", p[0].Args[0].(int))
	os.Exit(0)
}

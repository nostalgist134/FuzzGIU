package main

import (
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/opt"
	"github.com/nostalgist134/FuzzGIU/components/output/tviewOutput"
	"github.com/nostalgist134/FuzzGIU/components/plugin"
	"github.com/nostalgist134/FuzzGIU/libfgiu"
	"log"
	"os"
)

// initEnv 初始化函数，目前功能仅有创建插件目录
func initEnv() {
	dirs := []string{
		plugin.BaseDir,
		plugin.BaseDir + plugin.RelPathPlGen,
		plugin.BaseDir + plugin.RelPathPlProc,
		plugin.BaseDir + plugin.RelPathPreprocessor,
		plugin.BaseDir + plugin.RelPathRequester,
		plugin.BaseDir + plugin.RelPathReactor,
		plugin.BaseDir + plugin.RelPathIterator,
	}

	for _, dir := range dirs {
		fmt.Printf("Checking directory %s...", dir)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err = os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Failed to create directory %s: %v\n", dir, err)
			} else {
				fmt.Println("Created.")
			}
			continue
		}

		// 存在但不是目录（比如是个文件）
		if info, err := os.Stat(dir); err == nil && !info.IsDir() {
			fmt.Printf("Warning: %s exists but is not a directory\n", dir)
		} else {
			fmt.Println("exists.")
		}
	}
}

func main() {
	o := opt.ParseOptCmdline()

	if len(os.Args) == 1 { // 无任何启动参数，则初始化环境后退出
		fmt.Println("checking/initializing environment...")
		initEnv()
		fmt.Println("done.")
		fmt.Println("for help, use -h flag")
		return
	}

	var (
		fuzzer *libfgiu.Fuzzer
		err    error
	)

	if o.ApiConfig.HttpApi { // api模式运行
		webApiCfg := libfgiu.WebApiConfig{
			ServAddr:     o.ApiConfig.ApiAddr,
			TLS:          o.ApiConfig.ApiTLS,
			CertFileName: o.ApiConfig.TLSCertFile,
			CertKeyName:  o.ApiConfig.TLSKeyFile,
		}
		fuzzer, err = libfgiu.NewFuzzer(20, webApiCfg)
		if err != nil {
			log.Fatalf("failed to create fuzzer: %v\n", err)
		}
		fuzzer.Start()
		fmt.Printf("listening at %s\n", webApiCfg.ServAddr)
		fmt.Println("access token:", fuzzer.GetApiToken())
		fuzzer.Wait()
	} else { // 普通模式运行
		var j *fuzzTypes.Fuzz
		j, err = libfgiu.Opt2fuzz(o)
		if err != nil {
			log.Fatalf("failed to create job: %v\n", err)
		}
		fuzzer, err = libfgiu.NewFuzzer(20)
		if err != nil {
			log.Fatalf("failed to create fuzzer: %v\n", err)
		}
		err = fuzzer.Start()
		if err != nil {
			log.Fatalf("failed to start fuzzer: %v\n", err)
		}
		_, err = fuzzer.Submit(j)
		if err != nil {
			log.Fatalf("failed to execute fuzz: %v\n", err)
		}
		fuzzer.Wait()
		err = fuzzer.Stop()
		tviewOutput.QuitTview()
		if err != nil {
			log.Fatalf("failed to stop fuzzer: %v\n", err)
		}
	}
}

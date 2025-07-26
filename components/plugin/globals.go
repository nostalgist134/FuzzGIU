package plugin

import (
	"runtime"
)

const (
	pluginEntry         = "PluginWrapper"
	BaseDir             = "./plugins/"
	RelPathPlGen        = "payloadGenerators/"
	RelPathPlProc       = "payloadProcessors/"
	RelPathPreprocessor = "preprocessors/"
	RelPathReqSender    = "requestSenders/"
	RelPathReactor      = "reactors/"
)

var binSuffix = ""

func init() {
	operSys := runtime.GOOS
	switch operSys {
	case "windows":
		binSuffix = ".dll"
	case "linux":
		binSuffix = ".so"
	case "darwin":
		binSuffix = ".dylib"
	}
}

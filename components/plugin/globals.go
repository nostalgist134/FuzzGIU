package plugin

import (
	"runtime"
)

type Plugin struct {
	Name string
	Args []any
}

const (
	pluginEntry         = "PluginWrapper"
	BaseDir             = "./plugins/"
	RelPathPlGen        = "payloadGenerators/"
	RelPathPlProc       = "payloadProcessors/"
	RelPathPreprocessor = "preprocessors/"
	RelPathReqSender    = "requestSenders/"
	RelPathReactor      = "reactors/"
)

const (
	PTypePlGen     = "payloadGenerator"
	PTypePreProc   = "preprocessor"
	PTypePlProc    = "payloadProcessor"
	PTypeReactor   = "reactor"
	PTypeReqSender = "reqSender"
)

var binSuffix = ""

func init() {
	operSys := runtime.GOOS
	switch operSys {
	case "Windows":
		binSuffix = ".dll"
	case "Linux":
		binSuffix = ".so"
	case "Darwin":
		binSuffix = ".dylib"
	}
}

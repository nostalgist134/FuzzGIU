package plugin

import (
	"runtime"
)

const (
	pluginEntry = "PluginWrapper"

	RelPathPlGen        = "payloadGenerators/"
	RelPathPlProc       = "payloadProcessors/"
	RelPathPreprocessor = "preprocessors/"
	RelPathRequester    = "requesters/"
	RelPathReactor      = "reactors/"
	RelPathIterator     = "iterators/"

	SelectIterIndex = 0
	SelectIterLen   = 1
)

var BaseDir = "./plugins/"
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

package plugin

type Plugin struct {
	Name string
	Args []any
}

const (
	suffix              = ".dll"
	pluginEntry         = "PluginWrapper"
	pluginBase          = "./plugins/"
	relPathPlGen        = "payloadGenerators/"
	relPathPlProc       = "payloadProcessors/"
	relPathPreprocessor = "preprocessors/"
	relPathReqSender    = "requestSenders/"
	relPathReactor      = "reactors/"
)

const (
	PTypePlGen     = "payloadGenerator"
	PTypePreProc   = "preprocessor"
	PTypePlProc    = "payloadProcessor"
	PTypeReactor   = "reactor"
	PTypeReqSender = "reqSender"
)

package outputable

var fmts = map[string]bool{
	"json":      true,
	"json-line": true,
	"xml":       true,
	"native":    true,
}

func FormatSupported(format string) bool {
	if v, ok := fmts[format]; ok && v {
		return true
	}
	return false
}

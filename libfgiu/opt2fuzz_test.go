package libfgiu

import (
	"fmt"
	"testing"
)

func TestParseWordlistArg(t *testing.T) {
	w, k := parseWordlistArg("test.txt,H:\\test.txt,C:\\nishigiu\\woshigiu.txt::FUZZ")
	fmt.Println(w)
	fmt.Println(k)
}

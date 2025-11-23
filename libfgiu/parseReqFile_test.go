package libfgiu

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestParseHttpReq(t *testing.T) {
	fName := "C:/Users/patrick/Desktop/req.txt"
	f, err := toHttpRequest(fName)
	b, _ := json.MarshalIndent(f, "", "  ")
	fmt.Println(string(b))
	fmt.Println(string(f.Data))
	fmt.Println(err)
}

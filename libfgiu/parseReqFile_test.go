package libfgiu

import (
	"fmt"
	"testing"
)

func TestParseHttpReq(t *testing.T) {
	fName := "C:/Users/patrick/Desktop/req.txt"
	req, raw, err := parseRequestFile(fName)
	fmt.Println(req)
	fmt.Println(string(raw))
	fmt.Println(req.HttpSpec.Proto)
	fmt.Println(err)
}

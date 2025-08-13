package inputHandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nostalgist134/FuzzGIU/components/fuzz/fuzzCommon"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"strings"
)

var errAddJobFail = errors.New("failed to add job")

func stopJob([]string, []byte) (any, error) {
	return nil, fuzzCommon.ErrJobStop
}

func addJob(_ []string, jobJson []byte) (any, error) {
	newJob := new(fuzzTypes.Fuzz)
	err := json.Unmarshal(jobJson, newJob)
	if err != nil {
		return nil, err
	}
	if !fuzzCommon.AddJob(newJob) {
		return nil, errAddJobFail
	}
	return bytesOk, nil
}

// job 内含两个子命令：stop和add
func job(args []string, data []byte) (any, error) {
	switch strings.ToLower(args[0]) {
	case "stop":
		return stopJob(args, data)
	case "addJob":
		return addJob(args, data)
	}
	return nil, fmt.Errorf("unknown operation '%s' over job", args[0])
}

package fileOutput

import (
	"os"
	"sync"
)

type Ctx struct {
	f               *os.File
	fLog            *os.File
	mu              *sync.Mutex
	muLog           *sync.Mutex
	outputFmt       string
	fileDir         string
	outputVerbosity int
	outputEmpty     bool
	closed          bool
}

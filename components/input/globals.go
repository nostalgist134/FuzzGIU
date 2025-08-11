package input

import (
	"sync"
)

type Input struct {
	cmd  string
	args []string
	data []byte
}

type controlQueue struct {
	list       []Input
	listCursor int
	mu         sync.Mutex
}

var globCq = new(controlQueue)

func init() {
	globCq.list = make([]Input, 0)
	globCq.listCursor = -1
}

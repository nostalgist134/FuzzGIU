package input

import (
	"net"
	"sync"
)

type Input struct {
	Cmd  string
	Args []string
	Data []byte
	Peer net.Conn
}

type inputStack struct {
	list   []*Input
	cursor int
	mu     sync.Mutex
}

var inputStk = new(inputStack)

var Enabled bool

func init() {
	inputStk.init(64)
}

package chanOutput

import "errors"

var (
	errChanFull          = errors.New("output channel is already full")
	errLogNotImplemented = errors.New("log function not implemented")
)

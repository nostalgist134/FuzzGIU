package chanOutput

import "errors"

var (
	errChanFull          = errors.New("output channel is already full")
	errChanClosed        = errors.New("channel is already closed")
	errLogNotImplemented = errors.New("log function not implemented")
)

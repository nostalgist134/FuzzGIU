package outputErrors

import (
	"errors"
)

var (
	ErrTviewConflict      = errors.New("use of tview and stdout output simultaneously causes conflict")
	ErrRegisterNilCounter = errors.New("try to register a nil counter")
	ErrCtxClosed          = errors.New("output context is already closed")
)

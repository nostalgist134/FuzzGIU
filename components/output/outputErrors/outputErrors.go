package outputErrors

import (
	"errors"
)

var (
	ErrTviewConflict      = errors.New("use of tview and stdout output simultaneously causes conflict")
	ErrRegisterNilCounter = errors.New("try to register a nil counter")
	ErrNilJobCtx          = errors.New("nil job context to be used")
	ErrNilOutputSetting   = errors.New("nil output setting to be used")
	ErrCtxClosed          = errors.New("output context is already closed")
)

package stdoutOutput

import "errors"

var (
	errTviewConflict = errors.New("use tview and stdout output simultaneously causes conflict")
	errCtxClosed     = errors.New("stdout fuzzCtx is already closed")
	errChanNil       = errors.New("can't register counter because register chan is nil")
	errRegisterNil   = errors.New("try to register a nil counter")
)

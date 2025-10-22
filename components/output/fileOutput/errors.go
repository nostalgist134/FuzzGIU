package fileOutput

import (
	"errors"
)

var (
	errEmptyFName     = errors.New("empty file name")
	errCtxClosed      = errors.New("file output fuzzCtx is already closed")
	errFileOutCtxNil  = errors.New("file output fuzzCtx is nil")
	errFilePointerNil = errors.New("file pointer is nil")
)

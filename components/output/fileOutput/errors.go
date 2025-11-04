package fileOutput

import (
	"errors"
)

var (
	errEmptyFName     = errors.New("empty file name")
	errFileOutCtxNil  = errors.New("file output context is nil")
	errFilePointerNil = errors.New("file pointer is nil")
)

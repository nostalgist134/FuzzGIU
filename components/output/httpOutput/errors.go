package httpOutput

import "errors"

var (
	errNilURLToPost = errors.New("try to post to a nil url")
	errNilHttpCli   = errors.New("the http client of this context is nil")
)

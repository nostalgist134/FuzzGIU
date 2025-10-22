package httpOutput

import (
	"net/http"
	"net/url"
)

type Ctx struct {
	url url.URL
	cli http.Client
}

func InitOutput() {

}

package httpOutput

import (
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"net/http"
	"net/url"
)

type Ctx struct {
	u   *url.URL
	cli *http.Client
}

func NewHttpOutputCtx(outSetting *fuzzTypes.OutputSetting, _ int) (*Ctx, error) {
	u, err := url.Parse(outSetting.HttpURL)
	if err != nil {
		return nil, err
	}

}

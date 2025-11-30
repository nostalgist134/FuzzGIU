package httpOutput

import (
	"bytes"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputErrors"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"net/http"
	"net/url"
)

type Ctx struct {
	u      *url.URL
	cli    *http.Client
	closed bool
}

func NewHttpOutputCtx(outSetting *fuzzTypes.OutputSetting, _ int) (*Ctx, error) {
	u, err := url.Parse(outSetting.HttpURL)
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	return &Ctx{
		u:   u,
		cli: cli,
	}, nil
}

func (c *Ctx) Output(obj *outputable.OutObj) error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	if c.cli == nil {
		return errNilHttpCli
	}
	if c.u == nil {
		return errNilURLToPost
	}

	objBytes := obj.ToFormatBytes("json", false, 0)
	_, err := c.cli.Post(c.u.String(), "application/json", bytes.NewReader(objBytes))
	return err
}

func (c *Ctx) Close() error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	c.cli.CloseIdleConnections()
	c.closed = true
	c.cli = nil
	return nil
}

func (c *Ctx) Log(log *outputable.Log) error {
	if c.closed {
		return outputErrors.ErrCtxClosed
	}
	if c.cli == nil {
		return errNilHttpCli
	}
	if c.u == nil {
		return errNilURLToPost
	}

	logBytes := log.ToFormatBytes("json")
	_, err := c.cli.Post(c.u.String(), "application/json", bytes.NewReader(logBytes))
	return err
}

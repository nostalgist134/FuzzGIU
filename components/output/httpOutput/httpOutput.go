package httpOutput

import (
	"bytes"
	"errors"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"github.com/nostalgist134/FuzzGIU/components/output/outputable"
	"net/http"
	"net/url"
)

type Ctx struct {
	u      *url.URL
	cli    *http.Client
	closed bool
}

var errClosed = errors.New("http output ctx is already closed")

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
	objBytes := obj.ToFormatBytes("json", false, 0)
	_, err := c.cli.Post(c.u.String(), "application/json", bytes.NewReader(objBytes))
	return err
}

func (c *Ctx) Close() error {
	if c.closed {
		return errClosed
	}
	c.cli.CloseIdleConnections()
	c.closed = true
	return nil
}

func (c *Ctx) Log(log *outputable.Log) error {
	logBytes := log.ToFormatBytes("json")
	_, err := c.cli.Post(c.u.String(), "application/json", bytes.NewReader(logBytes))
	return err
}

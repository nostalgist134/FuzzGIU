package stageSend

import (
	"bytes"
	"context"
	"github.com/nostalgist134/FuzzGIU/components/fuzzTypes"
	"net"
	"net/url"
	"time"
)

func resolveSubdomain(subdomain string, timeout int) (bool, []string, error) {

	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{Timeout: time.Duration(timeout) * time.Second}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	ips, err := resolver.LookupHost(ctx, subdomain)
	if err != nil {
		return false, nil, err
	}
	return true, ips, nil
}

func sendRequestDns(req *fuzzTypes.Req, timeout int) *fuzzTypes.Resp {
	rawResp := bytes.NewBuffer(nil)
	resp := new(fuzzTypes.Resp)
	URL, _ := url.Parse(req.URL)
	_, ips, err := resolveSubdomain(URL.Host, timeout)
	if err != nil {
		resp.ErrMsg = err.Error()
	}
	for _, ip := range ips {
		rawResp.WriteString(ip + "\n")
	}
	resp.RawResponse = rawResp.Bytes()
	return resp
}

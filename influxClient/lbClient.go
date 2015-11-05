package influxClient

import (
	"bytes"
	"net/http"
	"net/url"
	"time"

	"github.com/qiniu/rpc.v2"
	"github.com/qiniu/rpc.v3/lb"
	"qiniu.com/auth/authstub.v1"
	"qiniu.com/auth/proto.v1"
)

const (
	DefaultDialTimeout = time.Second * 10
	DefaultRespTimeout = time.Second * 10
)

type Client struct {
	Client *lb.Client
}

func NewClient(hosts []string) (*Client, error) {
	c, err := NewLbClient(hosts)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: c,
	}, nil
}

func NewLbClient(hosts []string) (lbclient *lb.Client, err error) {

	var t http.RoundTripper
	tc := &rpc.TransportConfig{
		DialTimeout:           DefaultDialTimeout,
		ResponseHeaderTimeout: DefaultRespTimeout,
	}
	t = rpc.NewTransport(tc)

	si := &proto.SudoerInfo{}
	t = authstub.NewTransport(si, t)

	lbConfig := &lb.Config{
		Http:              &http.Client{Transport: t},
		FailRetryInterval: 0,
		TryTimes:          1,
	}

	lbclient, err = lb.New(hosts, lbConfig)
	if err != nil {
		return nil, err
	}

	return
}

func (c *Client) Write(points []byte) (err error) {

	err = c.Client.CallWith(nil, nil, "POST", "/write?db=testDB", "text/plain", bytes.NewBuffer(points), len(points))
	if err != nil && err.Error() != "No Content" {
		return
	}
	return nil
}

func (c *Client) Query(sql string) (ret map[string]interface{}, err error) {
	err = c.Client.Call(nil, &ret, "GET", "/query?db=testDB&q="+url.QueryEscape(sql))
	if err != nil {
		return nil, err
	}
	return
}

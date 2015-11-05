package influxClient

import (
	"fmt"
)

type Point struct {
	Data string
}

type InfluxClient struct {
	WriteReqCount     uint64
	QueryReqCount     uint64
	WriteReqFailCount uint64
	QueryReqFailCount uint64

	Pause chan struct{}

	Client interface {
		Query(sql string) (ret map[string]interface{}, err error)
		Write(points []byte) (err error)
	}
}

func NewInfluxClient() *InfluxClient {
	return &InfluxClient{
		WriteReqCount:     0,
		QueryReqCount:     0,
		WriteReqFailCount: 0,
		QueryReqFailCount: 0,
		Pause:             make(chan struct{}),
	}
}

func (c *InfluxClient) WritePoints(points []byte) error {

	return c.Client.Write(points)
}

func (c *InfluxClient) Query(sql string) (ret map[string]interface{}, err error) {

	return c.Client.Query(sql)
}

func (c *InfluxClient) Stat() (ret map[string]interface{}, err error) {
	fmt.Printf("writeReq:%d,queryReq:%d writeFailReq:%d,queryFailReq:%d \n",
		c.WriteReqCount, c.QueryReqCount, c.WriteReqFailCount, c.QueryReqFailCount)
	return
}

func (c *InfluxClient) Action(t string) (err error) {
	//t may be: pause write points
	return
}

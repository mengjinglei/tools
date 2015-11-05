package influxClient

import (
	"time"

	"github.com/qiniu/log.v1"
)

func Main() {
	hosts := []string{"http://127.0.0.1:9086", "http://127.0.0.1:10086", "http://127.0.0.1:11086"}
	lbCliet, err := NewClient(hosts)
	if err != nil {
		log.Error(err)
		return
	}
	inClient := NewInfluxClient()
	inClient.Client = lbCliet

	_, err = inClient.Query("create database testDB")
	if err != nil {
		log.Error(err)
		return
	}
	for {
		points := Point{
			Data: "obj,color=red,sharp=rect value=0.64",
		}
		err = inClient.WritePoints([]byte(points.Data))
		if err != nil {
			log.Error(err)
			inClient.WriteReqFailCount++
		}
		inClient.WriteReqCount++
		time.Sleep(time.Millisecond * 200)
		if (inClient.WriteReqCount+inClient.WriteReqFailCount)%100 == 0 {
			inClient.Stat()
		}
	}

}

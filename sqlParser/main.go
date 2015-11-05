package main

import (
	"strings"
	"time"

	"github.com/qiniu/log.v1"
	"qbox.us/errors"
	"qiniu.com/pandora/influxql.v0.9.0"
)

const offsetSeconds = 8 * 60 * 60

func main() {
	sql := "select value from xxx where host='server01' group by  host, region"

	st, err := influxql.NewParser(strings.NewReader(sql)).ParseStatement()
	if err != nil {
		err = errors.Info(err).Detail(err)
		log.Debug(err)
		return
	}

	//ensure the sql statement is a select statement
	stmt, ok := st.(*influxql.SelectStatement)
	if !ok {
		log.Info("not a valid select statement")
		return
	}

	//get min and max of sql
	log.Info(stmt.Condition)
	min, max := influxql.TimeRange(stmt.Condition)
	log.Printf("Min:%v, Max:%v", min, max)

	log.Printf("Min:%d, Max:%d", min.Unix()+offsetSeconds, max.Unix()+offsetSeconds)

	loc, err := time.LoadLocation("")

	acientTime := time.Date(1970, 1, 1, 0, 0, 0, 0, loc)
	log.Println(acientTime.Unix(), acientTime.UnixNano(), acientTime)
	log.Info("...", min.Unix(), max.Unix(), acientTime.Unix())
	start, end := true, true
	if err != nil {
		return
	}

	if !min.Round(time.Microsecond).Equal(acientTime.Round(time.Microsecond)) {
		log.Println("<<<<<<<<<<<<")
		start = true
	}
	if !max.Equal(acientTime) {
		end = true
	}

	stmt.SetPartialTimeRange(start, end, min.Add(time.Second*offsetSeconds), max.Add(time.Second*offsetSeconds))

	log.Println(stmt.String())

	//get group by
	log.Println(stmt.Dimensions)

	//host := "127.0.0.1:9092"
	log.Println(time.Now().UTC())

	log.Println(".............................")
	testTimeRange()
}

func testTimeRange() {
	sql := "select vlaue from req where time > now() - 1h"

	st, err := influxql.NewParser(strings.NewReader(sql)).ParseStatement()
	if err != nil {
		err = errors.Info(err).Detail(err)
		log.Debug(err)
		return
	}

	//ensure the sql statement is a select statement
	stmt, ok := st.(*influxql.SelectStatement)
	if !ok {
		log.Info("not a valid select statement")
		return
	}

	//get min and max of sql
	log.Info(stmt.Condition)
	min, max := influxql.TimeRange(stmt.Condition)
	log.Printf("Min:%v, Max:%v", min, max)

	log.Printf("Min:%d, Max:%d", min.Unix()+offsetSeconds, max.Unix()+offsetSeconds)

}

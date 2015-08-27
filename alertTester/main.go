package main

import (
	"fmt"
	"io/ioutil"

	"bosun.org/_third_party/github.com/influxdb/client"

	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "bosun.org/cmd/bosun/conf"

	"github.com/qiniu/log.v1"
)

const (
	InfluxHost = "127.0.0.1"
	InfluxPort = 8086
	MyDB       = "bosunTestDB"
	RepoLimit  = 10000
)

func main() {
	srv := New()
	log.Debug("new success")
	go srv.run()
	log.Debug("start run")
	go collect(srv.colls)
	log.Debug("start collect")
	<-srv.done
}

func (s *Service) run() {

	repoids := make([]string, 0)
	for i := 0; i < RepoLimit; i++ {
		repoid := "repoid_" + strconv.FormatInt(int64(i), 10)

		//insert repoid into repo collection
		s.colls.Repo.Insert(M{"id": repoid})

		//write data point into influxdb
		repoids = append(repoids, repoid)
	}
	log.Println(repoids)
	go writes(s.influxClient, repoids)
	log.Debug("write points")
	go NewAlert(repoids)
	<-s.done
}

func NewAlert(repoids []string) {

	req := "req"
	//creat notification
	for _, repoid := range repoids {
		defaultNotification := fmt.Sprintf("\n notification default {\n get = http://127.0.0.1:8800/notify?id=%s&alert=%s\n next = default\n timeout = 1m\n}", repoid, repoid+"_"+req)
		url := fmt.Sprintf("http://127.0.0.1:9090/v1/repos/%s/notifications", repoid)
		log.Info(defaultNotification)
		resp, err := post("create notification default", url, []byte(defaultNotification))
		if err != nil {
			log.Debug(resp, err)
			return
		}
	}

	//create alerts
	for pos, repoid := range repoids {
		alert := fmt.Sprintf("lookup req{\n entry code=* {\n high=1\n} } \n alert %s_req {\ncrit=max(influx(\"%s\",\"select value from %s group by code,host\",\"8h10m\",\"0m\",\"code,host\")) > 2 \ncritNotification=default \n template=default\n}", repoid, MyDB, repoids[pos%300])
		url := fmt.Sprintf("http://127.0.0.1:9090/v1/repos/%s/alert", repoid)
		resp, err := post("create alert req", url, []byte(alert))
		if err != nil {
			log.Debug(resp, err)
			return
		}
		log.Debug("create alert for ", repoid)
		time.Sleep(time.Second)
	}
}

func writes(c *client.Client, repoids []string) {
	codes := []int{200}
	hosts := []int{1}
	for {
		if len(repoids) > 300 {
			repoids = repoids[:300]
		}
		for _, repoid := range repoids {
			point := fmt.Sprintf("%s,code=%d,host=%d value=%d", repoid, codes[rand.Intn(1)], hosts[rand.Intn(1)], rand.Intn(10)+1)
			resp, err := writePoints(c, []string{point})
			if err != nil {
				log.Debug(resp, err)
			}

		}
		time.Sleep(time.Millisecond * 300)
	}
}

func NewInfluxClient() *client.Client {
	u, err := url.Parse(fmt.Sprintf("http://%s:%d", InfluxHost, InfluxPort))
	log.Debug("influx host", u)
	if err != nil {
		log.Fatal(err)
	}

	conf := client.Config{
		URL:      *u,
		Username: os.Getenv("INFLUX_USER"),
		Password: os.Getenv("INFLUX_PWD"),
	}

	con, err := client.NewClient(conf)
	if err != nil {
		log.Fatal(err)
	}

	dur, ver, err := con.Ping()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Create new client success! %v, %s", dur, ver)

	ret, err := queryDB(con, fmt.Sprintf("drop database %s", MyDB))
	if err != nil {
		log.Info(err)
	}
	fmt.Println(ret)

	ret, err = queryDB(con, fmt.Sprintf("create database %s", MyDB))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ret)

	return con
}

func queryDB(con *client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: MyDB,
	}
	if response, err := con.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	}
	return
}

func writePoints(con *client.Client, data []string) (ret string, err error) {

	bps := client.BatchPoints{
		TextPoints:      data,
		Database:        MyDB,
		RetentionPolicy: "default",
	}
	_, err = con.Write(bps)
	if err != nil {
		log.Fatal(err)
		ret = "write point failse"
	}

	return
}

func post(cmd, url string, dat []byte) (ret []byte, err error) {

	client := &http.Client{}
	log.Info(">>>>>>> "+cmd, "url", url)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(dat)))
	if err != nil {
		log.Error(err)
	}

	req.Header.Set("Authorization", "QiniuStub uid=1&ut=4")
	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}

	defer resp.Body.Close()

	_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}
	log.Println(resp.StatusCode, resp.Status, string(_bytes))

	ret = _bytes
	return
}

func get(cmd, action, url string) (ret []byte, err error) {

	client := &http.Client{}
	log.Info(">>>>>>> "+cmd, "url", url)
	req, err := http.NewRequest(action, url, nil)
	if err != nil {
		log.Error(err)
	}

	req.Header.Set("Authorization", "QiniuStub uid=1&ut=4")

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err)
	}

	defer resp.Body.Close()

	_bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
	}
	log.Println(resp.StatusCode, resp.Status, string(_bytes))

	ret = _bytes
	return
}

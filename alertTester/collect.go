package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"bosun.org/_third_party/github.com/influxdb/client"
	"github.com/qiniu/db/mgoutil.v3"
	"github.com/qiniu/log.v1"
	"qbox.us/cc/config"
)

type TesterConfig struct {
	Port      int    `json:"port"`
	MongoHost string `json:"mongo"`
}
type collection struct {
	Notify mgoutil.Collection `coll:"notify"`
	Repo   mgoutil.Collection `coll:"repo"`
}
type Service struct {
	colls        collection
	influxClient *client.Client
	done         chan bool
}
type web struct {
	colls collection
}

type M map[string]interface{}

func (service *web) notifyHandler(w http.ResponseWriter, r *http.Request) {

	log.Info("recieve req url:", r.RequestURI)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	log.Info("req data:", string(data))

	//get repoid and alert name
	repoid := r.FormValue("id")
	alert := r.FormValue("alert")

	if repoid == "" || alert == "" {
		log.Debug("repoid or alert is empty")
		fmt.Fprintf(w, "repoid or alert is empty")
		return
	}

	//insert repoid, alert, and time
	err = service.colls.Notify.Insert(M{"id": repoid, "alert": alert, "time": time.Now().Unix()})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprint(w, "ok")
}

func (service *web) checkHandler(w http.ResponseWriter, r *http.Request) {

}

func New() *Service {

	config.Init("f", "tester", "collect.conf")
	var conf TesterConfig
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
		return nil
	}

	//init mongo
	var colls collection
	_, err := mgoutil.Open(&colls, &mgoutil.Config{Host: "127.0.0.1", DB: "pandora_test"})
	if err != nil {
		log.Fatal("open mongo fail:", err)
	}
	c := NewInfluxClient()

	//init service
	srv := &Service{colls: colls, influxClient: c, done: make(chan bool)}
	return srv
}

func collect(c collection) {
	w := &web{colls: c}

	http.HandleFunc("/notify", w.notifyHandler)
	http.HandleFunc("/check", w.checkHandler)
	err := http.ListenAndServe(":"+strconv.Itoa(8800), nil)
	if err != nil {
		log.Error(err)
	}
}

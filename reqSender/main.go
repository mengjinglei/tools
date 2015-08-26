package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/qiniu/log.v1"
	"qbox.us/cc/config"
)

type Proxy struct {
	Url         string `json:"url"`
	Port        int    `json:"port"`
	CrossDomain bool   `json:"cross_domain"`
	DebugLevel  int    `json:"debug_level"`
}

func (proxy *Proxy) sendHandler(w http.ResponseWriter, r *http.Request) {
	index, _ := template.ParseFiles("index.html")
	index.Execute(w, nil)
}

func (proxy *Proxy) staticHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[1:])
}

func (proxy *Proxy) handler(w http.ResponseWriter, r *http.Request) {

	if proxy.CrossDomain {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization,Origin, X-Requested-With, Content-Type, Accept")
	}

	log.Info("recieve req url:", r.RequestURI)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	log.Info("req header:", r.Header)
	log.Info("req data:", string(data))

	url := proxy.Url + r.RequestURI
	req, err := http.NewRequest(r.Method, url, bytes.NewBuffer(data))
	if err != nil {
		log.Error(err)
	}

	log.Println(url)
	copyHeader(req.Header, r.Header)

	client := &http.Client{}
	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return
	}

	ret, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	log.Printf(string(ret))
	log.Println(".............", resp.StatusCode)
	fmt.Fprint(w, string(ret))

}

//copy header
func copyHeader(dst, src http.Header) {

	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func main() {

	config.Init("f", "proxy", "proxy.conf")
	var proxy Proxy
	if err := config.Load(&proxy); err != nil {
		log.Fatal("config.Load failed:", err)
		return
	}
	log.SetOutputLevel(proxy.DebugLevel)

	http.HandleFunc("/send/", proxy.sendHandler)
	http.HandleFunc("/static/", proxy.staticHandler)
	http.HandleFunc("/", proxy.handler)
	err := http.ListenAndServe(":"+strconv.Itoa(proxy.Port), nil)
	if err != nil {
		log.Error(err)
	}
}

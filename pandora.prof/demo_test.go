package main

import (
	"bufio"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qiniu/http/rpcutil.v1"
)

func TestHandleHi_Recorder(t *testing.T) {
	rw := httptest.NewRecorder()
	handleHi(rw, req(t, "GET / HTTP/1.0\r\n\r\n"))
	if !strings.Contains(rw.Body.String(), "visitor number") {
		t.Errorf("Unexpected output: %s", rw.Body)
	}
}

func req(t *testing.T, v string) *http.Request {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(v)))
	if err != nil {
		t.Fatal(err)
	}
	return req
}

func TestHandleHi_TestServer(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleHi))
	defer ts.Close()
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Error(err)
		return
	}
	if g, w := res.Header.Get("Content-Type"), "text/html; charset=utf-8"; g != w {
		t.Errorf("Content-Type = %q; want %q", g, w)
	}
	slurp, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Got: %s", slurp)
}

func BenchmarkHi(b *testing.B) {
	b.ReportAllocs()

	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader("GET / HTTP/1.0\r\n\r\n")))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		rw := httptest.NewRecorder()
		handleHi(rw, req)
	}
}

func BenchmarkWritePoints(b *testing.B) {
	b.ReportAllocs()

	points := ""
	for i := 0; i < 2048; i++ {
		points += "cputest,host=1,region=hz value=0.64"
		if i != 2047 {
			points += "\n"
		}
	}

	req, err := http.NewRequest("POST", "localhost:8000", bufio.NewReader(strings.NewReader(points)))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rw := httptest.NewRecorder()
		env := &rpcutil.Env{
			W:   rw,
			Req: req,
		}
		args := &cmdArgs{CmdArgs: []string{"appid", "repoid"}}
		WritePoints(args, env)
	}
}

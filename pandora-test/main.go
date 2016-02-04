package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/qiniu/http/restrpc.v1"
	"github.com/qiniu/http/rpcutil.v1"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/rpc.v3"
	"github.com/qiniu/rpc.v3/lb"
	"qbox.us/cc/config"
	influxql "qiniu.com/pandora/influxql.v0.9.0"

	"qiniu.com/auth/authstub.v1"
	"qiniu.com/auth/proto.v1"
)

type Config struct {
	InfluxdbHost []string `json:"influxdb_hosts"`
	PandoraHost  []string `json:"pandora_hosts"`
	MgrHost      []string `json:"mgr_hosts"`
	Port         string   `json:"port"`
}

type Service struct {
	Config
	InfluxdbClient *lb.Client
	PandoraClient  *lb.Client
	MgrClient      *lb.Client
}

type cmdArgs struct {
	CmdArgs []string
}

func isWhiteSpace(r rune) bool {
	return r == ' '
}

func New(conf Config) (p *Service, err error) {

	pandoraClient, err := newLbClient(conf.PandoraHost)
	if err != nil {
		return nil, err
	}
	influxdbClient, err := newLbClient(conf.InfluxdbHost)
	if err != nil {
		return nil, err
	}
	mgrClient, err := newLbClient(conf.MgrHost)
	if err != nil {
		return nil, err
	}
	p = &Service{
		Config:         conf,
		PandoraClient:  pandoraClient,
		InfluxdbClient: influxdbClient,
		MgrClient:      mgrClient,
	}

	return
}

func main() {

	config.Init("f", "pandora", "pandora-test.conf")

	var conf Config
	if err := config.Load(&conf); err != nil {
		log.Fatal("config.Load failed:", err)
		return
	}

	svr, err := New(conf)

	router := restrpc.Router{
		Factory:       restrpc.Factory,
		PatternPrefix: "/v4",
		Mux:           restrpc.NewServeMux(),
	}

	err = http.ListenAndServe(conf.Port, router.Register(svr))

	if err != nil {
		log.Fatal(err)
	}

}

func (s *Service) PostApps_Repos_Points(args *cmdArgs, env *rpcutil.Env) (err error) {

	appid := args.CmdArgs[0]
	repoid := args.CmdArgs[1]
	data, err := ioutil.ReadAll(env.Req.Body)
	if err != nil {
		return err
	}

	reqSql := strings.TrimLeftFunc(string(data), isWhiteSpace)
	point := strings.TrimRightFunc(reqSql, isWhiteSpace)

	wurl := fmt.Sprintf("/v4/apps/%s/repos/%s/points", appid, repoid)
	err = s.PandoraClient.CallWith(nil, nil, "POST", wurl, "text/plain", bytes.NewBuffer([]byte(point)), len(point))
	if err != nil {
		return fmt.Errorf("write pandora fail, error:%v", err)
	}

	err = s.InfluxdbClient.CallWith(nil, nil, "POST", "/write?db=pandora&rp=default", "text/plain", bytes.NewBuffer([]byte(point)), len(point))
	if err != nil && err.Error() != "No Content" {
		return fmt.Errorf("write influxdb fail, error:%v", err)
	}

	return
}

func (s *Service) PostApps_Repos_Series_(args *cmdArgs, env *rpcutil.Env) (err error) {

	appid := args.CmdArgs[0]
	repoid := args.CmdArgs[1]
	seires := args.CmdArgs[2]

	data, err := ioutil.ReadAll(env.Req.Body)
	if err != nil {
		return
	}

	wurl := fmt.Sprintf("/v4/apps/%s/repos/%s/series/%s", appid, repoid, seires)
	err = s.MgrClient.Call(nil, nil, "DELETE", wurl)
	if err != nil {
		return fmt.Errorf("delete series %v fail, error:%v", seires, err)
	}

	wurl = fmt.Sprintf("/v4/apps/%s/repos/%s/series/%s", appid, repoid, seires)
	err = s.MgrClient.CallWith(nil, nil, "POST", wurl, "text/plain", bytes.NewBuffer(data), len(data))
	if err != nil {
		return fmt.Errorf("create series %v fail, error:%v", seires, err)
	}
	sql := "drop measurement " + seires
	err = s.InfluxdbClient.Call(nil, nil, "GET", "/query?db=pandora&q="+url.QueryEscape(sql))
	if err != nil {
		return fmt.Errorf("drop measurement fail, error:%v", err)
	}

	return
}

type QueryReq struct {
	Sql string `json:"sql"`
}

func (s *Service) PostApps_Repos_Query(args *cmdArgs, env *rpcutil.Env) (ret map[string]interface{}, err error) {

	appid := args.CmdArgs[0]
	repoid := args.CmdArgs[1]

	data, err := ioutil.ReadAll(env.Req.Body)
	if err != nil {
		return
	}

	var reqSql string = ""
	contentType := env.Req.Header.Get("Content-Type")
	if contentType == "text/plain" || contentType == "application/text" {
		//get sql
		reqSql = string(data)
	} else {
		//check the req body
		reqBody := make(map[string]interface{})
		err = json.Unmarshal(data, &reqBody)
		if err != nil {
			return
		}
		if reqBody["sql"] == nil {
			return
		}

		//get sql
		sql, ok := reqBody["sql"].(string)
		if !ok {
			return
		}
		reqSql = sql
	}

	req := &QueryReq{}
	req.Sql = reqSql

	stmt, err := influxql.NewParser(strings.NewReader(req.Sql)).ParseStatement()
	if err != nil {
		return
	}

	st, ok := stmt.(*influxql.SelectStatement)
	if !ok {
		return nil, fmt.Errorf("not valid select statement")
	}

	err = rewriteTimeConditions(st)
	if err != nil {
		return
	}

	req.Sql = st.String()

	qurl := fmt.Sprintf("/v4/apps/%v/repos/%s/query", appid, repoid)
	pandoraRet := make(map[string]interface{})
	err = s.PandoraClient.CallWithJson(nil, &pandoraRet, "POST", qurl, req)
	if err != nil {
		return nil, fmt.Errorf("query pandora fail %v", err)
	}

	influxdbRet := make(map[string]interface{})
	err = s.InfluxdbClient.Call(nil, &influxdbRet, "GET", "/query?db=pandora&q="+url.QueryEscape(req.Sql))
	if err != nil {
		return nil, fmt.Errorf("query influxdb fail %v", err)
	}
	ret = make(map[string]interface{})
	ret["result"] = "success"
	if !reflect.DeepEqual(influxdbRet, pandoraRet) {
		ret["result"] = "fail"
		ret["influxdb"] = influxdbRet
		ret["pandora"] = pandoraRet
		return
	}
	ret = influxdbRet
	return
}

func newLbClient(hosts []string) (lbclient *lb.Client, err error) {

	dialTimeout, err := time.ParseDuration("30s")
	if err != nil {
		return nil, errors.New("invalid portal dial_timeout")
	}
	respTimeout, err := time.ParseDuration("30s")
	if err != nil {
		return nil, errors.New("invalid portal resp_timeout")
	}

	var t http.RoundTripper
	tc := &rpc.TransportConfig{
		DialTimeout:           dialTimeout,
		ResponseHeaderTimeout: respTimeout,
	}
	t = rpc.NewTransport(tc)

	si := &proto.SudoerInfo{
		UserInfo: proto.UserInfo{
			Uid:   1,
			Utype: 4,
		},
	}
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

func rewriteTimeConditions(stmt *influxql.SelectStatement) (err error) {
	//condition: time > now() - 2h and time < now() - 1h and host='server01'
	if stmt.Condition == nil {
		return nil
	}

	exprs := influxql.ExtractNowValue(stmt.Condition)
	if len(exprs) < 1 || len(exprs) > 2 {
		return fmt.Errorf("too many time conditions")
	}

	for _, expr := range exprs {
		expr, ok := expr.(*influxql.BinaryExpr)
		if !ok {
			return fmt.Errorf("not valid time condition")
		}
		switch rhs := expr.RHS.(type) {
		case *influxql.BinaryExpr:
			//now() - 2h
			ilhs, ok := rhs.LHS.(*influxql.Call)
			if !ok {
				return fmt.Errorf("not valid time condition")
			}
			if ilhs.Name != "now" {
				return fmt.Errorf("not valid time condition, only now() function supported")
			}
			irhs, ok := rhs.RHS.(*influxql.DurationLiteral)
			if !ok {
				return fmt.Errorf("not valid time condition, not valid time duration")
			}
			if rhs.Op == influxql.SUB {
				expr.RHS = &influxql.TimeLiteral{Val: time.Now().Add(-irhs.Val)}
			} else if rhs.Op == influxql.ADD {
				expr.RHS = &influxql.TimeLiteral{Val: time.Now().Add(-irhs.Val)}
			} else {
				return fmt.Errorf("not valid time condition, not valid operator")
			}

		case *influxql.Call:
			//now()
			if rhs.Name != "now" {
				return fmt.Errorf("not valid time condition, only now() function supported")
			}
			expr.RHS = &influxql.TimeLiteral{Val: time.Now()}
		case *influxql.TimeLiteral, *influxql.NumberLiteral:
			//'2016-02-03 03:45:02.068164'
			//1454471474754245632
			//do nothing
		default:
			//do nothing
		}

	}
	_, err = stmt.RewriteWithoutTimeDimensions()
	if err != nil {
		return err
	}

	var cond string
	if stmt.Condition != nil {
		cond = stmt.Condition.String() + " and "
	}

	for i, expr := range exprs {
		cond += fmt.Sprintf("%v", expr.String())
		if i != len(exprs)-1 {
			cond += " and "
		}
	}

	e, err := influxql.ParseExpr(cond)
	if err != nil {
		return nil
	}
	stmt.Condition = e

	return nil
}

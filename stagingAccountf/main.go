package main

import (
	"strconv"

	"github.com/qiniu/db/mgoutil.v3"
	"github.com/qiniu/log.v1"

	adminAcc "qbox.us/admin_api/account.v2"
	"qbox.us/cc/config"
	"qbox.us/qconf/qconfapi"
	"qiniu.com/auth/account.v1"
	"qiniu.com/pandora/pandora.v4/common"
)

type AdminAcc struct {
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Collections struct {
	Export mgoutil.Collection `coll:"export"`
}

type Config struct {
	Qconf qconfapi.Config `json:"qconfg"`
	Admin AdminAcc        `json:"admin"`
	UID   uint32          `json:"uid"`
	AK    string          `json:"ak"`
	M     mgoutil.Config  `json:"mgo"`
}

type ExportConfig struct {
	AppId      string                 `json:"appId"  bson:"appId"`
	RepoName   string                 `json:"repo"   bson:"repo"`
	ExportName string                 `json:"name"   bson:"name"`
	Type       string                 `json:"type"   bson:"type"`
	Spec       map[string]interface{} `json:"spec"   bson:"spec"`
	Whence     string                 `json:"whence" bson:"whence"`
}

type M map[string]interface{}

func main() {

	config.Init("f", "exportd", "exportd-default.conf")

	var cfg Config
	if err := config.Load(&cfg); err != nil {
		log.Fatal("load config failed:", err)
	}

	var colls Collections
	_, err := mgoutil.Open(&colls, &cfg.M)
	if err != nil {
		log.Fatal("Open MongoDB failed: ", err)
		return
	}

	acc := account.New(&account.Config{Qconfg: cfg.Qconf})
	accessInfo, err := acc.GetAccessInfo(nil, cfg.AK)
	if err != nil {
		log.Error(err)
		return
	}

	if accessInfo.Uid != cfg.UID {
		log.Errorf("ak:%v does not match uid:%v,%v", cfg.AK, cfg.UID, accessInfo.Uid)
	} else {
		log.Infof("success, ak:%v match uid:%v", cfg.AK, cfg.UID)
	}

	adAcc, err := adminAcc.NewService(cfg.Admin.Host, "", "", cfg.Admin.Username, cfg.Admin.Password)
	if err != nil {
		return
	}

	iter := colls.Export.Find(M{}).Select(M{}).Iter()
	defer common.SafeCloseIter(iter)
	export := ExportConfig{}
	for iter.Next(&export) {
		//get uid to uint32
		uid, err := strconv.ParseInt(export.AppId, 10, 32)
		if err != nil {
			log.Error(err)
			continue
		}
		info, err := adAcc.UserInfoByUid(uint32(uid), nil)
		if err != nil {
			return
		}
		_, err = colls.Export.Upsert(M{"appId": export.AppId, "name": export.ExportName, "repo": export.RepoName}, M{"$set": M{"email": info.Email}})
		if err != nil {
			return
		}
	}
}

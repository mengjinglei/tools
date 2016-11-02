package main

import (
	"log"

	sdkbase "qiniu.com/pandora/base"
	sdk "qiniu.com/pandora/pipeline"
)

func main() {
	ak := "uz5NdxgSYR-bdRvP6BH58kW913PItx60UOhm7son" // 替换成自己的AK/SK
	sk := "p2XHGrGzHV0kYAf4cEKhe72fpLVZ2VzqXa_Zs4-y"

	// 生成配置文件
	cfg := sdk.NewConfig().
		WithAccessKeySecretKey(ak, sk).
		WithEndpoint("https://pipeline.qiniu.com").
		WithLogger(sdkbase.NewDefaultLogger()).
		WithLoggerLevel(sdkbase.LogDebug)

	// 生成client实例
	client, err := sdk.New(cfg)
	if err != nil {
		log.Println(err)
		return
	}

	schema := []sdk.RepoSchemaEntry{sdk.RepoSchemaEntry{Key: "testKey", ValueType: "string", Required: true}}
	err = client.CreateRepo(&sdk.CreateRepoInput{RepoName: "testRepo", Region: "nb", Schema: schema}) // 创建repo
	if err != nil {
		log.Println(err)
		return
	}

	repos, err := client.ListRepos(&sdk.ListReposInput{}) // 列举repo
	if err != nil {
		log.Println(err)

		return
	}
	log.Println(repos)
}

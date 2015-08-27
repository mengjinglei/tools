package main

import (
	"bosun.org/cmd/bosun/conf/parse"
	"github.com/qiniu/log.v1"
)

func main() {
	config := []byte(`notification name {
		email = hello@qiniu.com
		get = 127.0.0.1
		}`)
	tree, err := parse.Parse("text", string(config))
	if err != nil {
		return
	}
	log.Println(tree)
	log.Println(tree.Name, tree.Root.String(), tree.Root.Nodes, tree.Root.NodeType, tree.Parse(string(config)))
}

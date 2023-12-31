package main

import (
	"fmt"

	"github.com/seveye/goms/rpc"
	"github.com/seveye/goms/util"
	"github.com/seveye/goms/watch"
	"github.com/seveye/goms/watch_config"
)

func main() {
	util.GosLogInit("db0", "./", true, 0)

	client, err := watch.NewWatchClient("127.0.0.1:12345")
	if err != nil {
		fmt.Println(err)
		return
	}
	client.RegisterCallback(watch_config.NodeRegisterKey, rpc.WatchNodeRegister)
	client.Start()

	rpc.RegisterService("DB", &DbService{})

	//rpc节点服务
	rpc.InitNode(&rpc.NodeConfig{
		Client:   client,
		Id:       0,
		Nodename: "db0",
		Nodetype: "db",
		Set:      "",
		Host:     "127.0.0.1",
		Port:     10001,
		Region:   0,
	})
}

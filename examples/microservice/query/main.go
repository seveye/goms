package main

import (
	"fmt"

	"gitee.com/jkkkls/goms/rpc"
	"gitee.com/jkkkls/goms/util"
	"gitee.com/jkkkls/goms/watch"
	"gitee.com/jkkkls/goms/watch_config"
)

func main() {
	util.GosLogInit("query0", "./", true, 0)

	client, err := watch.NewWatchClient("127.0.0.1:12345")
	if err != nil {
		fmt.Println(err)
		return
	}
	client.RegisterCallback(watch_config.NodeRegisterKey, rpc.WatchNodeRegister)
	client.Start()

	rpc.RegisterService("Qeury", &QeuryService{})

	//rpc节点服务
	rpc.InitNode(&rpc.NodeConfig{
		Client:   client,
		Id:       1,
		Nodename: "query0",
		Nodetype: "query",
		Set:      "",
		Host:     "127.0.0.1",
		Port:     10011,
		Region:   0,
	})
}

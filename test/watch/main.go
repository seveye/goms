package main

import (
	"fmt"
	"log"

	"gitee.com/jkkkls/goms/rpc"
	"gitee.com/jkkkls/goms/watch"
	"gitee.com/jkkkls/goms/watch_config"
)

func main() {
	client, err := watch.NewWatchClient("127.0.0.1:12345")
	if err != nil {
		fmt.Println(err)
		return
	}
	client.RegisterCallback(watch_config.NodeRegisterKey, rpc.WatchNodeRegister)
	client.Start()

	// watch_config.RegisterNode(client, &watch_config.NodeInfo{
	// 	Id:   1,
	// 	Name: "common1",
	// 	Type: "common",
	// 	Service: []*watch_config.ServiceInfo{
	// 		{Name: "Login"},
	// 		{Name: "Red"},
	// 	},
	// })

	// watch_config.RegisterNode(client, &watch_config.NodeInfo{
	// 	Id:   2,
	// 	Name: "common2",
	// 	Type: "common",
	// 	Service: []*watch_config.ServiceInfo{
	// 		{Name: "Login"},
	// 		{Name: "Red"},
	// 	},
	// })

	// watch_config.RegisterNode(client, &watch_config.NodeInfo{
	// 	Id:   1,
	// 	Name: "game1",
	// 	Type: "game",
	// 	Service: []*watch_config.ServiceInfo{
	// 		{Name: "Game"},
	// 		{Name: "Red"},
	// 	},
	// })

	// watch_config.RegisterNode(client, &watch_config.NodeInfo{
	// 	Id:   2,
	// 	Name: "game2",
	// 	Type: "game",
	// 	Service: []*watch_config.ServiceInfo{
	// 		{Name: "Game"},
	// 		{Name: "Red"},
	// 	},
	// })

	node := watch_config.AllocServiceNode(client, "Login")
	log.Println("Login->", node)

	watch_config.AddServiceLoad(client, node, "Login", 1)
	node = watch_config.AllocServiceNode(client, "Login")
	log.Println("Login->", node)
	// node = watch_config.AllocServiceNode(client, "Red")
	// log.Println("Red->", node)

	// node := watch_config.AllocServiceNode(client, "Login")
	// log.Println("node", node)
}

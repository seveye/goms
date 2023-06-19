package watch_config

import (
	"fmt"
	"strconv"

	"github.com/seveye/goms/util"
	"github.com/seveye/goms/watch"
)

func allocLoad(client *watch.WatchClient, key string) string {
	values := client.Hgetall(key)

	var (
		names   []string
		weights []uint64
	)

	for i := 0; i < len(values); i = i + 2 {
		n, _ := strconv.Atoi(values[i+1])
		weights = append(weights, uint64(n))
		names = append(names, values[i])
	}

	i := util.RandLessWight(weights)
	if i < 0 {
		return ""
	}
	return names[i]
}

// AllocServiceNode 请求一个服务节点
func AllocServiceNode(client *watch.WatchClient, serviceName string) string {
	key := fmt.Sprintf("%v:%v", ServiceLoadKey, serviceName)
	return allocLoad(client, key)
}

// AddServiceLoad 更新一个服务负载
func AddServiceLoad(client *watch.WatchClient, nodeName, serviceName string, add int) {
	key := fmt.Sprintf("%v:%v", ServiceLoadKey, serviceName)
	client.Hincrby(key, nodeName, add)
}

// AllocNode 请求一个类型节点
func AllocNode(client *watch.WatchClient, NodeType string) string {
	key := fmt.Sprintf("%v:%v", NodeLoadKey, NodeType)
	return allocLoad(client, key)
}

// AddNodeLoad 更新一个服务负载
func AddNodeLoad(client *watch.WatchClient, nodeName, NodeType string, add int) {
	key := fmt.Sprintf("%v:%v", NodeLoadKey, NodeType)
	client.Hincrby(key, nodeName, add)
}

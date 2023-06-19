package watch_config

// Copyright 2017 guangbo. All rights reserved.

//节点配置，节点启动读取
//key: node:nodeName
//节点通用配置，用于服务注册服务发现
//key: nodeRegister

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/seveye/goms/watch"
)

const (
	NodeKey         = "node:"
	ServiceKey      = "service:"
	NodeRegisterKey = "nodeRegister"
	NodeLoadKey     = "node_load"
	ServiceLoadKey  = "service_load"
)

//节点配置函数

func GetAllName(client *watch.WatchClient, prefix string) []string {
	var ret []string

	nodes := client.KeyPrefix(prefix)
	for i := 0; i < len(nodes); i++ {
		arr := strings.Split(nodes[i], ":")
		if len(arr) == 2 {
			ret = append(ret, arr[1])
		}
	}

	return ret
}

func GetNodeAllName(client *watch.WatchClient) []string {
	return GetAllName(client, NodeKey)
}

func GetNodeAllConfig(client *watch.WatchClient, nodeName string) map[string]string {
	m := make(map[string]string)
	values := client.Hgetall(NodeKey + nodeName)
	for i := 0; i < len(values); i = i + 2 {
		m[values[i]] = values[i+1]
	}
	return m
}

func SetNodeConfig(client *watch.WatchClient, nodeName, key, value string) {
	client.Hset(NodeKey+nodeName, key, value)
}

func DelNodeConfig(client *watch.WatchClient, nodeName string) {
	client.Del(NodeKey + nodeName)
}

func GetNodeConfig(client *watch.WatchClient, nodeName, key string) string {
	return client.Hget(NodeKey+nodeName, key)
}

type FunctionInfo struct {
	Name string `json:"name,omitempty"`
	Cmd  uint16 `json:"cmd,omitempty"`
}

type ServiceInfo struct {
	Name string          `json:"name,omitempty"`
	Func []*FunctionInfo `json:"func,omitempty"`
}

// 节点注册函数
type NodeInfo struct {
	Id      uint64         `json:"id,omitempty"`
	Name    string         `json:"name,omitempty"`    //
	Type    string         `json:"type,omitempty"`    //
	Address string         `json:"address,omitempty"` //ip:port
	Service []*ServiceInfo `json:"service,omitempty"` //服务列表
	Region  uint32         `json:"region,omitempty"`
	Set     string         `json:"set,omitempty"`
}

func GetAllRegisterNode(client *watch.WatchClient, set string) map[string]*NodeInfo {
	m := make(map[string]*NodeInfo)
	key := fmt.Sprintf("%v:%v", NodeRegisterKey, set)
	values := client.Hgetall(key)
	for i := 0; i < len(values); i = i + 2 {
		info := &NodeInfo{}
		json.Unmarshal([]byte(values[i+1]), info)
		m[values[i]] = info
	}
	return m
}

// RegisterNode 注册节点和服务
func RegisterNode(client *watch.WatchClient, info *NodeInfo) {
	//注册节点
	key := fmt.Sprintf("%v:%v", NodeRegisterKey, info.Set)
	buff, _ := json.Marshal(info)
	client.Hset(key, info.Name, string(buff))

	//注册节点类型负载
	key = fmt.Sprintf("%v:%v", NodeLoadKey, info.Type)
	client.Hset(key, info.Name, "0")

	//注册服务类型负载
	for _, v := range info.Service {
		key = fmt.Sprintf("%v:%v", ServiceLoadKey, v.Name)
		client.Hset(key, info.Name, "0")
	}
}

func DelRegisterNode(client *watch.WatchClient, info *NodeInfo) {
	//删除节点
	key := fmt.Sprintf("%v:%v", NodeRegisterKey, info.Set)
	client.Hset(key, info.Name, "")

	//删除节点类型负载
	skey := fmt.Sprintf("%v:%v", NodeLoadKey, info.Type)
	client.Hset(skey, info.Name, "")

	for _, v := range info.Service {
		key = fmt.Sprintf("%v:%v", ServiceLoadKey, v.Name)
		client.Hset(key, info.Name, "0")
	}
}

func GetRegisterNode(client *watch.WatchClient, nodeName, set string) *NodeInfo {
	key := fmt.Sprintf("%v:%v", NodeRegisterKey, set)
	value := client.Hget(key, nodeName)
	if value == "" {
		return nil
	}
	info := &NodeInfo{}
	json.Unmarshal([]byte(value), info)
	return info
}

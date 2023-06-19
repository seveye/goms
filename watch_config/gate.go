package watch_config

// Copyright 2017 guangbo. All rights reserved.

//网关信息

import (
	"encoding/json"

	"github.com/seveye/goms/watch"
)

const (
	GateKey = "gate"
)

// GateInfo 网关节点信息
// 网关定时更新，网关列表服务定时拉取
type GateInfo struct {
	Name       string
	Address    string
	SSLAddress string
	KCPAddress string
	Region     uint32
	Time       int64
}

// GetAllGate 获取所有网关信息
func GetAllGate(client *watch.WatchClient) []*GateInfo {
	var gates []*GateInfo
	values := client.Hgetall(GateKey)
	for i := 0; i < len(values); i = i + 2 {
		gate := &GateInfo{}
		json.Unmarshal([]byte(values[i+1]), gate)
		gates = append(gates, gate)
	}
	return gates
}

// SetGateInfo 保存网关信息
func SetGateInfo(client *watch.WatchClient, gate *GateInfo) {
	buff, _ := json.Marshal(gate)
	client.Hset(GateKey, gate.Name, string(buff))
}

// GetGateInfo 获取指定网关信息
func GetGateInfo(client *watch.WatchClient, name string) *GateInfo {
	value := client.Hget(GateKey, name)
	if value == "" {
		return nil
	}
	gate := &GateInfo{}
	err := json.Unmarshal([]byte(value), gate)
	if err != nil {
		return nil
	}
	return gate
}

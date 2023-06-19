package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gitee.com/jkkkls/goms/util"

	"gitee.com/jkkkls/goms/examples/microservice/pb"
	"gitee.com/jkkkls/goms/rpc"
)

// Query 查询
func Query(key string) string {
	req := &pb.QueryReq{
		Key: key,
	}
	rsp := &pb.QueryRsp{}
	ret, err := rpc.Call(rpc.EmptyContext(), "DB.Query", req, rsp)
	log.Println(ret, err)
	return rsp.Value
}

// Update 查询
func Update(key, value string) {
	req := &pb.UpdateReq{
		Key:   key,
		Value: value,
	}
	rsp := &pb.UpdateRsp{}
	ret, err := rpc.Call(rpc.EmptyContext(), "DB.Update", req, rsp)
	log.Println(ret, err)

	cs := rpc.QueryNodeStatus()
	for _, c := range cs {
		log.Println("node status", c.Info.Name, c.Info.Address, c.Close)
	}
}

// QeuryService
type QeuryService struct {
	kvs sync.Map
}

// Exit 退出处理
func (service *QeuryService) Exit()                                 {}
func (service *QeuryService) OnEvent(eventName string, args ...any) {}

func (service *QeuryService) NodeConn(name string) {
	log.Println("NodeConn---", name)
}
func (service *QeuryService) NodeClose(name string) {
	log.Println("NodeClose----", name)
}

// Run 服务启动函数
func (service *QeuryService) Run() error {

	go func() {
		util.Recover()

		i := 0
		for {
			Update("a", fmt.Sprintf("%v", i))
			i++

			time.Sleep(10 * time.Second)
		}
	}()

	go func() {
		util.Recover()

		for {
			time.Sleep(10 * time.Second)
			util.Info("Query", "value", Query("a"))
		}
	}()
	return nil
}

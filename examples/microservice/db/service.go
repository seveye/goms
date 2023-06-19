package main

import (
	"sync"

	"github.com/seveye/goms/util"

	"github.com/seveye/goms/examples/microservice/pb"
	"github.com/seveye/goms/rpc"
)

// DbService
type DbService struct {
	kvs sync.Map
}

// Exit 退出处理
func (service *DbService) Exit() {}

// Run 服务启动函数
func (service *DbService) Run() error {
	return nil
}
func (service *DbService) OnEvent(eventName string, args ...any) {}

func (service *DbService) NodeConn(name string)  {}
func (service *DbService) NodeClose(name string) {}

func (service *DbService) Update(context *rpc.Context, req *pb.UpdateReq, rsp *pb.UpdateRsp) (uint32, error) {
	util.Info("Update", "req", req)
	service.kvs.Store(req.Key, req.Value)
	return 0, nil
}

func (service *DbService) Query(context *rpc.Context, req *pb.QueryReq, rsp *pb.QueryRsp) (uint32, error) {
	v, ok := service.kvs.Load(req.Key)
	if ok {
		rsp.Value = v.(string)
	}

	return 0, nil
}

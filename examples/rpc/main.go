package main

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/seveye/goms/rpc"
	"google.golang.org/protobuf/proto"
)

type User struct{}

func (t *User) Add(conn *rpc.Context, args *AddReq, reply proto.Message) (uint32, error) {
	log.Println("server print1, ", args.A)
	// reply.C = args.A + args.B
	return 123, nil
}

func (t *User) Add1(conn *rpc.Context, args *AddReq, reply *AddRsp) (uint32, error) {
	log.Println("server print2, ", conn.Remote, conn.CallId, args.String())
	reply.C = args.A + args.B
	return 123, nil
}

func main() {
	s := rpc.NewServer()

	s.RegisterName("AddServer", &User{})
	s.RpcCallBack = func(context *rpc.Context, methon string, req, rsp proto.Message, ret uint16, err error, cost time.Duration) {
		// log.Println("--1-", conn.Remote, methon)
		// log.Println("--1-", req.String())
		// log.Println("--1-", rsp.String())
		// log.Println("--1-", ret, err, cost)
	}

	l, err := net.Listen("tcp", ":12323")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	go s.Accept(l)

	// var rsp AddRsp
	// conn := &rpc.Context{
	// 	Remote: "127.0.0.1:1111",
	// }
	// ret, _ := s.InternalCall(conn, "AddServer.Add", &AddReq{123123, 3}, nil)
	// log.Println(ret, rsp.C)

	// req := &AddReq{123123, 3}
	// // var rsp AddRsp
	// var buff1 []byte
	// buff, _ := proto.Marshal(req)
	// ret, buff1, _ = s.RawCall("AddServer.Add", buff)
	// proto.Unmarshal(buff1, &rsp)
	// log.Println(ret, rsp.C)

	c, _ := rpc.Dial("tcp", "127.0.0.1:12323")

	log.Println("----client1----")
	req := &AddReq{A: 123123, B: 3}
	buff, _ := json.Marshal(req)
	conn := &rpc.Context{
		Remote: "127.0.0.1:1111",
		CallId: 123,
	}
	ret, buff1, err1 := c.JsonCall(conn, "AddServer.Add1", buff)
	log.Println("----client2----", ret, err1, string(buff1))

	// req := &AddReq{123123, 3}
	// var rsp AddRsp
	// conn := &rpc.Context{
	// 	Remote: "127.0.0.1:1111",
	// }
	// ret, err := c.Call(conn, "AddServer.Add1", req, &rsp)
	// log.Println("client print1, ", ret, err, rsp.C)

	// log.Println("-----------------------")

	// for i := 0; i < 100; i++ {
	// 	// var rsp AddRsp

	// 	j := uint64(i)
	// 	req := &AddReq{j, 3}
	// 	c.Send(conn, "AddServer.Add", req)
	// 	// time.Sleep(100 * time.Millisecond)
	// }

	// // c.Send(conn, "AddServer.Add", req)
	// log.Println("client print2")

	// req1 := &AddReq{2, 3}
	// ret, err = c.Call(conn, "AddServer.Add1", req1, &rsp)
	// log.Println("client print3, ", ret, err, rsp.C)
	time.Sleep(100 * time.Second)
}

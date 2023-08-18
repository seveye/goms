package main

import (
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"

	"github.com/google/gops/agent"

	"github.com/seveye/goms/benchmark/rpc/proto"
	"github.com/seveye/goms/rpc"
	"github.com/seveye/goms/util"
)

type User struct{}

func (t *User) Add(conn *rpc.Context, args *proto.AddReq, reply *proto.AddRsp) (uint32, error) {
	reply.C = args.A + args.B
	return 123, nil
}

func main() {
	go http.ListenAndServe(":6063", nil)

	util.SetUlimit()

	if err := agent.Listen(agent.Options{}); err != nil {
		log.Println(err)
		return
	}

	s := rpc.NewServer()

	s.RegisterName("AddServer", &User{})

	l, err := net.Listen("tcp", ":12323")
	if err != nil {
		log.Fatal("listen error:", err)
	}

	log.Println("rpc server start")
	s.Accept(l)
}

package main

import (
	"log"
	"net"

	"github.com/google/gops/agent"

	"github.com/seveye/goms/benchmark/rpc/proto"
	"github.com/seveye/goms/rpc"
)

type User struct{}

func (t *User) Add(conn *rpc.Context, args *proto.AddReq, reply *proto.AddRsp) (uint32, error) {
	reply.C = args.A + args.B
	return 123, nil
}

func main() {
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
	s.Accept(l)

}

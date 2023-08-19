package main

import (
	"log"
	"net"

	"github.com/google/gops/agent"
	"google.golang.org/protobuf/proto"

	pb "github.com/seveye/goms/benchmark/rpc/proto"
	"github.com/seveye/goms/rpc"
)

type User struct{}

func (t *User) Add(conn *rpc.Context, args *pb.AddReq, msg proto.Message) (uint32, error) {
	// func (t *User) Add(conn *rpc.Context, args *pb.AddReq, reply *pb.AddRsp) (uint32, error) {
	// reply.C = args.A + args.B
	// reply.Name = args.Name
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

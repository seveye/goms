package main

import (
	"log"
	"net"

	"net/http"
	_ "net/http/pprof"

	"google.golang.org/protobuf/proto"

	pb "github.com/seveye/goms/benchmark/rpc/proto"
	"github.com/seveye/goms/rpc"
)

type User struct{}

func (t *User) Add2(conn *rpc.Context, args *pb.AddReq, msg proto.Message) (uint16, error) {
	return 123, nil
}
func (t *User) Add(conn *rpc.Context, args *pb.AddReq, reply *pb.AddRsp) (uint16, error) {
	reply.C = args.A + args.B
	reply.Name = args.Name
	return 123, nil
}

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6061", nil))
	}()

	s := rpc.NewServer()

	s.RegisterName("AddServer", &User{})

	l, err := net.Listen("tcp", ":12323")
	if err != nil {
		log.Fatal("listen error:", err)
	}
	s.Accept(l)

}

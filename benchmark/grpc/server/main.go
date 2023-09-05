package main

import (
	"context"
	"log"
	"net"

	pb "github.com/seveye/goms/benchmark/grpc/proto"
	"google.golang.org/grpc"
)

type AddServer struct {
	*pb.UnimplementedAddServiceServer
}

func (s *AddServer) Add(ctx context.Context, req *pb.AddReq) (*pb.AddRsp, error) {
	return &pb.AddRsp{C: req.A + req.B, Name: req.Name}, nil
}

func main() {
	l, err := net.Listen("tcp", ":12324")
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := l.Close(); err != nil {
			log.Println("Failed to close", "err", err)
		}
	}()

	ctx := context.Background()
	s := grpc.NewServer()

	pb.RegisterAddServiceServer(s, &AddServer{})

	go func() {
		defer s.GracefulStop()
		<-ctx.Done()
	}()

	s.Serve(l)
}

package main

import (
	"log"
	"time"

	pb "gitee.com/jkkkls/goms/benchmark/rpc/proto"
	"gitee.com/jkkkls/goms/rpc"
	"github.com/golang/protobuf/proto"
)

var host = "192.168.0.119:12323"

func testRpcCall(n, c, t int) {
	req := &pb.AddReq{A: 123123, B: 3}
	buff, _ := proto.Marshal(req)
	ch := make(chan int, 1)

	info := &rpc.Context{
		Remote: "127.0.0.1:1111",
	}

	begin := time.Now().Unix()
	for i := 0; i < n; i++ {
		conn, _ := rpc.Dial("tcp", host)

		for j := 0; j < c; j++ {
			go func() {
				for k := 0; k < t; k++ {
					conn.RawCall(info, "AddServer.Add", buff)
				}

				ch <- 1
			}()
		}
	}

	for i := 0; i < c*n; i++ {
		<-ch
	}
	log.Println("同步调用测试", n*c*t, time.Now().Unix()-begin)
}

func testRpcSend(n, c, t int) {
	req := &pb.AddReq{A: 123123, B: 3}
	buff, _ := proto.Marshal(req)
	ch := make(chan int, 1)

	info := &rpc.Context{
		Remote: "127.0.0.1:1111",
	}

	begin := time.Now().Unix()
	for i := 0; i < n; i++ {
		conn, _ := rpc.Dial("tcp", host)

		for j := 0; j < c; j++ {
			go func() {
				for k := 0; k < t; k++ {
					conn.RawSend(info, "AddServer.Add", buff)
				}

				ch <- 1
			}()
		}
	}

	for i := 0; i < c*n; i++ {
		<-ch
	}
	log.Println("异步调用测试", n*c*t, time.Now().Unix()-begin)
}

func main() {
	//异步调用
	testRpcSend(1, 1, 1000000)
	testRpcSend(1, 10, 100000)
	testRpcSend(10, 1, 100000)
	testRpcSend(10, 10, 10000)
	testRpcSend(10, 100, 10000)

	//同步调用
	testRpcCall(1, 1, 1000000)
	testRpcCall(1, 10, 100000)
	testRpcCall(10, 1, 100000)
	testRpcCall(10, 10, 10000)
	testRpcCall(10, 100, 10000)
}

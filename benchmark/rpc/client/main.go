package main

import (
	"log"
	"strings"
	"sync"
	"time"

	pb "github.com/seveye/goms/benchmark/rpc/proto"
	"github.com/seveye/goms/rpc"
	"google.golang.org/protobuf/proto"
)

var host = "127.0.0.1:12323"

func testRpcCall(n, c, t int) {
	req := &pb.AddReq{A: 123123, B: 3, Name: strings.Repeat("1234", 32)}
	buff, _ := proto.Marshal(req)

	info := &rpc.Context{
		Remote: "127.0.0.1:1111",
	}

	begin := time.Now()
	var w sync.WaitGroup
	for i := 0; i < n; i++ {
		conn, _ := rpc.Dial("tcp", host)

		for j := 0; j < c; j++ {
			w.Add(1)
			go func() {
				defer w.Done()
				for k := 0; k < t; k++ {
					conn.RawCall(info, "AddServer.Add", buff)
				}
			}()
		}
	}

	w.Wait()
	log.Println("同步调用测试", n*c*t, time.Since(begin).String())
}

func testRpcSend(n, c, t int) {
	req := &pb.AddReq{A: 123123, B: 3}
	buff, _ := proto.Marshal(req)

	info := &rpc.Context{
		Remote: "127.0.0.1:1111",
	}

	begin := time.Now()
	var w sync.WaitGroup
	for i := 0; i < n; i++ {
		conn, _ := rpc.Dial("tcp", host)

		for j := 0; j < c; j++ {
			w.Add(1)
			go func() {
				defer w.Done()
				for k := 0; k < t; k++ {
					conn.RawSend(info, "AddServer.Add", buff)
				}
			}()
		}
	}

	w.Wait()
	log.Println("异步调用测试", n*c*t, time.Since(begin).String())
}

func main() {
	// testRpcSend(1, 1, 1)

	// 	2023/08/19 13:26:44 异步调用测试 1000000 18.896395987s
	// 2023/08/19 13:26:45 异步调用测试 1000000 920.00817ms
	// 2023/08/19 13:26:55 异步调用测试 1000000 9.79035734s
	// 2023/08/19 13:26:57 异步调用测试 1000000 1.880022947s
	// 2023/08/19 13:27:08 异步调用测试 10000000 11.175221108s

	//异步调用
	testRpcSend(1, 1, 1000000)
	testRpcSend(1, 10, 100000)
	testRpcSend(10, 1, 100000)
	testRpcSend(10, 10, 10000)
	testRpcSend(10, 100, 10000)

	// 2023/08/19 13:16:53 同步调用测试 1000000 1m41.476370857s
	// 2023/08/19 13:16:55 同步调用测试 1000000 1.608302892s
	// 2023/08/19 13:17:22 同步调用测试 1000000 26.936508476s
	// 2023/08/19 13:17:24 同步调用测试 1000000 1.886314927s
	// 2023/08/19 13:17:43 同步调用测试 10000000 18.954123831s

	// 2023/08/19 13:21:39 同步调用测试 1000000 1m50.571781066s
	// 2023/08/19 13:21:41 同步调用测试 1000000 1.832360947s
	// 2023/08/19 13:22:07 同步调用测试 1000000 25.541016672s
	// 2023/08/19 13:22:09 同步调用测试 1000000 2.50515229s
	// 2023/08/19 13:22:28 同步调用测试 10000000 18.692950235s
	//同步调用
	// testRpcCall(1, 1, 1000000)
	// testRpcCall(1, 10, 100000)
	// testRpcCall(10, 1, 100000)
	// testRpcCall(10, 10, 10000)
	// testRpcCall(10, 100, 10000)
}

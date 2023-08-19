package main

import (
	"log"
	// "net/http"
	// _ "net/http/pprof"
	"strings"
	"sync"
	"sync/atomic"
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

	var n1, n2 int64

	begin := time.Now()
	var w sync.WaitGroup
	for i := 0; i < n; i++ {
		conn, _ := rpc.Dial("tcp", host)

		for j := 0; j < c; j++ {
			w.Add(1)
			go func() {
				defer w.Done()
				for k := 0; k < t; k++ {
					ret, _, err := conn.RawCall(info, "AddServer.Add", buff)
					if err != nil || ret != 123 {
						log.Println(ret, err)
						atomic.AddInt64(&n1, 1)
					} else {
						atomic.AddInt64(&n2, 1)
					}
				}
			}()
		}
	}

	w.Wait()
	log.Println("同步调用测试", n*c*t, time.Since(begin).String(), n1, n2)
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
	// go http.ListenAndServe("0.0.0.0:6060", nil)
	// testRpcSend(1, 1, 1)

	// 	2023/08/19 13:26:44 异步调用测试 1000000 18.896395987s
	// 2023/08/19 13:26:45 异步调用测试 1000000 920.00817ms
	// 2023/08/19 13:26:55 异步调用测试 1000000 9.79035734s
	// 2023/08/19 13:26:57 异步调用测试 1000000 1.880022947s
	// 2023/08/19 13:27:08 异步调用测试 10000000 11.175221108s

	//异步调用
	// testRpcSend(1, 1, 1000000)
	// testRpcSend(1, 10, 100000)
	// testRpcSend(10, 1, 100000)
	// testRpcSend(10, 10, 10000)
	// testRpcSend(10, 100, 10000)

	// 2023/08/19 13:16:53 同步调用测试 1000000 1m41.476370857s
	// 2023/08/19 13:16:55 同步调用测试 1000000 1.608302892s
	// 2023/08/19 13:17:22 同步调用测试 1000000 26.936508476s
	// 2023/08/19 13:17:24 同步调用测试 1000000 1.886314927s
	// 2023/08/19 13:17:43 同步调用测试 10000000 18.954123831s

	// 2023/08/19 17:57:40 同步调用测试 1000000 2m49.620081021s 0 1000000
	// 2023/08/19 17:47:48 同步调用测试 1000000 26.726200078s 0 1000000
	// 2023/08/19 17:48:17 同步调用测试 1000000 29.04003047s 0 1000000
	// 2023/08/19 17:48:36 同步调用测试 1000000 19.881092106s 0 1000000
	// 2023/08/19 17:52:31 同步调用测试 10000000 3m54.878909705s 0 10000000
	//同步调用
	testRpcCall(1, 1, 1000000)
	// testRpcCall(1, 10, 100000)
	// testRpcCall(10, 1, 100000)
	// testRpcCall(10, 10, 10000)
	// testRpcCall(10, 100, 10000)

}

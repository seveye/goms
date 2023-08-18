package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	pb "github.com/seveye/goms/benchmark/rpc/proto"
	"github.com/seveye/goms/rpc"
	"google.golang.org/protobuf/proto"
)

var host = "192.168.186.141:12323"

func testRpcCall(n, c, t int) {
	req := &pb.AddReq{A: 123123, B: 3}
	buff, _ := proto.Marshal(req)
	var w sync.WaitGroup
	var err error

	info := &rpc.Context{
		Remote: "127.0.0.1:1111",
	}

	begin := time.Now().Unix()
	conns := make([]*rpc.Client, n)
	for i := 0; i < n; i++ {
		conns[i], err = rpc.Dial("tcp", host)
		if err != nil {
			log.Println(err)
			return
		}
		for j := 0; j < c; j++ {
			w.Add(1)
			conn := conns[i]
			go func() {
				defer w.Done()
				for k := 0; k < t; k++ {
					conn.RawCall(info, "AddServer.Add", buff)
				}
			}()
		}
	}

	w.Wait()
	log.Println("同步调用测试", n*c*t, time.Now().Unix()-begin)
}

func testRpcSend(n, c, t int) {
	req := &pb.AddReq{A: 123123, B: 3}
	buff, _ := proto.Marshal(req)

	info := &rpc.Context{
		Remote: "127.0.0.1:1111",
	}

	var w sync.WaitGroup
	var err error
	begin := time.Now()
	conns := make([]*rpc.Client, n)
	for i := 0; i < n; i++ {
		conns[i], err = rpc.Dial("tcp", host)
		if err != nil {
			log.Println(err)
			return
		}

		w.Add(1)
		conn := conns[i]
		go func() {
			defer w.Done()

			var w1 sync.WaitGroup
			for j := 0; j < c; j++ {
				w1.Add(1)
				go func() {
					defer w1.Done()
					for k := 0; k < t; k++ {
						conn.RawSend(info, "AddServer.Add", buff)
					}
				}()
			}
			w1.Wait()
		}()
	}

	w.Wait()
	log.Println("异步调用测试", n*c*t, time.Since(begin).String())
}

func main() {
	go http.ListenAndServe(":6062", nil)
	//异步调用
	// testRpcSend(1, 1, 1000000)
	// testRpcSend(1, 10, 100000)
	// testRpcSend(10, 1, 100000)
	// testRpcSend(10, 10, 10000)
	testRpcSend(10, 100, 10000)

	//同步调用
	// testRpcCall(1, 1, 1000000)
	// testRpcCall(1, 10, 100000)
	// testRpcCall(10, 1, 100000)
	// testRpcCall(10, 10, 10000)
	// testRpcCall(10, 100, 10000)
}

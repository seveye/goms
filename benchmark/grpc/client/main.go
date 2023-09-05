package main

import (
	"context"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	pb "github.com/seveye/goms/benchmark/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {

	// 2023/08/25 18:10:47 同步调用测试 1000000 24.277007093s 0 1000000
	// 2023/08/25 18:11:15 同步调用测试 1000000 27.832550785s 0 1000000
	// 2023/08/25 18:11:26 同步调用测试 1000000 10.814157271s 0 1000000
	testRpcCall(1, 10, 100000)
	testRpcCall(10, 1, 100000)
	testRpcCall(10, 10, 10000)

	time.Sleep(time.Hour)
}

var host = "127.0.0.1:12324"

func testRpcCall(n, c, t int) {
	req := &pb.AddReq{A: 123123, B: 3, Name: strings.Repeat("1234", 32)}
	// buff, _ := proto.Marshal(req)

	// info := &rpc.Context{
	// 	Remote: "127.0.0.1:1111",
	// }

	var n1, n2 int64

	begin := time.Now()
	var w sync.WaitGroup
	for i := 0; i < n; i++ {
		conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Println(err)
			return
		}
		client := pb.NewAddServiceClient(conn)

		for j := 0; j < c; j++ {
			w.Add(1)
			go func() {
				defer w.Done()
				for k := 0; k < t; k++ {
					rsp, err := client.Add(context.Background(), req)
					if err != nil || rsp.C != 123126 {
						atomic.AddInt64(&n1, 1)
					} else {
						atomic.AddInt64(&n2, 1)
					}
				}
			}()
		}
	}

	w.Wait()
	log.Println("//同步调用测试", n*c*t, time.Since(begin).String(), n1, n2)
}

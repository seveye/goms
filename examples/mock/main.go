package main

import (
	"log"

	"gitee.com/jkkkls/goms/rpc"
)

func main() {
	rpc.InitMock()
	defer rpc.RemoveMock()
	rpc.InsertMock("Player.GetInfo", &rpc.MockRsp{
		Rsp:   []byte("ok"),
		Ret:   1,
		Error: nil,
	})

	rpc.InsertMock("Player.GetInfo2", &rpc.MockRsp{
		Rsp: &rpc.Context{
			Remote: "127.0.0.1:123123",
			CallId: 123123,
		},
		Ret:   2,
		Error: nil,
	})

	ret, buff, err := rpc.RawCall(rpc.EmptyContext(), "Player.GetInfo", nil)
	log.Println(ret)
	log.Println(string(buff))
	log.Println(err)

	rsp := &rpc.Context{}
	ret, err = rpc.Call(rpc.EmptyContext(), "Player.GetInfo2", nil, rsp)
	log.Println(ret)
	log.Println(rsp.String())
	log.Println(err)
}

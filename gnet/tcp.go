package gnet

import (
	"fmt"
	"log"
	"net"
	"os"

	"gitee.com/jkkkls/goms/util"
)

func RunTCPServer(port int) error {
	util.Info("启动tcp服务器", "port", port)
	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", port))
	if err != nil {
		return err
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Println("Accept tcp, err:", err)
			os.Exit(0)
		}
		util.Submit(func() {
			gater.handleConn(conn, conn.RemoteAddr().String(), "tcp")
		})
	}
}

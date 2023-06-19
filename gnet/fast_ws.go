package gnet

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"gitee.com/jkkkls/goms/util"
	"github.com/fasthttp/websocket"
)

// WsConn2 websocket封装
type WsConn2 struct {
	Conn *websocket.Conn
	Buff bytes.Buffer
}

func (c *WsConn2) Write(b []byte) (int, error) {
	err := c.Conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

func (c *WsConn2) Read(b []byte) (int, error) {
	if c.Buff.Len() > 0 {

		return c.Buff.Read(b)
	}
	_, buff, err := c.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	c.Buff.Write(buff)
	return c.Buff.Read(b)
}

func (c *WsConn2) Close() error {
	return c.Conn.Close()
}

func (s *WsConn2) SetWriteDeadline(t time.Time) error {
	return s.Conn.SetWriteDeadline(t)
}

func RunWSServer2(port int) error {
	util.Info("启动websocket", "port", port)
	http.HandleFunc("/ws", HandleWebsocket2)
	err := http.ListenAndServe(fmt.Sprintf(":%v", port), nil)
	if err != nil {
		return err
	}

	return nil
}

func RunWSSServer2(port int, crtFile, keyFile string) error {
	util.Info("启动websocket2 ssl", "port", port)
	http.HandleFunc("/wss", HandleWebsocket2)
	err := http.ListenAndServeTLS(fmt.Sprintf(":%v", port), crtFile, keyFile, nil)
	if err != nil {
		return err
	}

	return nil
}

var upgrader2 = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//不检查origin
	//https://time-track.cn/websocket-and-golang.html
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebsocket websocket新连接回调
func HandleWebsocket2(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader2.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	util.Submit(func() {
		gater.handleConn(&WsConn2{Conn: conn}, getRemote(r), "ws")
	})
}

// Copyright 2017 guangbo. All rights reserved.

//watch模块服务端
//使用示例参考gitee.com/goxiang2/server/example/watch
package watch

import (
	"bufio"
	"log"
	"net"
	"time"

	"strconv"
	"sync"

	"gitee.com/jkkkls/goms/util"
)

//watch连接信息
type WatchConnect struct {
	Id        int           //客户端连接id
	Conn      net.Conn      //
	Status    int           //0-未初始化 1-正常 2-断开连接
	Reader    *bufio.Reader //
	Writer    *bufio.Writer //
	Remote    string        //
	WatchKeys []string      //观察的key列表

	Out   chan *WatchMessage
	In    chan *WatchMessage
	Close chan interface{}
}

type WatchServer struct {
	Mutex       sync.Mutex
	Clients     map[int]*WatchConnect
	KeyWatchers map[string]map[int]*WatchConnect
	Set         *util.HashSet
}

func NewWatchServer() *WatchServer {
	return &WatchServer{
		Clients:     make(map[int]*WatchConnect),
		KeyWatchers: make(map[string]map[int]*WatchConnect),
		Set:         util.NewHashSet(),
	}
}

// Start 启动WatchServer
func (ws *WatchServer) Start(port string) error {
	// ws.Set.Load()

	listener, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	log.Println("WatchServer start, host:", port)

	id := 0
	for {
		conn, err1 := listener.Accept()
		if err1 != nil {
			return err1
		}
		// log.Println("new connect ", conn.RemoteAddr().String())

		id++
		go ws.runConn(id, conn)
	}
}

// runConn 处理客户端连接
func (ws *WatchServer) runConn(id int, conn net.Conn) {
	wc := &WatchConnect{
		Id:     id,
		Conn:   conn,
		Status: 0,
		Reader: bufio.NewReader(conn),
		Writer: bufio.NewWriter(conn),
		Remote: conn.RemoteAddr().String(),

		In:    make(chan *WatchMessage, 32),
		Out:   make(chan *WatchMessage, 32),
		Close: make(chan interface{}),
	}

	defer conn.Close()
	defer util.Recover()

	ws.Mutex.Lock()
	ws.Clients[id] = wc
	ws.Mutex.Unlock()

	go ws.read(wc)
	for {
		select {
		case msg := <-wc.Out:
			WriteWatchMessage(wc.Writer, msg)
		case msg := <-wc.In:
			if msg.Cmd == "initialize" {
				ws.initializeProcess(wc, msg)
			} else if msg.Cmd == "heartbeat" {
				ws.heartbeatProcess(wc, msg)
			} else if msg.Cmd == "hset" {
				ws.hsetProcess(wc, msg)
			} else if msg.Cmd == "hget" {
				ws.hgetProcess(wc, msg)
			} else if msg.Cmd == "key_prefix" {
				ws.KeyPrefixtProcess(wc, msg)
			} else if msg.Cmd == "hgetall" {
				ws.hgetallProcess(wc, msg)
			} else if msg.Cmd == "del" {
				ws.delProcess(wc, msg)
			} else if msg.Cmd == "hincrby" {
				ws.hincrbyProcess(wc, msg)
			}
		case <-wc.Close:
			// log.Println("close,", wc.Remote)
			return
		case <-time.After(11 * time.Second):
			// log.Println("connect timeout,", wc.Remote)
			wc.Conn.Close()
			return
		}
	}
}

func (ws *WatchServer) read(wc *WatchConnect) {
	defer util.Recover()
	defer func() {
		close(wc.Close)
		wc.Status = 2
	}()

	for {
		msg, err := ReadWatchMessage(wc.Reader)
		if err != nil {
			return
		}

		wc.In <- msg
	}
}

// writeMsgToClient 回复客户端信息
func (ws *WatchServer) writeMsgToClient(wc *WatchConnect, msg *WatchMessage) {
	wc.Out <- msg
}

// initializeProcess 客户端初始化自己信息
func (ws *WatchServer) initializeProcess(wc *WatchConnect, msg *WatchMessage) {
	wc.Status = 1
	wc.WatchKeys = msg.Values
	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd: msg.Cmd,
		Seq: msg.Seq,
	})

	ws.Mutex.Lock()
	for i := 0; i < len(wc.WatchKeys); i++ {
		key := wc.WatchKeys[i]
		clients, ok := ws.KeyWatchers[key]
		if !ok {
			clients = make(map[int]*WatchConnect)
			ws.KeyWatchers[key] = clients
		}
		clients[wc.Id] = wc
	}
	ws.Mutex.Unlock()
}

func (ws *WatchServer) notify(key, field, value string) {
	//向watch的客户端推送数据
	ws.Mutex.Lock()
	defer ws.Mutex.Unlock()

	clients, ok := ws.KeyWatchers[key]
	if !ok {
		return
	}
	var delId []int
	for id, client := range clients {
		if client.Status == 2 {
			delId = append(delId, id)
			continue
		}
		ws.writeMsgToClient(client, &WatchMessage{
			Cmd:    "watch",
			Seq:    0,
			Values: []string{key, field, value},
		})
	}

	for i := 0; i < len(delId); i++ {
		delete(clients, delId[i])
	}
}

// heartbeatProcess 客户端心跳
func (ws *WatchServer) heartbeatProcess(wc *WatchConnect, msg *WatchMessage) {
	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd: msg.Cmd,
		Seq: msg.Seq,
	})
}

// key_prefix
func (ws *WatchServer) KeyPrefixtProcess(wc *WatchConnect, msg *WatchMessage) {
	key := msg.Values[0]
	value := ws.Set.KeyPrefix(key)
	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd:    msg.Cmd,
		Seq:    msg.Seq,
		Values: value,
	})
}

func (ws *WatchServer) delProcess(wc *WatchConnect, msg *WatchMessage) {
	key := msg.Values[0]
	ws.Set.Del(key)
	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd: msg.Cmd,
		Seq: msg.Seq,
	})

	ws.notify(key, "", "")

}

// hget 客户端请求key,field信息
func (ws *WatchServer) hgetProcess(wc *WatchConnect, msg *WatchMessage) {
	key := msg.Values[0]
	field := msg.Values[1]
	value := ws.Set.Hget(key, field)
	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd:    msg.Cmd,
		Seq:    msg.Seq,
		Values: []string{value},
	})
}

// hgetall 客户端请求key信息
func (ws *WatchServer) hgetallProcess(wc *WatchConnect, msg *WatchMessage) {
	key := msg.Values[0]
	values := ws.Set.Hgetall(key)

	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd:    msg.Cmd,
		Seq:    msg.Seq,
		Values: values,
	})
}

// hset 设置key信息
func (ws *WatchServer) hsetProcess(wc *WatchConnect, msg *WatchMessage) {
	key := msg.Values[0]
	field := msg.Values[1]
	value := msg.Values[2]
	ws.Set.Hset(key, field, value)

	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd: msg.Cmd,
		Seq: msg.Seq,
	})

	ws.notify(key, field, value)
}

func (ws *WatchServer) hincrbyProcess(wc *WatchConnect, msg *WatchMessage) {
	key := msg.Values[0]
	field := msg.Values[1]
	add, _ := strconv.Atoi(msg.Values[2])

	value := ws.Set.Hincrby(key, field, add)

	ws.writeMsgToClient(wc, &WatchMessage{
		Cmd:    msg.Cmd,
		Seq:    msg.Seq,
		Values: []string{value},
	})

	ws.notify(key, field, value)
}

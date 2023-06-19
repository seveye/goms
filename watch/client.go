// Copyright 2017 guangbo. All rights reserved.

//watch模块客户端
//使用示例参考gitee.com/goxiang2/server/example/watch
package watch

import (
	"bufio"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"gitee.com/jkkkls/goms/util"
)

type WatchCallback func(string, string)

type WatchClient struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer

	Host string
	Keys []string

	Wm          sync.Mutex
	Close       chan interface{}
	WatchChan   chan *WatchMessage
	MessageChan sync.Map

	Calls sync.Map //watch回调

	Seq int32
}

// Dial 启动WatchServer
// * host watch服务器地址
// * keys 关心的keys
func NewWatchClient(host string) (*WatchClient, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	wc := &WatchClient{
		Host:      host,
		Conn:      conn,
		Reader:    bufio.NewReader(conn),
		Writer:    bufio.NewWriter(conn),
		WatchChan: make(chan *WatchMessage, 1),
		Close:     make(chan interface{}),
	}

	return wc, nil
}

// Reconnect 重连
func (wc *WatchClient) Reconnect() {
	conn, err := net.Dial("tcp", wc.Host)
	if err == nil {
		util.Warn("重连watch服务器成功", "host", wc.Host)
		wc.Conn = conn
		wc.Reader = bufio.NewReader(conn)
		wc.Writer = bufio.NewWriter(conn)
		wc.Close = make(chan interface{})
		wc.Start()
		return
	}

	util.Error("重连watch服务器失败，稍后再尝试", "host", wc.Host)
	time.AfterFunc(10*time.Second, wc.Reconnect)
}

// RegisterCallback注册需要watch的关键key
func (wc *WatchClient) RegisterCallback(key string, call WatchCallback) {
	wc.Keys = append(wc.Keys, key)
	wc.Calls.Store(key, call)
}

// Start 启动读写协程
func (wc *WatchClient) Start() {
	//处理消息接受
	go func() {
		defer util.Recover()
		defer func() {
			wc.Conn.Close()
			close(wc.Close)
		}()

		for {
			msg, err := ReadWatchMessage(wc.Reader)
			if err != nil {
				util.Error("watch服务器断开连接，错误信息:", "err", err)
				time.AfterFunc(10*time.Second, wc.Reconnect)
				return
			}
			c, ok := wc.MessageChan.Load(msg.Seq)
			if ok {
				c.(chan *WatchMessage) <- msg
			} else if msg.Cmd == "watch" {
				wc.WatchChan <- msg
			} else {
				continue
			}
		}
	}()

	//处理心跳
	go func() {
		defer util.Recover()
		defer func() {
			wc.Conn.Close()
		}()

		for {
			select {
			case <-wc.Close:
				return
			case <-time.After(5 * time.Second):
				wc.heartbeat()
			}
		}
	}()

	go func() {
		for {
			key, k, v := wc.Watch()
			if key == "" {
				return
			}
			call, ok := wc.Calls.Load(key)
			if ok {
				call.(WatchCallback)(k, v)
			}
		}
	}()

	wc.initialize(wc.Keys)
}

// Shutdown 关闭客户端连接
func (wc *WatchClient) Shutdown() {
	wc.Conn.Close()
}

// SendAndRecvMessage 重连
func (wc *WatchClient) SendAndRecvMessage(req *WatchMessage) (*WatchMessage, error) {
	c := make(chan *WatchMessage, 1)
	wc.MessageChan.Store(req.Seq, c)
	defer func() {
		wc.MessageChan.Delete(req.Seq)
		close(c)
	}()

	wc.Wm.Lock()
	err := WriteWatchMessage(wc.Writer, req)
	if err != nil {
		wc.Wm.Unlock()
		return nil, err
	}
	wc.Wm.Unlock()

	return readChan(c), nil
}

func (wc *WatchClient) RawCall(param []string) []string {
	msg, err := wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    param[0],
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: param[1:],
	})

	if err != nil {
		return []string{}
	}

	return msg.Values
}

// initialize 初始化watch信息
func (wc *WatchClient) initialize(keys []string) {
	wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    "initialize",
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: wc.Keys,
	})
}

// heartbeat 心跳
func (wc *WatchClient) heartbeat() {
	wc.SendAndRecvMessage(&WatchMessage{
		Cmd: "heartbeat",
		Seq: int(atomic.AddInt32(&wc.Seq, 1)),
	})
}

// Hget 客户端请求key,field信息
func (wc *WatchClient) Hget(key string, field string) string {
	msg, err := wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    "hget",
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: []string{key, field},
	})

	if err != nil {
		return ""
	}

	return msg.Values[0]
}

// Hgetall 客户端请求key信息
func (wc *WatchClient) Hgetall(key string) []string {
	msg, err := wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    "hgetall",
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: []string{key},
	})

	if err != nil {
		return []string{}
	}

	return msg.Values
}

// Hset 设置key信息
func (wc *WatchClient) Hset(key string, field string, value string) {
	wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    "hset",
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: []string{key, field, value},
	})
}

// Hset 设置key信息
func (wc *WatchClient) KeyPrefix(prefix string) []string {
	msg, err := wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    "key_prefix",
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: []string{prefix},
	})
	if err != nil {
		return []string{}
	}
	return msg.Values
}

func (wc *WatchClient) Del(key string) {
	wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    "del",
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: []string{key},
	})
}

func (wc *WatchClient) Hincrby(key, field string, add int) int {
	msg, err := wc.SendAndRecvMessage(&WatchMessage{
		Cmd:    "hincrby",
		Seq:    int(atomic.AddInt32(&wc.Seq, 1)),
		Values: []string{key, field, strconv.Itoa(add)},
	})

	if err != nil {
		return 0
	}

	value, _ := strconv.Atoi(msg.Values[0])
	return value
}

// Watch 监听关心key的信息
func (wc *WatchClient) Watch() (string, string, string) {
	rsp := <-wc.WatchChan
	return rsp.Values[0], rsp.Values[1], rsp.Values[2]
}

//readChan 读取房间返回，1分钟超时
func readChan(c chan *WatchMessage) *WatchMessage {
	t := time.NewTimer(10 * time.Second)
	defer t.Stop()
	select {
	case ret := <-c:
		return ret
	case <-t.C:
		return nil
	}
}

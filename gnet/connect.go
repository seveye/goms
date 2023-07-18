package gnet

import (
	"bufio"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/seveye/goms/rpc"

	"github.com/seveye/goms/util"
	"google.golang.org/protobuf/proto"
)

// ClientConn 客户端连接信息
type ClientConn struct {
	Mutex      sync.Mutex
	Rwc        io.ReadWriteCloser
	Context    *rpc.Context
	Uid        string
	CommonNode string //通用逻辑节点
	GameNode   string //游戏逻辑节点
	T          *time.Timer
	App        string
	Mask       uint8
	RpcClient  *rpc.Client
}

type InterceptorNewFunc func(*ClientConn) error                                  //新连接
type InterceptorReqFunc func(*ClientConn, *Request) (uint16, error)              //新消息
type InterceptorRspFunc func(*ClientConn, *Request, []byte, uint16, error) error //服务端回复

type Gater struct {
	Id        int
	IDCounter uint64 //连接id
	Ids       sync.Map
	Uids      sync.Map
	SyncCall  bool
	Count     int32

	NewFuncs []InterceptorNewFunc //新连接拦截器列表
	ReqFuncs []InterceptorReqFunc //新消息拦截器列表
	RspFuncs []InterceptorRspFunc //服务端回复拦截器列表
	DelFuncs []InterceptorNewFunc //连接断开拦截器列表
}

type GaterOption func(*Gater)

var (
	gater *Gater
)

// NewGater 初始化网关
func NewGater(syncCall bool, options ...GaterOption) *Gater {
	gater = &Gater{
		IDCounter: 1,
		SyncCall:  syncCall,
	}

	for _, o := range options {
		o(gater)
	}
	return gater
}

// RegisterNewFuncs 注册新连接拦截器
func (g *Gater) RegisterNewFuncs(f InterceptorNewFunc) {
	g.NewFuncs = append(g.NewFuncs, f)
}

// RegisterReqFuncs 注册新请求拦截器
func (g *Gater) RegisterReqFuncs(f InterceptorReqFunc) {
	g.ReqFuncs = append(g.ReqFuncs, f)
}

// RegisterRspFuncs 注册收到响应拦截器
func (g *Gater) RegisterRspFuncs(f InterceptorRspFunc) {
	g.RspFuncs = append(g.RspFuncs, f)
}

// RegisterDelFuncs 注册收到连接断开拦截器
func (g *Gater) RegisterDelFuncs(f InterceptorNewFunc) {
	g.DelFuncs = append(g.DelFuncs, f)
}

// newID 生成一个连接id，前两个字节保存节点id，后四个字节递增
func (g *Gater) newID() uint64 {
	for {
		id := atomic.AddUint64(&g.IDCounter, 1)
		id = id&0xFFFFFFFF | (rpc.NodeIntance.Id&0xFFFF)<<32
		if _, ok := g.Ids.Load(id); !ok {
			return id
		}
	}
}

// GetConnByUid 根据uid返回连接
func (g *Gater) GetConnByUid(uid interface{}) *ClientConn {
	v, ok := g.Uids.Load(uid)
	if !ok {
		return nil
	}
	return v.(*ClientConn)
}

// GetConnByID 根据id返回连接
func (g *Gater) GetConnByID(id uint64) *ClientConn {
	v, ok := g.Ids.Load(id)
	if !ok {
		return nil
	}
	return v.(*ClientConn)
}

// NewClientConn 初始化一个新链接
func (g *Gater) NewClientConn(remote string, rwc io.ReadWriteCloser) *ClientConn {
	conn := &ClientConn{
		Rwc: rwc,
		T:   time.NewTimer(30 * time.Second),
		Context: &rpc.Context{
			Remote:   remote,
			Id:       g.newID(),
			GateName: rpc.NodeIntance.Name,
		},
	}

	atomic.AddInt32(&g.Count, 1)
	g.Ids.Store(conn.Context.Id, conn)
	return conn
}

// DelClientConn 销毁一个链接
func (g *Gater) DelClientConn(conn *ClientConn) {
	atomic.AddInt32(&g.Count, -1)
	g.Ids.Delete(conn.Context.Id)
}

// GetCount ...
func (g *Gater) GetCount() int32 {
	return atomic.LoadInt32(&g.Count)
}

func (conn *ClientConn) SendMessage(msg *Request) {
	msg.App = conn.App
	msg.Mask = conn.Mask
	util.Submit(
		func() {
			conn.Mutex.Lock()
			defer conn.Mutex.Unlock()
			err := WriteMessage(conn.Rwc, msg)
			if err != nil {
				// api.Debug("writeMessage error", "remote", conn.Context.Remote, "uid", conn.ConnInfo.Uid, "err", err)
				conn.Rwc.Close()
				return

			}
			if msg.Disconn {
				util.Debug("断开连接", "remote", conn.Context.Remote, "uid", conn.Context.Uid)
				conn.Rwc.Close()
			}
		},
	)
}

// handleConn send back everything it received
func (g *Gater) handleConn(rwc io.ReadWriteCloser, remote string, connType string) {
	r := bufio.NewReader(rwc)
	remote = strings.Split(remote, ":")[0]
	conn := g.NewClientConn(remote, rwc)
	defer func() {
		util.Trace("连接断开", "remote", remote, "connType", connType, "id", conn.Context.Id)
		g.DelClientConn(conn)
		rwc.Close()

		for _, v := range g.DelFuncs {
			v(conn)
		}
	}()
	util.Trace("新连接", "remote", remote, "connType", connType, "id", conn.Context.Id)

	//新连接拦截器
	for _, v := range g.NewFuncs {
		if err := v(conn); err != nil {
			return
		}
	}

	//连接心跳定时器
	util.Submit(func() {
		for range conn.T.C {
			rwc.Close()
			conn.T.Stop()
			return
		}
	})

	//读取数据逻辑
	for {
	LOOP:
		//读取消息
		msg, err := ReadMessage(r)
		if err != nil {
			util.Debug("readMessage error", "remote", conn.Context.Remote, "uid", conn.Context.Uid, "error", err)
			return
		}

		if msg.App == "" {
			conn.SendMessage(&Request{Cmd: msg.Cmd, Seq: msg.Seq, Ret: 1})
			goto LOOP
		} else if conn.App != "" && conn.App != msg.App {
			conn.SendMessage(&Request{Cmd: msg.Cmd, Seq: msg.Seq, Ret: 1})
			goto LOOP
		} else if conn.App == "" {
			conn.App = msg.App
			conn.Mask = msg.Mask

			var isJson int64
			if msg.isJson() {
				isJson = 1
			}

			conn.Context.Ps = map[string]int64{
				"isJson": isJson,
			}
		}

		//新请求拦截器
		for _, v := range g.ReqFuncs {
			if ret, _ := v(conn, msg); ret != 0 {
				conn.SendMessage(&Request{Cmd: msg.Cmd, Seq: msg.Seq, Ret: uint16(ret)})
				goto LOOP
			}
		}

		//重置定时器
		conn.T.Reset(120 * time.Second)
		//转发消息
		serverMethod := rpc.FindServiceMethod(msg.NodeType, msg.Cmd)
		if serverMethod != "" {
			f := func() {
				var (
					ret     uint16
					rspBuff []byte
					retErr  error
					rpcConn = proto.Clone(conn.Context).(*rpc.Context)
				)
				rpcConn.CallId = rpc.NewId(0)
				rpcConn.Game = msg.App

				if msg.NodeName != "" {
					if msg.isJson() {
						ret, rspBuff, retErr = rpc.NodeJsonCallWithConn(rpcConn, msg.NodeName, serverMethod, msg.Buff)
					} else {
						ret, rspBuff, retErr = rpc.NodeRawCallWithConn(rpcConn, msg.NodeName, serverMethod, msg.Buff)
					}
				} else {
					if msg.isJson() {
						ret, rspBuff, retErr = rpc.JsonCall(rpcConn, serverMethod, msg.Buff)
					} else {
						ret, rspBuff, retErr = rpc.RawCall(rpcConn, serverMethod, msg.Buff)
					}
				}

				if retErr != nil && ret == 0 {
					ret = 1
				}

				//响应拦截器
				for _, v := range g.RspFuncs {
					v(conn, msg, rspBuff, ret, retErr)
				}

				conn.SendMessage(&Request{
					Cmd:  msg.Cmd,
					Seq:  msg.Seq,
					Ret:  uint16(ret),
					Buff: rspBuff,
				})
			}
			if g.SyncCall {
				f()
			} else {
				util.Submit(f)
			}
		} else {
			conn.SendMessage(&Request{
				Cmd: msg.Cmd,
				Seq: msg.Seq,
				Ret: 2,
			})
			util.Info("connect cmd error", "cmd", msg.Cmd)
		}
	}
}

// ForConn 遍历所有链接
func (g *Gater) ForConn(f func(conn *ClientConn)) {
	var ids []uint64
	g.Ids.Range(func(k, v interface{}) bool {
		ids = append(ids, k.(uint64))
		return true
	})
	for i := 0; i < len(ids); i++ {
		wsconn := g.GetConnByID(ids[i])
		if wsconn == nil {
			continue
		}
		f(wsconn)
	}
}

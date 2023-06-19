// Copyright 2017 guangbo. All rights reserved.

// rpc编码解码模块
package rpc

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/seveye/goms/util/bytes_cache"
)

// tooBig 内部通讯最大消息长度 1G
const tooBig = 1 << 30

var errBadCount = errors.New("invalid message length")

func writeFrame(w *bufio.Writer, buf []byte) error {
	l := len(buf)
	if l >= tooBig {
		return errBadCount
	}

	lenBuf := bytes_cache.Get(4)
	defer bytes_cache.Put(lenBuf)
	binary.BigEndian.PutUint32(lenBuf[:], uint32(l))
	_, err := w.Write(lenBuf[:])
	if err != nil {
		return err
	}
	_, err = w.Write(buf)
	return err
}

func readFrame(r *bufio.Reader) ([]byte, error) {
	header := bytes_cache.Get(4)
	defer bytes_cache.Put(header)
	_, err := io.ReadFull(r, header)
	if err != nil {
		return nil, err
	}

	l := binary.BigEndian.Uint32(header)
	if l >= tooBig {
		return nil, errBadCount
	}

	buff := bytes_cache.Get(int(l))
	_, err = io.ReadFull(r, buff)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func encode(w *bufio.Writer, raw int32, m interface{}) error {
	if pb, ok := m.(proto.Message); ok {
		if raw == 2 {
			buf, err := json.Marshal(pb)
			if err != nil {
				return err
			}
			return writeFrame(w, buf)
		}

		buf, err := proto.Marshal(pb)
		if err != nil {
			return err
		}
		return writeFrame(w, buf)
	}
	return fmt.Errorf("%T does not implement proto.Message", m)
}

func decode(r *bufio.Reader, raw int32, m interface{}) error {
	buff, err := readFrame(r)
	if err != nil {
		return err
	}

	if m == nil {
		return nil
	}

	if raw == 2 {
		return json.Unmarshal(buff, m)
		// return jsonpb.UnmarshalString(string(buff), m.(proto.Message))
	}

	return proto.Unmarshal(buff, m.(proto.Message))
}

type PbClientCodec struct {
	req  ReqHeader
	resp RspHeader
	c    io.ReadWriteCloser
	w    *bufio.Writer
	r    *bufio.Reader
}

func NewPbClientCodec(rwc io.ReadWriteCloser) ClientCodec {
	return &PbClientCodec{
		r: bufio.NewReaderSize(rwc, 4096),
		w: bufio.NewWriterSize(rwc, 4096),
		c: rwc,
	}
}

func (c *PbClientCodec) WriteRequest(r *Request, body interface{}) error {
	c.req.Reset()
	c.req.Method = r.ServiceMethod
	c.req.Seq = r.Seq
	c.req.NoResp = r.NoResp
	c.req.Raw = int32(r.Raw)
	if r.Conn != nil {
		c.req.Context = &Context{
			GateName: r.Conn.GateName,
			Remote:   r.Conn.Remote,
			Id:       r.Conn.Id,
			Uid:      r.Conn.Uid,
			Token:    r.Conn.Token,
			CallId:   r.Conn.CallId,
			Kvs:      r.Conn.Kvs,
			Ps:       r.Conn.Ps,
			Game:     r.Conn.Game,
		}
	}
	err := encode(c.w, 0, &c.req)
	if err != nil {
		return err
	}
	if err = encode(c.w, 0, body); err != nil {
		return err
	}
	return c.w.Flush()
}

func (c *PbClientCodec) ReadResponseHeader(r *Response) error {
	c.resp.Reset()
	err := decode(c.r, 0, &c.resp)
	if err != nil {
		return err
	}
	r.ServiceMethod = c.resp.Method
	r.Seq = c.resp.Seq
	r.Error = c.resp.Error
	r.Ret = uint16(c.resp.Ret)
	return nil
}

func (c *PbClientCodec) ReadResponseBody(raw int32, body interface{}) error {
	return decode(c.r, raw, body)
}

func (c *PbClientCodec) WriteByteRequest(r *Request, buf []byte) error {
	c.req.Reset()
	c.req.Method = r.ServiceMethod
	c.req.Seq = r.Seq
	c.req.Raw = int32(r.Raw)
	if r.Conn != nil {
		c.req.Context = &Context{
			GateName: r.Conn.GateName,
			Remote:   r.Conn.Remote,
			Id:       r.Conn.Id,
			Uid:      r.Conn.Uid,
			Token:    r.Conn.Token,
			CallId:   r.Conn.CallId,
			Kvs:      r.Conn.Kvs,
			Ps:       r.Conn.Ps,
			Game:     r.Conn.Game,
		}
	}

	err := encode(c.w, 0, &c.req)
	if err != nil {
		return err
	}
	if err = writeFrame(c.w, buf); err != nil {
		return err
	}
	return c.w.Flush()
}

func (c *PbClientCodec) ReadByteResponseBody() ([]byte, error) {
	return readFrame(c.r)
}

func (c *PbClientCodec) Close() error {
	return c.c.Close()
}

type PbServerCodec struct {
	mu   sync.Mutex // exclusive writer lock
	req  ReqHeader
	resp RspHeader
	w    *bufio.Writer
	r    *bufio.Reader

	c io.Closer
}

func NewPbServerCodec(rwc io.ReadWriteCloser) ServerCodec {
	return &PbServerCodec{
		r: bufio.NewReaderSize(rwc, 4096),
		w: bufio.NewWriterSize(rwc, 4096),
		c: rwc,
	}
}

func (c *PbServerCodec) WriteResponse(resp *Response, body interface{}) error {
	c.mu.Lock()
	c.resp.Method = resp.ServiceMethod
	c.resp.Seq = resp.Seq
	c.resp.Error = resp.Error
	c.resp.Ret = uint32(resp.Ret)
	c.resp.Raw = int32(resp.Raw)

	err := encode(c.w, 0, &c.resp)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	if err = encode(c.w, c.resp.Raw, body); err != nil {
		c.mu.Unlock()
		return err
	}
	err = c.w.Flush()
	c.mu.Unlock()
	return err
}

func (c *PbServerCodec) ReadRequestHeader(req *Request) error {
	c.req.Reset()

	err := decode(c.r, 0, &c.req)
	if err != nil {
		return err
	}

	req.ServiceMethod = c.req.Method
	req.Seq = c.req.Seq
	req.NoResp = c.req.NoResp
	req.Raw = int(c.req.Raw)
	if c.req.Context != nil {
		req.Conn = &Context{
			GateName: c.req.Context.GateName,
			Remote:   c.req.Context.Remote,
			Id:       c.req.Context.Id,
			Uid:      c.req.Context.Uid,
			Token:    c.req.Context.Token,
			CallId:   c.req.Context.CallId,
			Kvs:      c.req.Context.Kvs,
			Ps:       c.req.Context.Ps,
			Game:     c.req.Context.Game,
		}
	}
	return nil
}

func (c *PbServerCodec) ReadRequestBody(raw int32, body interface{}) error {
	return decode(c.r, raw, body)
}

func (c *PbServerCodec) Close() error { return c.c.Close() }

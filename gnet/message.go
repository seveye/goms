package gnet

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"

	"github.com/seveye/goms/util"
	"github.com/seveye/goms/util/bytes_cache"
)

// 通讯密钥
var (
	AesKey   = []byte("12345678901234567890123456789012")
	emptyKey = []byte("")
)

// Message 请求，协议头固定10字节：
// 请求响应，协议头固定10字节：
//
//	G标志（1字节） | 掩码（1字节）| 应用id（4字节）| 长度（2字节）| cmd（2字节）| 序列号（2字节）| 返回值（2字节）| 加密数据（n字节）
type Request struct {
	App     string
	Mask    uint8
	Length  uint16
	Cmd     uint16
	Seq     uint16
	Ret     uint16
	Buff    []byte
	Disconn bool

	//所属节点数据
	NodeName string
	NodeType string
}

// isCrypto 是否加密
func (m *Request) isCrypto() bool {
	return util.GetBit(uint32(m.Mask), 0)
}

// isJson 是否json模式
func (m *Request) isJson() bool {
	return util.GetBit(uint32(m.Mask), 1)
}

// ReadMessage 通用读取请求接口
func ReadMessage(r *bufio.Reader) (*Request, error) {
	return ReadMessageWithKey(r, emptyKey)
}

// WriteMessage 通用写请求接口
func WriteMessage(conn io.ReadWriteCloser, req *Request) error {
	return WriteMessageWithKey(conn, req, emptyKey)
}

// ReadMessageWithKey 通用读取请求接口, key为通讯密钥，长度为32字节
func ReadMessageWithKey(r *bufio.Reader, key []byte) (*Request, error) {
	var (
		req    = &Request{}
		header = bytes_cache.Get(14)
	)
	defer bytes_cache.Put(header)

	_, err := io.ReadFull(r, header[:])
	if err != nil {
		return nil, err
	}

	msgFlag := uint8(header[0])
	if msgFlag != 'G' {
		return nil, errors.New("error message flag")
	}
	req.Mask = uint8(header[1])
	req.App = string(header[2:6])
	req.Length = binary.BigEndian.Uint16(header[6:8])
	req.Cmd = binary.BigEndian.Uint16(header[8:10])
	req.Seq = binary.BigEndian.Uint16(header[10:12])
	req.Ret = binary.BigEndian.Uint16(header[12:14])

	if req.Length == 0 {
		return req, nil
	}

	//生存期间不确定，暂时不用交回给缓冲池
	buff := bytes_cache.Get(int(req.Length))
	_, err = io.ReadFull(r, buff)
	if err != nil {
		return nil, err
	}

	//解密
	if len(key) == 0 && req.isCrypto() {
		key = AesKey
	}
	if len(key) > 0 {
		req.Buff, err = util.AesDecrypt(buff, key)
		if err != nil {
			return nil, err
		}
	} else {
		req.Buff = buff
	}

	return req, nil
}

// WriteMessageWithKey 通用写请求接口, key为通讯密钥，长度为32字节
func WriteMessageWithKey(conn io.ReadWriteCloser, req *Request, key []byte) error {
	var (
		buff = req.Buff
		err  error
		b    bytes.Buffer
	)
	if len(key) == 0 && req.isCrypto() {
		key = AesKey
	}
	if len(key) > 0 {
		buff, err = util.AesEncrypt(req.Buff, key)
		if err != nil {
			return err
		}
	}

	b.WriteByte('G')
	b.WriteByte(req.Mask)
	b.Write([]byte(req.App))
	binary.Write(&b, binary.BigEndian, uint16(len(buff)))
	binary.Write(&b, binary.BigEndian, req.Cmd)
	binary.Write(&b, binary.BigEndian, req.Seq)
	binary.Write(&b, binary.BigEndian, req.Ret)
	b.Write(buff)

	_, err = conn.Write(b.Bytes())
	return err
}

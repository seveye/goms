// Code generated by protoc-gen-go. DO NOT EDIT.
// source: message.proto

/*
Package proto is a generated protocol buffer package.

It is generated from these files:
	message.proto

It has these top-level messages:
	AddReq
	AddRsp
*/
package proto

import proto1 "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto1.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto1.ProtoPackageIsVersion2 // please upgrade the proto package

type AddReq struct {
	A uint64 `protobuf:"varint,1,opt,name=a" json:"a,omitempty"`
	B uint64 `protobuf:"varint,2,opt,name=b" json:"b,omitempty"`
}

func (m *AddReq) Reset()                    { *m = AddReq{} }
func (m *AddReq) String() string            { return proto1.CompactTextString(m) }
func (*AddReq) ProtoMessage()               {}
func (*AddReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AddReq) GetA() uint64 {
	if m != nil {
		return m.A
	}
	return 0
}

func (m *AddReq) GetB() uint64 {
	if m != nil {
		return m.B
	}
	return 0
}

type AddRsp struct {
	C uint64 `protobuf:"varint,1,opt,name=c" json:"c,omitempty"`
}

func (m *AddRsp) Reset()                    { *m = AddRsp{} }
func (m *AddRsp) String() string            { return proto1.CompactTextString(m) }
func (*AddRsp) ProtoMessage()               {}
func (*AddRsp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *AddRsp) GetC() uint64 {
	if m != nil {
		return m.C
	}
	return 0
}

func init() {
	proto1.RegisterType((*AddReq)(nil), "proto.AddReq")
	proto1.RegisterType((*AddRsp)(nil), "proto.AddRsp")
}

func init() { proto1.RegisterFile("message.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 91 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0xcd, 0x4d, 0x2d, 0x2e,
	0x4e, 0x4c, 0x4f, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x4a, 0x2a, 0x5c,
	0x6c, 0x8e, 0x29, 0x29, 0x41, 0xa9, 0x85, 0x42, 0x3c, 0x5c, 0x8c, 0x89, 0x12, 0x8c, 0x0a, 0x8c,
	0x1a, 0x2c, 0x41, 0x8c, 0x89, 0x20, 0x5e, 0x92, 0x04, 0x13, 0x84, 0x97, 0xa4, 0x24, 0x06, 0x51,
	0x55, 0x5c, 0x00, 0x12, 0x4f, 0x86, 0xa9, 0x4a, 0x4e, 0x62, 0x03, 0x1b, 0x62, 0x0c, 0x08, 0x00,
	0x00, 0xff, 0xff, 0xf9, 0x6d, 0xb2, 0x67, 0x5c, 0x00, 0x00, 0x00,
}
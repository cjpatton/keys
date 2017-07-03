// Code generated by protoc-gen-go. DO NOT EDIT.
// source: store.proto

/*
Package store is a generated protocol buffer package.

It is generated from these files:
	store.proto

It has these top-level messages:
	StoreParams
	StoreTable
	HelloRequest
	HelloReply
*/
package store

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type StoreParams struct {
	TableLen       int32  `protobuf:"varint,1,opt,name=table_len,json=tableLen" json:"table_len,omitempty"`
	MaxOutputBytes int32  `protobuf:"varint,2,opt,name=max_output_bytes,json=maxOutputBytes" json:"max_output_bytes,omitempty"`
	RowBytes       int32  `protobuf:"varint,3,opt,name=row_bytes,json=rowBytes" json:"row_bytes,omitempty"`
	TagBytes       int32  `protobuf:"varint,4,opt,name=tag_bytes,json=tagBytes" json:"tag_bytes,omitempty"`
	SaltBytes      int32  `protobuf:"varint,5,opt,name=salt_bytes,json=saltBytes" json:"salt_bytes,omitempty"`
	Salt           []byte `protobuf:"bytes,6,opt,name=salt,proto3" json:"salt,omitempty"`
}

func (m *StoreParams) Reset()                    { *m = StoreParams{} }
func (m *StoreParams) String() string            { return proto.CompactTextString(m) }
func (*StoreParams) ProtoMessage()               {}
func (*StoreParams) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *StoreParams) GetTableLen() int32 {
	if m != nil {
		return m.TableLen
	}
	return 0
}

func (m *StoreParams) GetMaxOutputBytes() int32 {
	if m != nil {
		return m.MaxOutputBytes
	}
	return 0
}

func (m *StoreParams) GetRowBytes() int32 {
	if m != nil {
		return m.RowBytes
	}
	return 0
}

func (m *StoreParams) GetTagBytes() int32 {
	if m != nil {
		return m.TagBytes
	}
	return 0
}

func (m *StoreParams) GetSaltBytes() int32 {
	if m != nil {
		return m.SaltBytes
	}
	return 0
}

func (m *StoreParams) GetSalt() []byte {
	if m != nil {
		return m.Salt
	}
	return nil
}

type StoreTable struct {
	Params *StoreParams `protobuf:"bytes,1,opt,name=params" json:"params,omitempty"`
	Table  []byte       `protobuf:"bytes,2,opt,name=table,proto3" json:"table,omitempty"`
	Idx    []int32      `protobuf:"varint,3,rep,packed,name=idx" json:"idx,omitempty"`
}

func (m *StoreTable) Reset()                    { *m = StoreTable{} }
func (m *StoreTable) String() string            { return proto.CompactTextString(m) }
func (*StoreTable) ProtoMessage()               {}
func (*StoreTable) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *StoreTable) GetParams() *StoreParams {
	if m != nil {
		return m.Params
	}
	return nil
}

func (m *StoreTable) GetTable() []byte {
	if m != nil {
		return m.Table
	}
	return nil
}

func (m *StoreTable) GetIdx() []int32 {
	if m != nil {
		return m.Idx
	}
	return nil
}

// The request message containing the user's name.
type HelloRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *HelloRequest) Reset()                    { *m = HelloRequest{} }
func (m *HelloRequest) String() string            { return proto.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()               {}
func (*HelloRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

// The response message containing the greetings
type HelloReply struct {
	Message string `protobuf:"bytes,1,opt,name=message" json:"message,omitempty"`
}

func (m *HelloReply) Reset()                    { *m = HelloReply{} }
func (m *HelloReply) String() string            { return proto.CompactTextString(m) }
func (*HelloReply) ProtoMessage()               {}
func (*HelloReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *HelloReply) GetMessage() string {
	if m != nil {
		return m.Message
	}
	return ""
}

func init() {
	proto.RegisterType((*StoreParams)(nil), "store.StoreParams")
	proto.RegisterType((*StoreTable)(nil), "store.StoreTable")
	proto.RegisterType((*HelloRequest)(nil), "store.HelloRequest")
	proto.RegisterType((*HelloReply)(nil), "store.HelloReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Greeter service

type GreeterClient interface {
	// Sends a greeting
	SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error)
}

type greeterClient struct {
	cc *grpc.ClientConn
}

func NewGreeterClient(cc *grpc.ClientConn) GreeterClient {
	return &greeterClient{cc}
}

func (c *greeterClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloReply, error) {
	out := new(HelloReply)
	err := grpc.Invoke(ctx, "/store.Greeter/SayHello", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Greeter service

type GreeterServer interface {
	// Sends a greeting
	SayHello(context.Context, *HelloRequest) (*HelloReply, error)
}

func RegisterGreeterServer(s *grpc.Server, srv GreeterServer) {
	s.RegisterService(&_Greeter_serviceDesc, srv)
}

func _Greeter_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(GreeterServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/store.Greeter/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(GreeterServer).SayHello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Greeter_serviceDesc = grpc.ServiceDesc{
	ServiceName: "store.Greeter",
	HandlerType: (*GreeterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Greeter_SayHello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "store.proto",
}

func init() { proto.RegisterFile("store.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 298 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x51, 0x4d, 0x4b, 0xf4, 0x30,
	0x10, 0x7e, 0xfb, 0x76, 0xbb, 0x1f, 0xb3, 0x45, 0xd6, 0xe8, 0xa1, 0x28, 0x42, 0xc9, 0x41, 0x8a,
	0x87, 0x3d, 0xac, 0xde, 0x05, 0x2f, 0x7a, 0x10, 0x94, 0xac, 0xf7, 0x9a, 0xe2, 0x50, 0x84, 0xb4,
	0xa9, 0x49, 0xca, 0xb6, 0x3f, 0xcf, 0x7f, 0x26, 0x99, 0xb6, 0xb2, 0xde, 0xe6, 0xf9, 0xc8, 0xcc,
	0xf3, 0x10, 0x58, 0x5b, 0xa7, 0x0d, 0x6e, 0x1b, 0xa3, 0x9d, 0x66, 0x11, 0x01, 0xfe, 0x1d, 0xc0,
	0x7a, 0xef, 0xa7, 0x57, 0x69, 0x64, 0x65, 0xd9, 0x25, 0xac, 0x9c, 0x2c, 0x14, 0xe6, 0x0a, 0xeb,
	0x24, 0x48, 0x83, 0x2c, 0x12, 0x4b, 0x22, 0x9e, 0xb1, 0x66, 0x19, 0x6c, 0x2a, 0xd9, 0xe5, 0xba,
	0x75, 0x4d, 0xeb, 0xf2, 0xa2, 0x77, 0x68, 0x93, 0xff, 0xe4, 0x39, 0xa9, 0x64, 0xf7, 0x42, 0xf4,
	0x83, 0x67, 0xfd, 0x1a, 0xa3, 0x0f, 0xa3, 0x25, 0x1c, 0xd6, 0x18, 0x7d, 0xf8, 0x15, 0x9d, 0x2c,
	0x47, 0x71, 0x36, 0xdd, 0x28, 0x07, 0xf1, 0x0a, 0xc0, 0x4a, 0x35, 0x6d, 0x8f, 0x48, 0x5d, 0x79,
	0x66, 0x90, 0x19, 0xcc, 0x3c, 0x48, 0xe6, 0x69, 0x90, 0xc5, 0x82, 0x66, 0xfe, 0x0e, 0x40, 0x15,
	0xde, 0x7c, 0x4e, 0x76, 0x03, 0xf3, 0x86, 0xba, 0x50, 0xfc, 0xf5, 0x8e, 0x6d, 0x87, 0xda, 0x47,
	0x2d, 0xc5, 0xe8, 0x60, 0xe7, 0x10, 0x51, 0x39, 0x6a, 0x11, 0x8b, 0x01, 0xb0, 0x0d, 0x84, 0x9f,
	0x1f, 0x5d, 0x12, 0xa6, 0x61, 0x16, 0x09, 0x3f, 0x72, 0x0e, 0xf1, 0x13, 0x2a, 0xa5, 0x05, 0x7e,
	0xb5, 0x68, 0x9d, 0x4f, 0x51, 0xcb, 0x0a, 0xe9, 0xc2, 0x4a, 0xd0, 0xcc, 0xaf, 0x01, 0x46, 0x4f,
	0xa3, 0x7a, 0x96, 0xc0, 0xa2, 0x42, 0x6b, 0x65, 0x39, 0x99, 0x26, 0xb8, 0xbb, 0x87, 0xc5, 0xa3,
	0x41, 0x74, 0x68, 0xd8, 0x1d, 0x2c, 0xf7, 0xb2, 0xa7, 0x57, 0xec, 0x6c, 0x8c, 0x79, 0x7c, 0xe7,
	0xe2, 0xf4, 0x2f, 0xd9, 0xa8, 0x9e, 0xff, 0x2b, 0xe6, 0xf4, 0x81, 0xb7, 0x3f, 0x01, 0x00, 0x00,
	0xff, 0xff, 0xb8, 0x30, 0x31, 0xdd, 0xcf, 0x01, 0x00, 0x00,
}

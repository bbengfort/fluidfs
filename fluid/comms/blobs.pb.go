// Code generated by protoc-gen-go.
// source: blobs.proto
// DO NOT EDIT!

/*
Package comms is a generated protocol buffer package.

It is generated from these files:
	blobs.proto

It has these top-level messages:
	PushRequest
	PushReply
*/
package comms

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

// PushRequest initiates the first phase of Anti-Entropy by letting another
// node in the system know the current state of the blob tree.
type PushRequest struct {
	Count uint64 `protobuf:"varint,1,opt,name=count" json:"count,omitempty"`
	Size  uint64 `protobuf:"varint,2,opt,name=size" json:"size,omitempty"`
}

func (m *PushRequest) Reset()                    { *m = PushRequest{} }
func (m *PushRequest) String() string            { return proto.CompactTextString(m) }
func (*PushRequest) ProtoMessage()               {}
func (*PushRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *PushRequest) GetCount() uint64 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *PushRequest) GetSize() uint64 {
	if m != nil {
		return m.Size
	}
	return 0
}

// PushReply allows the remote node to respond with a list of blobs to request
// from the initiating server.
type PushReply struct {
	Sync bool `protobuf:"varint,1,opt,name=sync" json:"sync,omitempty"`
}

func (m *PushReply) Reset()                    { *m = PushReply{} }
func (m *PushReply) String() string            { return proto.CompactTextString(m) }
func (*PushReply) ProtoMessage()               {}
func (*PushReply) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *PushReply) GetSync() bool {
	if m != nil {
		return m.Sync
	}
	return false
}

func init() {
	proto.RegisterType((*PushRequest)(nil), "comms.PushRequest")
	proto.RegisterType((*PushReply)(nil), "comms.PushReply")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for BlobTalk service

type BlobTalkClient interface {
	PushHandler(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushReply, error)
}

type blobTalkClient struct {
	cc *grpc.ClientConn
}

func NewBlobTalkClient(cc *grpc.ClientConn) BlobTalkClient {
	return &blobTalkClient{cc}
}

func (c *blobTalkClient) PushHandler(ctx context.Context, in *PushRequest, opts ...grpc.CallOption) (*PushReply, error) {
	out := new(PushReply)
	err := grpc.Invoke(ctx, "/comms.BlobTalk/PushHandler", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for BlobTalk service

type BlobTalkServer interface {
	PushHandler(context.Context, *PushRequest) (*PushReply, error)
}

func RegisterBlobTalkServer(s *grpc.Server, srv BlobTalkServer) {
	s.RegisterService(&_BlobTalk_serviceDesc, srv)
}

func _BlobTalk_PushHandler_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PushRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BlobTalkServer).PushHandler(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/comms.BlobTalk/PushHandler",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BlobTalkServer).PushHandler(ctx, req.(*PushRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _BlobTalk_serviceDesc = grpc.ServiceDesc{
	ServiceName: "comms.BlobTalk",
	HandlerType: (*BlobTalkServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PushHandler",
			Handler:    _BlobTalk_PushHandler_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "blobs.proto",
}

func init() { proto.RegisterFile("blobs.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 156 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4e, 0xca, 0xc9, 0x4f,
	0x2a, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x4d, 0xce, 0xcf, 0xcd, 0x2d, 0x56, 0x32,
	0xe7, 0xe2, 0x0e, 0x28, 0x2d, 0xce, 0x08, 0x4a, 0x2d, 0x2c, 0x4d, 0x2d, 0x2e, 0x11, 0x12, 0xe1,
	0x62, 0x4d, 0xce, 0x2f, 0xcd, 0x2b, 0x91, 0x60, 0x54, 0x60, 0xd4, 0x60, 0x09, 0x82, 0x70, 0x84,
	0x84, 0xb8, 0x58, 0x8a, 0x33, 0xab, 0x52, 0x25, 0x98, 0xc0, 0x82, 0x60, 0xb6, 0x92, 0x3c, 0x17,
	0x27, 0x44, 0x63, 0x41, 0x4e, 0x25, 0x58, 0x41, 0x65, 0x5e, 0x32, 0x58, 0x17, 0x47, 0x10, 0x98,
	0x6d, 0xe4, 0xc8, 0xc5, 0xe1, 0x94, 0x93, 0x9f, 0x14, 0x92, 0x98, 0x93, 0x2d, 0x64, 0x0a, 0xb1,
	0xc5, 0x23, 0x31, 0x2f, 0x25, 0x27, 0xb5, 0x48, 0x48, 0x48, 0x0f, 0x6c, 0xb9, 0x1e, 0x92, 0xcd,
	0x52, 0x02, 0x28, 0x62, 0x05, 0x39, 0x95, 0x4a, 0x0c, 0x49, 0x6c, 0x60, 0xa7, 0x1a, 0x03, 0x02,
	0x00, 0x00, 0xff, 0xff, 0x10, 0x70, 0x38, 0x81, 0xb9, 0x00, 0x00, 0x00,
}
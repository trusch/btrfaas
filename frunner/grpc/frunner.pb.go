// Code generated by protoc-gen-go. DO NOT EDIT.
// source: frunner.proto

/*
Package grpc is a generated protocol buffer package.

It is generated from these files:
	frunner.proto

It has these top-level messages:
	FrunnerInputData
	FrunnerOutputData
*/
package grpc

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc1 "google.golang.org/grpc"
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

type FrunnerInputData struct {
	Data    []byte            `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	Options map[string]string `protobuf:"bytes,2,rep,name=options" json:"options,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *FrunnerInputData) Reset()                    { *m = FrunnerInputData{} }
func (m *FrunnerInputData) String() string            { return proto.CompactTextString(m) }
func (*FrunnerInputData) ProtoMessage()               {}
func (*FrunnerInputData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *FrunnerInputData) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

func (m *FrunnerInputData) GetOptions() map[string]string {
	if m != nil {
		return m.Options
	}
	return nil
}

type FrunnerOutputData struct {
	// if true -> last OutputData packet
	Ready bool `protobuf:"varint,1,opt,name=ready" json:"ready,omitempty"`
	// false if error while execution
	Success bool `protobuf:"varint,2,opt,name=success" json:"success,omitempty"`
	// eventually error message in case of !success
	ErrorMessage string `protobuf:"bytes,3,opt,name=errorMessage" json:"errorMessage,omitempty"`
	// streaming output data (stdout)
	Output []byte `protobuf:"bytes,4,opt,name=output,proto3" json:"output,omitempty"`
}

func (m *FrunnerOutputData) Reset()                    { *m = FrunnerOutputData{} }
func (m *FrunnerOutputData) String() string            { return proto.CompactTextString(m) }
func (*FrunnerOutputData) ProtoMessage()               {}
func (*FrunnerOutputData) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *FrunnerOutputData) GetReady() bool {
	if m != nil {
		return m.Ready
	}
	return false
}

func (m *FrunnerOutputData) GetSuccess() bool {
	if m != nil {
		return m.Success
	}
	return false
}

func (m *FrunnerOutputData) GetErrorMessage() string {
	if m != nil {
		return m.ErrorMessage
	}
	return ""
}

func (m *FrunnerOutputData) GetOutput() []byte {
	if m != nil {
		return m.Output
	}
	return nil
}

func init() {
	proto.RegisterType((*FrunnerInputData)(nil), "grpc.FrunnerInputData")
	proto.RegisterType((*FrunnerOutputData)(nil), "grpc.FrunnerOutputData")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc1.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc1.SupportPackageIsVersion4

// Client API for FunctionRunner service

type FunctionRunnerClient interface {
	Run(ctx context.Context, opts ...grpc1.CallOption) (FunctionRunner_RunClient, error)
}

type functionRunnerClient struct {
	cc *grpc1.ClientConn
}

func NewFunctionRunnerClient(cc *grpc1.ClientConn) FunctionRunnerClient {
	return &functionRunnerClient{cc}
}

func (c *functionRunnerClient) Run(ctx context.Context, opts ...grpc1.CallOption) (FunctionRunner_RunClient, error) {
	stream, err := grpc1.NewClientStream(ctx, &_FunctionRunner_serviceDesc.Streams[0], c.cc, "/grpc.FunctionRunner/Run", opts...)
	if err != nil {
		return nil, err
	}
	x := &functionRunnerRunClient{stream}
	return x, nil
}

type FunctionRunner_RunClient interface {
	Send(*FrunnerInputData) error
	Recv() (*FrunnerOutputData, error)
	grpc1.ClientStream
}

type functionRunnerRunClient struct {
	grpc1.ClientStream
}

func (x *functionRunnerRunClient) Send(m *FrunnerInputData) error {
	return x.ClientStream.SendMsg(m)
}

func (x *functionRunnerRunClient) Recv() (*FrunnerOutputData, error) {
	m := new(FrunnerOutputData)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for FunctionRunner service

type FunctionRunnerServer interface {
	Run(FunctionRunner_RunServer) error
}

func RegisterFunctionRunnerServer(s *grpc1.Server, srv FunctionRunnerServer) {
	s.RegisterService(&_FunctionRunner_serviceDesc, srv)
}

func _FunctionRunner_Run_Handler(srv interface{}, stream grpc1.ServerStream) error {
	return srv.(FunctionRunnerServer).Run(&functionRunnerRunServer{stream})
}

type FunctionRunner_RunServer interface {
	Send(*FrunnerOutputData) error
	Recv() (*FrunnerInputData, error)
	grpc1.ServerStream
}

type functionRunnerRunServer struct {
	grpc1.ServerStream
}

func (x *functionRunnerRunServer) Send(m *FrunnerOutputData) error {
	return x.ServerStream.SendMsg(m)
}

func (x *functionRunnerRunServer) Recv() (*FrunnerInputData, error) {
	m := new(FrunnerInputData)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _FunctionRunner_serviceDesc = grpc1.ServiceDesc{
	ServiceName: "grpc.FunctionRunner",
	HandlerType: (*FunctionRunnerServer)(nil),
	Methods:     []grpc1.MethodDesc{},
	Streams: []grpc1.StreamDesc{
		{
			StreamName:    "Run",
			Handler:       _FunctionRunner_Run_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "frunner.proto",
}

func init() { proto.RegisterFile("frunner.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 265 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x51, 0xcb, 0x4a, 0x03, 0x31,
	0x14, 0x35, 0x33, 0xd3, 0x87, 0xd7, 0x51, 0xea, 0xa5, 0xd4, 0xd0, 0x55, 0x19, 0x37, 0xb3, 0x1a,
	0xa4, 0x6e, 0xa4, 0xe8, 0x4e, 0x0b, 0x2e, 0xb4, 0x90, 0x3f, 0x88, 0xd3, 0x58, 0x44, 0x49, 0x86,
	0x3c, 0x84, 0xae, 0xfc, 0x16, 0xff, 0x54, 0xe6, 0xa6, 0x83, 0x56, 0xba, 0xbb, 0xe7, 0x24, 0xe7,
	0x91, 0x5c, 0x38, 0x7d, 0xb5, 0x41, 0x6b, 0x65, 0xab, 0xc6, 0x1a, 0x6f, 0x30, 0xdb, 0xd8, 0xa6,
	0x2e, 0xbe, 0x19, 0x8c, 0x96, 0x91, 0x7f, 0xd4, 0x4d, 0xf0, 0xf7, 0xd2, 0x4b, 0x44, 0xc8, 0xd6,
	0xd2, 0x4b, 0xce, 0x66, 0xac, 0xcc, 0x05, 0xcd, 0x78, 0x07, 0x03, 0xd3, 0xf8, 0x37, 0xa3, 0x1d,
	0x4f, 0x66, 0x69, 0x79, 0x32, 0xbf, 0xac, 0x5a, 0x83, 0xea, 0xbf, 0xb8, 0x5a, 0xc5, 0x5b, 0x0f,
	0xda, 0xdb, 0xad, 0xe8, 0x34, 0xd3, 0x05, 0xe4, 0x7f, 0x0f, 0x70, 0x04, 0xe9, 0xbb, 0xda, 0x52,
	0xc2, 0xb1, 0x68, 0x47, 0x1c, 0x43, 0xef, 0x53, 0x7e, 0x04, 0xc5, 0x13, 0xe2, 0x22, 0x58, 0x24,
	0x37, 0xac, 0xf8, 0x82, 0xf3, 0x5d, 0xca, 0x2a, 0xf8, 0xae, 0xe3, 0x18, 0x7a, 0x56, 0xc9, 0x75,
	0xb4, 0x18, 0x8a, 0x08, 0x90, 0xc3, 0xc0, 0x85, 0xba, 0x56, 0xce, 0x91, 0xcd, 0x50, 0x74, 0x10,
	0x0b, 0xc8, 0x95, 0xb5, 0xc6, 0x3e, 0x29, 0xe7, 0xe4, 0x46, 0xf1, 0x94, 0x52, 0xf6, 0x38, 0x9c,
	0x40, 0xdf, 0x50, 0x02, 0xcf, 0xe8, 0xe5, 0x3b, 0x34, 0x7f, 0x86, 0xb3, 0x65, 0xd0, 0x75, 0x5b,
	0x5f, 0x50, 0x0f, 0xbc, 0x85, 0x54, 0x04, 0x8d, 0x93, 0xc3, 0x7f, 0x30, 0xbd, 0xd8, 0xe3, 0x7f,
	0x5b, 0x17, 0x47, 0x25, 0xbb, 0x62, 0x2f, 0x7d, 0xda, 0xc0, 0xf5, 0x4f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x39, 0xc2, 0xfc, 0x39, 0x92, 0x01, 0x00, 0x00,
}

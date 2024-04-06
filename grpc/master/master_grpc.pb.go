// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v5.26.1
// source: master.proto

package master

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// MasterClient is the client API for Master service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MasterClient interface {
	KeepMeAlive(ctx context.Context, opts ...grpc.CallOption) (Master_KeepMeAliveClient, error)
	ConfirmUpload(ctx context.Context, in *FileUploadStatus, opts ...grpc.CallOption) (*emptypb.Empty, error)
	RegisterNode(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterRequest, error)
	RequestToUpload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*HostAddress, error)
	RequestToDonwload(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadResponse, error)
}

type masterClient struct {
	cc grpc.ClientConnInterface
}

func NewMasterClient(cc grpc.ClientConnInterface) MasterClient {
	return &masterClient{cc}
}

func (c *masterClient) KeepMeAlive(ctx context.Context, opts ...grpc.CallOption) (Master_KeepMeAliveClient, error) {
	stream, err := c.cc.NewStream(ctx, &Master_ServiceDesc.Streams[0], "/master.Master/KeepMeAlive", opts...)
	if err != nil {
		return nil, err
	}
	x := &masterKeepMeAliveClient{stream}
	return x, nil
}

type Master_KeepMeAliveClient interface {
	Send(*HeartBeat) error
	CloseAndRecv() (*emptypb.Empty, error)
	grpc.ClientStream
}

type masterKeepMeAliveClient struct {
	grpc.ClientStream
}

func (x *masterKeepMeAliveClient) Send(m *HeartBeat) error {
	return x.ClientStream.SendMsg(m)
}

func (x *masterKeepMeAliveClient) CloseAndRecv() (*emptypb.Empty, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(emptypb.Empty)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *masterClient) ConfirmUpload(ctx context.Context, in *FileUploadStatus, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/master.Master/ConfirmUpload", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *masterClient) RegisterNode(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterRequest, error) {
	out := new(RegisterRequest)
	err := c.cc.Invoke(ctx, "/master.Master/RegisterNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *masterClient) RequestToUpload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*HostAddress, error) {
	out := new(HostAddress)
	err := c.cc.Invoke(ctx, "/master.Master/RequestToUpload", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *masterClient) RequestToDonwload(ctx context.Context, in *DownloadRequest, opts ...grpc.CallOption) (*DownloadResponse, error) {
	out := new(DownloadResponse)
	err := c.cc.Invoke(ctx, "/master.Master/RequestToDonwload", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MasterServer is the server API for Master service.
// All implementations must embed UnimplementedMasterServer
// for forward compatibility
type MasterServer interface {
	KeepMeAlive(Master_KeepMeAliveServer) error
	ConfirmUpload(context.Context, *FileUploadStatus) (*emptypb.Empty, error)
	RegisterNode(context.Context, *RegisterRequest) (*RegisterRequest, error)
	RequestToUpload(context.Context, *UploadRequest) (*HostAddress, error)
	RequestToDonwload(context.Context, *DownloadRequest) (*DownloadResponse, error)
	mustEmbedUnimplementedMasterServer()
}

// UnimplementedMasterServer must be embedded to have forward compatible implementations.
type UnimplementedMasterServer struct {
}

func (UnimplementedMasterServer) KeepMeAlive(Master_KeepMeAliveServer) error {
	return status.Errorf(codes.Unimplemented, "method KeepMeAlive not implemented")
}
func (UnimplementedMasterServer) ConfirmUpload(context.Context, *FileUploadStatus) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ConfirmUpload not implemented")
}
func (UnimplementedMasterServer) RegisterNode(context.Context, *RegisterRequest) (*RegisterRequest, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterNode not implemented")
}
func (UnimplementedMasterServer) RequestToUpload(context.Context, *UploadRequest) (*HostAddress, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestToUpload not implemented")
}
func (UnimplementedMasterServer) RequestToDonwload(context.Context, *DownloadRequest) (*DownloadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RequestToDonwload not implemented")
}
func (UnimplementedMasterServer) mustEmbedUnimplementedMasterServer() {}

// UnsafeMasterServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MasterServer will
// result in compilation errors.
type UnsafeMasterServer interface {
	mustEmbedUnimplementedMasterServer()
}

func RegisterMasterServer(s grpc.ServiceRegistrar, srv MasterServer) {
	s.RegisterService(&Master_ServiceDesc, srv)
}

func _Master_KeepMeAlive_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MasterServer).KeepMeAlive(&masterKeepMeAliveServer{stream})
}

type Master_KeepMeAliveServer interface {
	SendAndClose(*emptypb.Empty) error
	Recv() (*HeartBeat, error)
	grpc.ServerStream
}

type masterKeepMeAliveServer struct {
	grpc.ServerStream
}

func (x *masterKeepMeAliveServer) SendAndClose(m *emptypb.Empty) error {
	return x.ServerStream.SendMsg(m)
}

func (x *masterKeepMeAliveServer) Recv() (*HeartBeat, error) {
	m := new(HeartBeat)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Master_ConfirmUpload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(FileUploadStatus)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServer).ConfirmUpload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/master.Master/ConfirmUpload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServer).ConfirmUpload(ctx, req.(*FileUploadStatus))
	}
	return interceptor(ctx, in, info, handler)
}

func _Master_RegisterNode_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServer).RegisterNode(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/master.Master/RegisterNode",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServer).RegisterNode(ctx, req.(*RegisterRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Master_RequestToUpload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServer).RequestToUpload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/master.Master/RequestToUpload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServer).RequestToUpload(ctx, req.(*UploadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Master_RequestToDonwload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(DownloadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MasterServer).RequestToDonwload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/master.Master/RequestToDonwload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MasterServer).RequestToDonwload(ctx, req.(*DownloadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Master_ServiceDesc is the grpc.ServiceDesc for Master service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Master_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "master.Master",
	HandlerType: (*MasterServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ConfirmUpload",
			Handler:    _Master_ConfirmUpload_Handler,
		},
		{
			MethodName: "RegisterNode",
			Handler:    _Master_RegisterNode_Handler,
		},
		{
			MethodName: "RequestToUpload",
			Handler:    _Master_RequestToUpload_Handler,
		},
		{
			MethodName: "RequestToDonwload",
			Handler:    _Master_RequestToDonwload_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "KeepMeAlive",
			Handler:       _Master_KeepMeAlive_Handler,
			ClientStreams: true,
		},
	},
	Metadata: "master.proto",
}
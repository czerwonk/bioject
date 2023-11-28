// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.24.4
// source: api.proto

package bioject

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	BioJectService_AddRoute_FullMethodName      = "/BioJectService/AddRoute"
	BioJectService_WithdrawRoute_FullMethodName = "/BioJectService/WithdrawRoute"
)

// BioJectServiceClient is the client API for BioJectService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type BioJectServiceClient interface {
	AddRoute(ctx context.Context, in *AddRouteRequest, opts ...grpc.CallOption) (*Result, error)
	WithdrawRoute(ctx context.Context, in *WithdrawRouteRequest, opts ...grpc.CallOption) (*Result, error)
}

type bioJectServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewBioJectServiceClient(cc grpc.ClientConnInterface) BioJectServiceClient {
	return &bioJectServiceClient{cc}
}

func (c *bioJectServiceClient) AddRoute(ctx context.Context, in *AddRouteRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, BioJectService_AddRoute_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *bioJectServiceClient) WithdrawRoute(ctx context.Context, in *WithdrawRouteRequest, opts ...grpc.CallOption) (*Result, error) {
	out := new(Result)
	err := c.cc.Invoke(ctx, BioJectService_WithdrawRoute_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// BioJectServiceServer is the server API for BioJectService service.
// All implementations should embed UnimplementedBioJectServiceServer
// for forward compatibility
type BioJectServiceServer interface {
	AddRoute(context.Context, *AddRouteRequest) (*Result, error)
	WithdrawRoute(context.Context, *WithdrawRouteRequest) (*Result, error)
}

// UnimplementedBioJectServiceServer should be embedded to have forward compatible implementations.
type UnimplementedBioJectServiceServer struct {
}

func (UnimplementedBioJectServiceServer) AddRoute(context.Context, *AddRouteRequest) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddRoute not implemented")
}
func (UnimplementedBioJectServiceServer) WithdrawRoute(context.Context, *WithdrawRouteRequest) (*Result, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WithdrawRoute not implemented")
}

// UnsafeBioJectServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to BioJectServiceServer will
// result in compilation errors.
type UnsafeBioJectServiceServer interface {
	mustEmbedUnimplementedBioJectServiceServer()
}

func RegisterBioJectServiceServer(s grpc.ServiceRegistrar, srv BioJectServiceServer) {
	s.RegisterService(&BioJectService_ServiceDesc, srv)
}

func _BioJectService_AddRoute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddRouteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BioJectServiceServer).AddRoute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BioJectService_AddRoute_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BioJectServiceServer).AddRoute(ctx, req.(*AddRouteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _BioJectService_WithdrawRoute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WithdrawRouteRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(BioJectServiceServer).WithdrawRoute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: BioJectService_WithdrawRoute_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(BioJectServiceServer).WithdrawRoute(ctx, req.(*WithdrawRouteRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// BioJectService_ServiceDesc is the grpc.ServiceDesc for BioJectService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var BioJectService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "BioJectService",
	HandlerType: (*BioJectServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "AddRoute",
			Handler:    _BioJectService_AddRoute_Handler,
		},
		{
			MethodName: "WithdrawRoute",
			Handler:    _BioJectService_WithdrawRoute_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api.proto",
}

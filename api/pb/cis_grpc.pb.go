// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: cis.proto

package pb

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

const (
	CIS_GetRoots_FullMethodName       = "/pb.CIS/GetRoots"
	CIS_GetCertificate_FullMethodName = "/pb.CIS/GetCertificate"
)

// CISClient is the client API for CIS service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CISClient interface {
	// Roots returns the root CAs
	GetRoots(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*RootsResponse, error)
	// GetCertificate returns the certificate
	GetCertificate(ctx context.Context, in *GetCertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error)
}

type cISClient struct {
	cc grpc.ClientConnInterface
}

func NewCISClient(cc grpc.ClientConnInterface) CISClient {
	return &cISClient{cc}
}

func (c *cISClient) GetRoots(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*RootsResponse, error) {
	out := new(RootsResponse)
	err := c.cc.Invoke(ctx, CIS_GetRoots_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cISClient) GetCertificate(ctx context.Context, in *GetCertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error) {
	out := new(CertificateResponse)
	err := c.cc.Invoke(ctx, CIS_GetCertificate_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CISServer is the server API for CIS service.
// All implementations should embed UnimplementedCISServer
// for forward compatibility
type CISServer interface {
	// Roots returns the root CAs
	GetRoots(context.Context, *emptypb.Empty) (*RootsResponse, error)
	// GetCertificate returns the certificate
	GetCertificate(context.Context, *GetCertificateRequest) (*CertificateResponse, error)
}

// UnimplementedCISServer should be embedded to have forward compatible implementations.
type UnimplementedCISServer struct {
}

func (UnimplementedCISServer) GetRoots(context.Context, *emptypb.Empty) (*RootsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetRoots not implemented")
}
func (UnimplementedCISServer) GetCertificate(context.Context, *GetCertificateRequest) (*CertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCertificate not implemented")
}

// UnsafeCISServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CISServer will
// result in compilation errors.
type UnsafeCISServer interface {
	mustEmbedUnimplementedCISServer()
}

func RegisterCISServer(s grpc.ServiceRegistrar, srv CISServer) {
	s.RegisterService(&CIS_ServiceDesc, srv)
}

func _CIS_GetRoots_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(emptypb.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CISServer).GetRoots(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CIS_GetRoots_FullMethodName,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(CISServer).GetRoots(ctx, req.(*emptypb.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _CIS_GetCertificate_Handler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(GetCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CISServer).GetCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: CIS_GetCertificate_FullMethodName,
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(CISServer).GetCertificate(ctx, req.(*GetCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CIS_ServiceDesc is the grpc.ServiceDesc for CIS service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CIS_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.CIS",
	HandlerType: (*CISServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetRoots",
			Handler:    _CIS_GetRoots_Handler,
		},
		{
			MethodName: "GetCertificate",
			Handler:    _CIS_GetCertificate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cis.proto",
}
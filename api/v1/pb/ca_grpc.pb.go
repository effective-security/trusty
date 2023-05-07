// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.22.3
// source: ca.proto

package pb

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

// CAClient is the client API for CA service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type CAClient interface {
	// ProfileInfo returns the certificate profile info
	ProfileInfo(ctx context.Context, in *CertProfileInfoRequest, opts ...grpc.CallOption) (*CertProfile, error)
	// GetIssuer returns the issuing CA
	GetIssuer(ctx context.Context, in *IssuerInfoRequest, opts ...grpc.CallOption) (*IssuerInfo, error)
	// ListIssuers returns the issuing CAs
	ListIssuers(ctx context.Context, in *ListIssuersRequest, opts ...grpc.CallOption) (*IssuersInfoResponse, error)
	// SignCertificate returns the certificate
	SignCertificate(ctx context.Context, in *SignCertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error)
	// GetCertificate returns the certificate
	GetCertificate(ctx context.Context, in *GetCertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error)
	// GetCRL returns the CRL
	GetCRL(ctx context.Context, in *GetCrlRequest, opts ...grpc.CallOption) (*CrlResponse, error)
	// SignOCSP returns OCSP response
	SignOCSP(ctx context.Context, in *OCSPRequest, opts ...grpc.CallOption) (*OCSPResponse, error)
	// RevokeCertificate returns the revoked certificate
	RevokeCertificate(ctx context.Context, in *RevokeCertificateRequest, opts ...grpc.CallOption) (*RevokedCertificateResponse, error)
	// PublishCrls returns published CRLs
	PublishCrls(ctx context.Context, in *PublishCrlsRequest, opts ...grpc.CallOption) (*CrlsResponse, error)
	// ListOrgCertificates returns the Org certificates
	ListOrgCertificates(ctx context.Context, in *ListOrgCertificatesRequest, opts ...grpc.CallOption) (*CertificatesResponse, error)
	// ListCertificates returns stream of Certificates
	ListCertificates(ctx context.Context, in *ListByIssuerRequest, opts ...grpc.CallOption) (*CertificatesResponse, error)
	// ListRevokedCertificates returns stream of Revoked Certificates
	ListRevokedCertificates(ctx context.Context, in *ListByIssuerRequest, opts ...grpc.CallOption) (*RevokedCertificatesResponse, error)
	// UpdateCertificateLabel returns the updated certificate
	UpdateCertificateLabel(ctx context.Context, in *UpdateCertificateLabelRequest, opts ...grpc.CallOption) (*CertificateResponse, error)
	// ListDelegatedIssuers returns the delegated issuing CAs
	ListDelegatedIssuers(ctx context.Context, in *ListIssuersRequest, opts ...grpc.CallOption) (*IssuersInfoResponse, error)
	// RegisterDelegatedIssuer creates new delegate issuer.
	// NOTE: the key and CSR is generated by the server, and request field must be empty
	RegisterDelegatedIssuer(ctx context.Context, in *SignCertificateRequest, opts ...grpc.CallOption) (*IssuerInfo, error)
	// ArchiveDelegatedIssuer archives a delegated issuer.
	ArchiveDelegatedIssuer(ctx context.Context, in *IssuerInfoRequest, opts ...grpc.CallOption) (*IssuerInfo, error)
	// RegisterProfile registers the certificate profile
	RegisterProfile(ctx context.Context, in *RegisterProfileRequest, opts ...grpc.CallOption) (*CertProfile, error)
}

type cAClient struct {
	cc grpc.ClientConnInterface
}

func NewCAClient(cc grpc.ClientConnInterface) CAClient {
	return &cAClient{cc}
}

func (c *cAClient) ProfileInfo(ctx context.Context, in *CertProfileInfoRequest, opts ...grpc.CallOption) (*CertProfile, error) {
	out := new(CertProfile)
	err := c.cc.Invoke(ctx, "/pb.CA/ProfileInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) GetIssuer(ctx context.Context, in *IssuerInfoRequest, opts ...grpc.CallOption) (*IssuerInfo, error) {
	out := new(IssuerInfo)
	err := c.cc.Invoke(ctx, "/pb.CA/GetIssuer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) ListIssuers(ctx context.Context, in *ListIssuersRequest, opts ...grpc.CallOption) (*IssuersInfoResponse, error) {
	out := new(IssuersInfoResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/ListIssuers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) SignCertificate(ctx context.Context, in *SignCertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error) {
	out := new(CertificateResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/SignCertificate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) GetCertificate(ctx context.Context, in *GetCertificateRequest, opts ...grpc.CallOption) (*CertificateResponse, error) {
	out := new(CertificateResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/GetCertificate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) GetCRL(ctx context.Context, in *GetCrlRequest, opts ...grpc.CallOption) (*CrlResponse, error) {
	out := new(CrlResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/GetCRL", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) SignOCSP(ctx context.Context, in *OCSPRequest, opts ...grpc.CallOption) (*OCSPResponse, error) {
	out := new(OCSPResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/SignOCSP", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) RevokeCertificate(ctx context.Context, in *RevokeCertificateRequest, opts ...grpc.CallOption) (*RevokedCertificateResponse, error) {
	out := new(RevokedCertificateResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/RevokeCertificate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) PublishCrls(ctx context.Context, in *PublishCrlsRequest, opts ...grpc.CallOption) (*CrlsResponse, error) {
	out := new(CrlsResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/PublishCrls", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) ListOrgCertificates(ctx context.Context, in *ListOrgCertificatesRequest, opts ...grpc.CallOption) (*CertificatesResponse, error) {
	out := new(CertificatesResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/ListOrgCertificates", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) ListCertificates(ctx context.Context, in *ListByIssuerRequest, opts ...grpc.CallOption) (*CertificatesResponse, error) {
	out := new(CertificatesResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/ListCertificates", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) ListRevokedCertificates(ctx context.Context, in *ListByIssuerRequest, opts ...grpc.CallOption) (*RevokedCertificatesResponse, error) {
	out := new(RevokedCertificatesResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/ListRevokedCertificates", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) UpdateCertificateLabel(ctx context.Context, in *UpdateCertificateLabelRequest, opts ...grpc.CallOption) (*CertificateResponse, error) {
	out := new(CertificateResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/UpdateCertificateLabel", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) ListDelegatedIssuers(ctx context.Context, in *ListIssuersRequest, opts ...grpc.CallOption) (*IssuersInfoResponse, error) {
	out := new(IssuersInfoResponse)
	err := c.cc.Invoke(ctx, "/pb.CA/ListDelegatedIssuers", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) RegisterDelegatedIssuer(ctx context.Context, in *SignCertificateRequest, opts ...grpc.CallOption) (*IssuerInfo, error) {
	out := new(IssuerInfo)
	err := c.cc.Invoke(ctx, "/pb.CA/RegisterDelegatedIssuer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) ArchiveDelegatedIssuer(ctx context.Context, in *IssuerInfoRequest, opts ...grpc.CallOption) (*IssuerInfo, error) {
	out := new(IssuerInfo)
	err := c.cc.Invoke(ctx, "/pb.CA/ArchiveDelegatedIssuer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *cAClient) RegisterProfile(ctx context.Context, in *RegisterProfileRequest, opts ...grpc.CallOption) (*CertProfile, error) {
	out := new(CertProfile)
	err := c.cc.Invoke(ctx, "/pb.CA/RegisterProfile", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// CAServer is the server API for CA service.
// All implementations should embed UnimplementedCAServer
// for forward compatibility
type CAServer interface {
	// ProfileInfo returns the certificate profile info
	ProfileInfo(context.Context, *CertProfileInfoRequest) (*CertProfile, error)
	// GetIssuer returns the issuing CA
	GetIssuer(context.Context, *IssuerInfoRequest) (*IssuerInfo, error)
	// ListIssuers returns the issuing CAs
	ListIssuers(context.Context, *ListIssuersRequest) (*IssuersInfoResponse, error)
	// SignCertificate returns the certificate
	SignCertificate(context.Context, *SignCertificateRequest) (*CertificateResponse, error)
	// GetCertificate returns the certificate
	GetCertificate(context.Context, *GetCertificateRequest) (*CertificateResponse, error)
	// GetCRL returns the CRL
	GetCRL(context.Context, *GetCrlRequest) (*CrlResponse, error)
	// SignOCSP returns OCSP response
	SignOCSP(context.Context, *OCSPRequest) (*OCSPResponse, error)
	// RevokeCertificate returns the revoked certificate
	RevokeCertificate(context.Context, *RevokeCertificateRequest) (*RevokedCertificateResponse, error)
	// PublishCrls returns published CRLs
	PublishCrls(context.Context, *PublishCrlsRequest) (*CrlsResponse, error)
	// ListOrgCertificates returns the Org certificates
	ListOrgCertificates(context.Context, *ListOrgCertificatesRequest) (*CertificatesResponse, error)
	// ListCertificates returns stream of Certificates
	ListCertificates(context.Context, *ListByIssuerRequest) (*CertificatesResponse, error)
	// ListRevokedCertificates returns stream of Revoked Certificates
	ListRevokedCertificates(context.Context, *ListByIssuerRequest) (*RevokedCertificatesResponse, error)
	// UpdateCertificateLabel returns the updated certificate
	UpdateCertificateLabel(context.Context, *UpdateCertificateLabelRequest) (*CertificateResponse, error)
	// ListDelegatedIssuers returns the delegated issuing CAs
	ListDelegatedIssuers(context.Context, *ListIssuersRequest) (*IssuersInfoResponse, error)
	// RegisterDelegatedIssuer creates new delegate issuer.
	// NOTE: the key and CSR is generated by the server, and request field must be empty
	RegisterDelegatedIssuer(context.Context, *SignCertificateRequest) (*IssuerInfo, error)
	// ArchiveDelegatedIssuer archives a delegated issuer.
	ArchiveDelegatedIssuer(context.Context, *IssuerInfoRequest) (*IssuerInfo, error)
	// RegisterProfile registers the certificate profile
	RegisterProfile(context.Context, *RegisterProfileRequest) (*CertProfile, error)
}

// UnimplementedCAServer should be embedded to have forward compatible implementations.
type UnimplementedCAServer struct {
}

func (UnimplementedCAServer) ProfileInfo(context.Context, *CertProfileInfoRequest) (*CertProfile, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ProfileInfo not implemented")
}
func (UnimplementedCAServer) GetIssuer(context.Context, *IssuerInfoRequest) (*IssuerInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetIssuer not implemented")
}
func (UnimplementedCAServer) ListIssuers(context.Context, *ListIssuersRequest) (*IssuersInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListIssuers not implemented")
}
func (UnimplementedCAServer) SignCertificate(context.Context, *SignCertificateRequest) (*CertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignCertificate not implemented")
}
func (UnimplementedCAServer) GetCertificate(context.Context, *GetCertificateRequest) (*CertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCertificate not implemented")
}
func (UnimplementedCAServer) GetCRL(context.Context, *GetCrlRequest) (*CrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetCRL not implemented")
}
func (UnimplementedCAServer) SignOCSP(context.Context, *OCSPRequest) (*OCSPResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SignOCSP not implemented")
}
func (UnimplementedCAServer) RevokeCertificate(context.Context, *RevokeCertificateRequest) (*RevokedCertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RevokeCertificate not implemented")
}
func (UnimplementedCAServer) PublishCrls(context.Context, *PublishCrlsRequest) (*CrlsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PublishCrls not implemented")
}
func (UnimplementedCAServer) ListOrgCertificates(context.Context, *ListOrgCertificatesRequest) (*CertificatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListOrgCertificates not implemented")
}
func (UnimplementedCAServer) ListCertificates(context.Context, *ListByIssuerRequest) (*CertificatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListCertificates not implemented")
}
func (UnimplementedCAServer) ListRevokedCertificates(context.Context, *ListByIssuerRequest) (*RevokedCertificatesResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListRevokedCertificates not implemented")
}
func (UnimplementedCAServer) UpdateCertificateLabel(context.Context, *UpdateCertificateLabelRequest) (*CertificateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateCertificateLabel not implemented")
}
func (UnimplementedCAServer) ListDelegatedIssuers(context.Context, *ListIssuersRequest) (*IssuersInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListDelegatedIssuers not implemented")
}
func (UnimplementedCAServer) RegisterDelegatedIssuer(context.Context, *SignCertificateRequest) (*IssuerInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterDelegatedIssuer not implemented")
}
func (UnimplementedCAServer) ArchiveDelegatedIssuer(context.Context, *IssuerInfoRequest) (*IssuerInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ArchiveDelegatedIssuer not implemented")
}
func (UnimplementedCAServer) RegisterProfile(context.Context, *RegisterProfileRequest) (*CertProfile, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RegisterProfile not implemented")
}

// UnsafeCAServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to CAServer will
// result in compilation errors.
type UnsafeCAServer interface {
	mustEmbedUnimplementedCAServer()
}

func RegisterCAServer(s grpc.ServiceRegistrar, srv CAServer) {
	s.RegisterService(&CA_ServiceDesc, srv)
}

func _CA_ProfileInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CertProfileInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).ProfileInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/ProfileInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).ProfileInfo(ctx, req.(*CertProfileInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_GetIssuer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssuerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).GetIssuer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/GetIssuer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).GetIssuer(ctx, req.(*IssuerInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_ListIssuers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListIssuersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).ListIssuers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/ListIssuers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).ListIssuers(ctx, req.(*ListIssuersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_SignCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).SignCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/SignCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).SignCertificate(ctx, req.(*SignCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_GetCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).GetCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/GetCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).GetCertificate(ctx, req.(*GetCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_GetCRL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetCrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).GetCRL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/GetCRL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).GetCRL(ctx, req.(*GetCrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_SignOCSP_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(OCSPRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).SignOCSP(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/SignOCSP",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).SignOCSP(ctx, req.(*OCSPRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_RevokeCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RevokeCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).RevokeCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/RevokeCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).RevokeCertificate(ctx, req.(*RevokeCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_PublishCrls_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PublishCrlsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).PublishCrls(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/PublishCrls",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).PublishCrls(ctx, req.(*PublishCrlsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_ListOrgCertificates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListOrgCertificatesRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).ListOrgCertificates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/ListOrgCertificates",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).ListOrgCertificates(ctx, req.(*ListOrgCertificatesRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_ListCertificates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListByIssuerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).ListCertificates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/ListCertificates",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).ListCertificates(ctx, req.(*ListByIssuerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_ListRevokedCertificates_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListByIssuerRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).ListRevokedCertificates(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/ListRevokedCertificates",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).ListRevokedCertificates(ctx, req.(*ListByIssuerRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_UpdateCertificateLabel_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UpdateCertificateLabelRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).UpdateCertificateLabel(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/UpdateCertificateLabel",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).UpdateCertificateLabel(ctx, req.(*UpdateCertificateLabelRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_ListDelegatedIssuers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListIssuersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).ListDelegatedIssuers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/ListDelegatedIssuers",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).ListDelegatedIssuers(ctx, req.(*ListIssuersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_RegisterDelegatedIssuer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignCertificateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).RegisterDelegatedIssuer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/RegisterDelegatedIssuer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).RegisterDelegatedIssuer(ctx, req.(*SignCertificateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_ArchiveDelegatedIssuer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(IssuerInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).ArchiveDelegatedIssuer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/ArchiveDelegatedIssuer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).ArchiveDelegatedIssuer(ctx, req.(*IssuerInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _CA_RegisterProfile_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterProfileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CAServer).RegisterProfile(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/pb.CA/RegisterProfile",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CAServer).RegisterProfile(ctx, req.(*RegisterProfileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// CA_ServiceDesc is the grpc.ServiceDesc for CA service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var CA_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "pb.CA",
	HandlerType: (*CAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ProfileInfo",
			Handler:    _CA_ProfileInfo_Handler,
		},
		{
			MethodName: "GetIssuer",
			Handler:    _CA_GetIssuer_Handler,
		},
		{
			MethodName: "ListIssuers",
			Handler:    _CA_ListIssuers_Handler,
		},
		{
			MethodName: "SignCertificate",
			Handler:    _CA_SignCertificate_Handler,
		},
		{
			MethodName: "GetCertificate",
			Handler:    _CA_GetCertificate_Handler,
		},
		{
			MethodName: "GetCRL",
			Handler:    _CA_GetCRL_Handler,
		},
		{
			MethodName: "SignOCSP",
			Handler:    _CA_SignOCSP_Handler,
		},
		{
			MethodName: "RevokeCertificate",
			Handler:    _CA_RevokeCertificate_Handler,
		},
		{
			MethodName: "PublishCrls",
			Handler:    _CA_PublishCrls_Handler,
		},
		{
			MethodName: "ListOrgCertificates",
			Handler:    _CA_ListOrgCertificates_Handler,
		},
		{
			MethodName: "ListCertificates",
			Handler:    _CA_ListCertificates_Handler,
		},
		{
			MethodName: "ListRevokedCertificates",
			Handler:    _CA_ListRevokedCertificates_Handler,
		},
		{
			MethodName: "UpdateCertificateLabel",
			Handler:    _CA_UpdateCertificateLabel_Handler,
		},
		{
			MethodName: "ListDelegatedIssuers",
			Handler:    _CA_ListDelegatedIssuers_Handler,
		},
		{
			MethodName: "RegisterDelegatedIssuer",
			Handler:    _CA_RegisterDelegatedIssuer_Handler,
		},
		{
			MethodName: "ArchiveDelegatedIssuer",
			Handler:    _CA_ArchiveDelegatedIssuer_Handler,
		},
		{
			MethodName: "RegisterProfile",
			Handler:    _CA_RegisterProfile_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "ca.proto",
}

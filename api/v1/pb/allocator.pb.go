// Code generated by protoc-gen-go-mock. DO NOT EDIT.

package pb

import (
	"net/http"

	"github.com/effective-security/porto/xhttp/httperror"
	"github.com/effective-security/porto/xhttp/marshal"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RequestAllocator defines constructor to allocate Protobuf request
type RequestAllocator func() any

// MethodInfo provides info about RPC method
type MethodInfo struct {
	Allocator    RequestAllocator
	AllowedRoles []string
}

// UnmarshalRequest unmarshals JSON body of HTTP request to protobuf request
func UnmarshalRequest(w http.ResponseWriter, r *http.Request) (any, *MethodInfo, error) {
	info := methods[r.URL.Path]
	if info == nil {
		err := httperror.NotFound("path not found: %s", r.URL.Path)
		marshal.WriteJSON(w, r, err)
		return nil, nil, err
	}

	req := info.Allocator()
	err := marshal.DecodeBody(w, r, req)
	if err != nil {
		// DecodeBody writes error response and logs, if invalid request
		return nil, nil, err
	}
	return req, info, nil
}

// GetMethodInfo returns MethodInfo
func GetMethodInfo(method string) *MethodInfo {
	return methods[method]
}

// methods defines map for routes
var methods = map[string]*MethodInfo{

	"/pb.CA/ProfileInfo": {
		Allocator: func() any { return new(CertProfileInfoRequest) },
	},

	"/pb.CA/GetIssuer": {
		Allocator: func() any { return new(IssuerInfoRequest) },
	},

	"/pb.CA/ListIssuers": {
		Allocator: func() any { return new(ListIssuersRequest) },
	},

	"/pb.CA/SignCertificate": {
		Allocator: func() any { return new(SignCertificateRequest) },
	},

	"/pb.CA/GetCertificate": {
		Allocator: func() any { return new(GetCertificateRequest) },
	},

	"/pb.CA/GetCRL": {
		Allocator: func() any { return new(GetCrlRequest) },
	},

	"/pb.CA/SignOCSP": {
		Allocator: func() any { return new(OCSPRequest) },
	},

	"/pb.CA/RevokeCertificate": {
		Allocator: func() any { return new(RevokeCertificateRequest) },
	},

	"/pb.CA/PublishCrls": {
		Allocator: func() any { return new(PublishCrlsRequest) },
	},

	"/pb.CA/ListOrgCertificates": {
		Allocator: func() any { return new(ListOrgCertificatesRequest) },
	},

	"/pb.CA/ListCertificates": {
		Allocator: func() any { return new(ListByIssuerRequest) },
	},

	"/pb.CA/ListRevokedCertificates": {
		Allocator: func() any { return new(ListByIssuerRequest) },
	},

	"/pb.CA/UpdateCertificateLabel": {
		Allocator: func() any { return new(UpdateCertificateLabelRequest) },
	},

	"/pb.CA/ListDelegatedIssuers": {
		Allocator: func() any { return new(ListIssuersRequest) },
	},

	"/pb.CA/RegisterDelegatedIssuer": {
		Allocator: func() any { return new(SignCertificateRequest) },
	},

	"/pb.CA/ArchiveDelegatedIssuer": {
		Allocator: func() any { return new(IssuerInfoRequest) },
	},

	"/pb.CA/RegisterProfile": {
		Allocator: func() any { return new(RegisterProfileRequest) },
	},

	"/pb.CIS/GetRoots": {
		Allocator: func() any { return new(emptypb.Empty) },
	},

	"/pb.CIS/GetCertificate": {
		Allocator: func() any { return new(GetCertificateRequest) },
	},

	"/pb.Status/Version": {
		Allocator: func() any { return new(emptypb.Empty) },
	},

	"/pb.Status/Server": {
		Allocator: func() any { return new(emptypb.Empty) },
	},

	"/pb.Status/Caller": {
		Allocator: func() any { return new(emptypb.Empty) },
	},
}
// Package tools for go mod.
// Add here tool dependencies that are not part of the build.

// +build tools

package tools

import (
	_ "github.com/cloudflare/cfssl/cmd/cfssl"
	_ "github.com/cloudflare/cfssl/cmd/cfssl-certinfo"
	_ "github.com/cloudflare/cfssl/cmd/cfssljson"
	_ "github.com/go-phorce/cov-report/cmd/cov-report"
	_ "github.com/mattn/goreman"
	_ "github.com/mattn/goveralls"
	_ "golang.org/x/lint/golint"
	_ "golang.org/x/tools/cmd/godoc"
	_ "golang.org/x/tools/cmd/goimports"
	_ "golang.org/x/tools/cmd/gorename"
	_ "golang.org/x/tools/cmd/guru"
	_ "golang.org/x/tools/cmd/stringer"
	//_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway"
	//_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2"
)

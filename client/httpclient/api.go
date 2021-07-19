package httpclient

import (
	"github.com/go-phorce/dolly/xhttp/header"
)

var jsonContentHeaders = map[string]string{
	header.Accept:      header.ApplicationJSON,
	header.ContentType: header.ApplicationJSON,
}

// API defines API client interface
type API interface {
	MartiniClient
}

// ensure compiles
var _ = interface{}(&Client{}).(API)

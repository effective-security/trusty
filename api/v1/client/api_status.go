package client

import (
	"context"

	"github.com/effective-security/porto/pkg/retriable"
	v1 "github.com/effective-security/trusty/api/v1"
	"github.com/effective-security/trusty/api/v1/pb"
)

// HTTPStatusClient provides Status over legacy HTTP
type HTTPStatusClient struct {
	client retriable.HTTPClient
}

// NewHTTPStatusClient returns legacy HTTP client
func NewHTTPStatusClient(client retriable.HTTPClient) *HTTPStatusClient {
	return &HTTPStatusClient{client: client}
}

// Version returns ServerVersion
func (c *HTTPStatusClient) Version(ctx context.Context) (*pb.ServerVersion, error) {
	r := new(pb.ServerVersion)
	_, _, err := c.client.Get(ctx, v1.PathForStatusVersion, r)
	if err != nil {
		return nil, err
	}
	return r, err
}

// Status returns ServerStatusResponse
func (c *HTTPStatusClient) Status(ctx context.Context) (*pb.ServerStatusResponse, error) {
	r := new(pb.ServerStatusResponse)
	_, _, err := c.client.Get(ctx, v1.PathForStatusServer, r)
	if err != nil {
		return nil, err
	}
	return r, err
}

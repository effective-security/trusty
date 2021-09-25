package httpclient

import (
	"context"

	"github.com/juju/errors"
	v1 "github.com/martinisecurity/trusty/api/v1"
)

// RefreshToken returns AuthTokenRefreshResponse
func (c *Client) RefreshToken(ctx context.Context) (*v1.AuthTokenRefreshResponse, error) {
	r := new(v1.AuthTokenRefreshResponse)
	_, _, err := c.Get(ctx, v1.PathForAuthTokenRefresh, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, err
}

package httpclient

import (
	"context"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/juju/errors"
)

// RefreshToken returns AuthTokenRefreshResponse
func (c *Client) RefreshToken(ctx context.Context) (*v1.AuthTokenRefreshResponse, error) {
	r := new(v1.AuthTokenRefreshResponse)
	_, err := c.Get(ctx, v1.PathForAuthTokenRefresh, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, err
}

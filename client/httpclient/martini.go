package httpclient

import (
	"context"
	"net/url"

	v1 "github.com/ekspand/trusty/api/v1"
	"github.com/juju/errors"
)

// MartiniClient client interface
type MartiniClient interface {
	// SearchCorps returns SearchOpenCorporatesResponse
	SearchCorps(ctx context.Context, name, jurisdiction string) (*v1.SearchOpenCorporatesResponse, error)
}

// SearchCorps returns SearchOpenCorporatesResponse
func (c *Client) SearchCorps(ctx context.Context, name, jurisdiction string) (*v1.SearchOpenCorporatesResponse, error) {
	u := v1.PathForMartiniSearchCorps + "?name=" + url.QueryEscape(name)
	if jurisdiction != "" {
		u += "&jurisdiction=" + url.QueryEscape(jurisdiction)
	}
	r := new(v1.SearchOpenCorporatesResponse)
	_, err := c.Get(ctx, u, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, err
}

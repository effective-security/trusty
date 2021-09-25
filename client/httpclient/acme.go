package httpclient

import (
	"context"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/api/v2acme"
)

// AcmeClient interface
type AcmeClient interface {
	// Directory returns directory response
	Directory(ctx context.Context) (map[string]string, error)
}

// Directory returns directory response
func (c *Client) Directory(ctx context.Context) (map[string]string, error) {
	var r map[string]string
	_, _, err := c.Get(ctx, v2acme.PathForDirectoryBase, &r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	nonceURL := r["newNonce"]
	if nonceURL != "" {
		p := nonceURL[len(c.CurrentHost()):]
		logger.Infof("noncePath=%q", p)
		c.WithNonce(p)
	}
	return r, err
}

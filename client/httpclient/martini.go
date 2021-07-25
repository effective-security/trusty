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

// Orgs returns user's Orgs
func (c *Client) Orgs(ctx context.Context) (*v1.OrgsResponse, error) {
	r := new(v1.OrgsResponse)
	_, err := c.Get(ctx, v1.PathForMartiniOrgs, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, nil
}

// FccFRN returns Fcc FRN
func (c *Client) FccFRN(ctx context.Context, filerID string) (*v1.FccFrnResponse, error) {
	r := new(v1.FccFrnResponse)
	_, err := c.Get(ctx, v1.PathForMartiniFccFrn+"?filer_id="+filerID, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, nil
}

// FccContact returns Fcc FRN Contact
func (c *Client) FccContact(ctx context.Context, frn string) (*v1.FccContactResponse, error) {
	r := new(v1.FccContactResponse)
	_, err := c.Get(ctx, v1.PathForMartiniFccContact+"?frn="+frn, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, nil
}

// RegisterOrg starts Org registration flow
func (c *Client) RegisterOrg(ctx context.Context, filerID string) (*v1.RegisterOrgResponse, error) {
	req := &v1.RegisterOrgRequest{
		FilerID: filerID,
	}

	res := new(v1.RegisterOrgResponse)
	_, err := c.PostRequest(ctx, v1.PathForMartiniRegisterOrg, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// ApproveOrg validates Org registration
func (c *Client) ApproveOrg(ctx context.Context, token, code string) (*v1.OrgResponse, error) {
	req := &v1.ApproveOrgRequest{
		Token: token,
		Code:  code,
	}

	res := new(v1.OrgResponse)
	_, err := c.PostRequest(ctx, v1.PathForMartiniApproveOrg, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

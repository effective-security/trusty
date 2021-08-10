package httpclient

import (
	"context"
	"net/url"
	"strings"

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
	_, _, err := c.Get(ctx, u, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, err
}

// Orgs returns user's Orgs
func (c *Client) Orgs(ctx context.Context) (*v1.OrgsResponse, error) {
	r := new(v1.OrgsResponse)
	_, _, err := c.Get(ctx, v1.PathForMartiniOrgs, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, nil
}

// Certificates returns user's Certificates
func (c *Client) Certificates(ctx context.Context) (*v1.CertificatesResponse, error) {
	r := new(v1.CertificatesResponse)
	_, _, err := c.Get(ctx, v1.PathForMartiniCerts, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, nil
}

// FccFRN returns Fcc FRN
func (c *Client) FccFRN(ctx context.Context, filerID string) (*v1.FccFrnResponse, error) {
	r := new(v1.FccFrnResponse)
	_, _, err := c.Get(ctx, v1.PathForMartiniFccFrn+"?filer_id="+filerID, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, nil
}

// FccContact returns Fcc FRN Contact
func (c *Client) FccContact(ctx context.Context, frn string) (*v1.FccContactResponse, error) {
	r := new(v1.FccContactResponse)
	_, _, err := c.Get(ctx, v1.PathForMartiniFccContact+"?frn="+frn, r)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return r, nil
}

// RegisterOrg starts Org registration flow
func (c *Client) RegisterOrg(ctx context.Context, filerID string) (*v1.OrgResponse, error) {
	req := &v1.RegisterOrgRequest{
		FilerID: filerID,
	}

	res := new(v1.OrgResponse)
	_, _, err := c.PostRequest(ctx, v1.PathForMartiniRegisterOrg, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// ApproveOrg approves Org registration
func (c *Client) ApproveOrg(ctx context.Context, token, code string) (*v1.OrgResponse, error) {
	req := &v1.ApproveOrgRequest{
		Token:  token,
		Code:   code,
		Action: "approve",
	}

	res := new(v1.OrgResponse)
	_, _, err := c.PostRequest(ctx, v1.PathForMartiniApproveOrg, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// ValidateOrg validates Org registration
func (c *Client) ValidateOrg(ctx context.Context, orgID string) (*v1.ValidateOrgResponse, error) {
	req := &v1.ValidateOrgRequest{
		OrgID: orgID,
	}

	res := new(v1.ValidateOrgResponse)
	_, _, err := c.PostRequest(ctx, v1.PathForMartiniValidateOrg, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// GetOrgAPIKeys returns Org API keys
func (c *Client) GetOrgAPIKeys(ctx context.Context, orgID string) (*v1.GetOrgAPIKeysResponse, error) {
	path := strings.Replace(v1.PathForMartiniOrgAPIKeys, ":org_id", orgID, 1)
	res := new(v1.GetOrgAPIKeysResponse)
	_, _, err := c.Get(ctx, path, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// CreateSubscription pays for Org validation
func (c *Client) CreateSubscription(ctx context.Context, req *v1.CreateSubscriptionRequest) (*v1.OrgResponse, error) {
	res := new(v1.OrgResponse)
	path := strings.Replace(v1.PathForMartiniOrgSubscription, ":org_id", req.OrgID, 1)
	_, _, err := c.PostRequest(ctx, path, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

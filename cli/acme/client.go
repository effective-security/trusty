package acme

import (
	"bytes"
	"context"
	"crypto"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/client/httpclient"
	"github.com/go-phorce/dolly/xhttp/header"
	"github.com/go-phorce/dolly/xhttp/retriable"
	"github.com/juju/errors"
)

// Client for ACME
type Client struct {
	Directory map[string]string
	Client    *httpclient.Client
	Jws       *JWS
	kid       string
}

// NewClient returns a new Client
func NewClient(httpClient *httpclient.Client, kid string, privateKey crypto.PrivateKey) (*Client, error) {
	dir, err := httpClient.Directory(context.Background())
	if err != nil {
		return nil, errors.Trace(err)
	}

	return &Client{
		Client:    httpClient,
		kid:       kid,
		Directory: dir,
		Jws:       NewJWS(privateKey, kid, httpClient),
	}, nil
}

// Account returns Account
// if contact is empty, then it will check for existing account
func (c *Client) Account(ctx context.Context, orgID string, hmac []byte, contact []string) (*v2acme.Account, string, error) {
	newAccountURL := c.Directory["newAccount"]
	eabJWS, err := c.signEABContent(newAccountURL, orgID, hmac)
	if err != nil {
		return nil, "", errors.Trace(err)
	}
	req := &v2acme.AccountRequest{
		Contact:                contact,
		TermsOfServiceAgreed:   true,
		OnlyReturnExisting:     len(contact) == 0,
		ExternalAccountBinding: []byte(eabJWS),
	}

	res := new(v2acme.Account)
	hdr, _, err := c.signedRequest(ctx, newAccountURL, req, res)
	if err != nil {
		return nil, "", errors.Trace(err)
	}
	keyID := hdr.Get(header.Location)
	c.Jws.SetKid(keyID)
	return res, keyID, nil
}

// Order submits the order
func (c *Client) Order(ctx context.Context, acc *v2acme.Account, req *v2acme.OrderRequest) (*v2acme.Order, string, error) {
	res := new(v2acme.Order)
	hdr, _, err := c.signedRequest(ctx, acc.OrdersURL, req, res)
	if err != nil {
		return nil, "", errors.Trace(err)
	}

	orderURL := hdr.Get(header.Location)

	return res, orderURL, nil
}

// GetOrder returns the order
func (c *Client) GetOrder(ctx context.Context, url string) (*v2acme.Order, string, error) {
	res := new(v2acme.Order)
	hdr, _, err := c.signedPostAsGet(ctx, url, res)
	if err != nil {
		return nil, "", errors.Trace(err)
	}

	orderURL := hdr.Get(header.Location)
	return res, orderURL, nil
}

// Authorization returns Authorization
func (c *Client) Authorization(ctx context.Context, authURL string) (*v2acme.Authorization, error) {
	res := new(v2acme.Authorization)
	_, _, err := c.signedPostAsGet(ctx, authURL, res)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}

// PostChallenge returns Challenge
func (c *Client) PostChallenge(ctx context.Context, url string, req interface{}) (*v2acme.Challenge, error) {
	res := new(v2acme.Challenge)
	_, _, err := c.signedRequest(ctx, url, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// Finalize returns Order
func (c *Client) Finalize(ctx context.Context, url string, req interface{}) (*v2acme.Order, error) {
	res := new(v2acme.Order)
	_, _, err := c.signedRequest(ctx, url, req, res)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// Certificate returns Certificate
func (c *Client) Certificate(ctx context.Context, url string) ([]byte, error) {
	b := bytes.NewBuffer([]byte{})
	_, _, err := c.signedPostAsGet(ctx, url, b)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return b.Bytes(), nil
}

func (c *Client) signEABContent(newAccountURL, keyID string, hmac []byte) (string, error) {
	eabJWS, err := c.Jws.SignEABContent(newAccountURL, keyID, hmac)
	if err != nil {
		return "", errors.Trace(err)
	}

	return eabJWS.FullSerialize(), nil
}

func (c *Client) signedRequest(ctx context.Context, url string, req interface{}, response interface{}) (http.Header, int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, 0, err
	}
	return c.signedPost(ctx, url, body, response)
}

func (c *Client) signedPostAsGet(ctx context.Context, url string, response interface{}) (http.Header, int, error) {
	return c.signedPost(ctx, url, []byte{}, response)
}

func (c *Client) signedPost(ctx context.Context, rawurl string, content []byte, response interface{}) (http.Header, int, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, 0, errors.Trace(err)
	}

	signedContent, err := c.Jws.SignContent(rawurl, content)
	if err != nil {
		return nil, 0, errors.Trace(err)
	}

	signedBody := []byte(signedContent.FullSerialize())
	ctx = retriable.WithHeaders(ctx, map[string]string{
		header.ContentType:   "application/jose+json",
		header.ContentLength: strconv.Itoa(len(signedBody)),
	})

	return c.Client.PostTo(ctx, []string{u.Scheme + "://" + u.Host}, u.Path, signedBody, response)
}

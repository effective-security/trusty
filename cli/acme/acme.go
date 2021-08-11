package acme

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/ekspand/trusty/acme"
	acmemodel "github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/ekspand/trusty/cli"
	"github.com/ekspand/trusty/pkg/csr"
	"github.com/ekspand/trusty/pkg/inmemcrypto"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
)

var logger = xlog.NewPackageLogger("github.com/ekspand/trusty/cli", "acme")

// GetAccountFlags for GetAccount command
type GetAccountFlags struct {
	OrgID  *string
	EabMAC *string
}

// GetAccount returns the account
func GetAccount(c ctl.Control, p interface{}) error {
	flags := p.(*GetAccountFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	accountsStorage, err := NewAccountsStorage("", client.CurrentHost(), *flags.OrgID)
	if err != nil {
		return errors.Trace(err)
	}
	privateKey, err := accountsStorage.GetPrivateKey()
	if err != nil {
		return errors.Trace(err)
	}

	var account *Account
	if accountsStorage.ExistsAccountFilePath() == nil {
		account, err = accountsStorage.LoadAccount(privateKey)
		if err != nil {
			return errors.Trace(err)
		}
	} else {
		fp, err := acmemodel.GetKeyFingerprint(privateKey)
		if err != nil {
			return errors.Trace(err)
		}

		account = &Account{
			OrgID:       *flags.OrgID,
			Fingerprint: fp,
			key:         privateKey,
		}
	}

	hm, err := base64.RawURLEncoding.DecodeString(*flags.EabMAC)
	if err != nil {
		return errors.Trace(err)
	}

	if account.Registration == nil || account.AccountURL == "" {
		// don't specify keyID for Account
		ac, err := NewClient(client, "", privateKey)
		if err != nil {
			return errors.Trace(err)
		}

		reg, keyID, err := ac.Account(context.Background(), *flags.OrgID, hm, nil)
		if err != nil {
			return errors.Trace(err)
		}

		account.AccountURL = keyID
		account.Registration = reg
		err = accountsStorage.Save(account)
		if err != nil {
			return errors.Trace(err)
		}
	}

	ctl.WriteJSON(c.Writer(), account)
	/* TODO
	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.Account(c.Writer(), account)
	}
	*/
	return nil
}

// RegisterAccountFlags for RegisterAccount command
type RegisterAccountFlags struct {
	OrgID   *string
	EabMAC  *string
	Contact *[]string
}

// RegisterAccount returns the account
func RegisterAccount(c ctl.Control, p interface{}) error {
	flags := p.(*RegisterAccountFlags)
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	accountsStorage, err := NewAccountsStorage("", client.CurrentHost(), *flags.OrgID)
	if err != nil {
		return errors.Trace(err)
	}
	privateKey, err := accountsStorage.GetPrivateKey()
	if err != nil {
		return errors.Trace(err)
	}

	var account *Account
	if accountsStorage.ExistsAccountFilePath() == nil {
		account, err = accountsStorage.LoadAccount(privateKey)
		if err != nil {
			return errors.Trace(err)
		}
	} else {
		fp, err := acmemodel.GetKeyFingerprint(privateKey)
		if err != nil {
			return errors.Trace(err)
		}

		account = &Account{
			OrgID:       *flags.OrgID,
			Fingerprint: fp,
			key:         privateKey,
		}
	}

	hm, err := base64.RawStdEncoding.DecodeString(*flags.EabMAC)
	if err != nil {
		return errors.Trace(err)
	}

	if account.Registration == nil {
		// don't specify keyID for Account
		ac, err := NewClient(client, "", privateKey)
		if err != nil {
			return errors.Trace(err)
		}

		reg, keyID, err := ac.Account(context.Background(), *flags.OrgID, hm, *flags.Contact)
		if err != nil {
			return errors.Trace(err)
		}

		account.AccountURL = keyID
		account.Registration = reg
		err = accountsStorage.Save(account)
		if err != nil {
			return errors.Trace(err)
		}
	}

	ctl.WriteJSON(c.Writer(), account)
	/* TODO
	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.Account(c.Writer(), account)
	}
	*/
	return nil
}

// OrderFlags flags for Order command
type OrderFlags struct {
	OrgID *string
	SPC   *string
}

// Order creates order
func Order(c ctl.Control, p interface{}) error {
	flags := p.(*OrderFlags)
	cli := c.(*cli.Cli)

	spc, err := ioutil.ReadFile(*flags.SPC)
	if err != nil {
		return errors.Annotatef(err, "unable to load SPC file")
	}

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	accountsStorage, err := NewAccountsStorage("", client.CurrentHost(), *flags.OrgID)
	if err != nil {
		return errors.Trace(err)
	}
	privateKey, err := accountsStorage.GetPrivateKey()
	if err != nil {
		return errors.Trace(err)
	}

	var account *Account
	if err = accountsStorage.ExistsAccountFilePath(); err != nil {
		return errors.Annotatef(err, "account does not exist")
	}

	account, err = accountsStorage.LoadAccount(privateKey)
	if err != nil {
		return errors.Annotatef(err, "unable to load account")
	}

	if account.Registration == nil {
		return errors.Errorf("account is not registered")
	}

	// Parse SPC
	claims := &acme.TkClaims{}
	jtoken, _, err := new(jwt.Parser).ParseUnverified(string(spc), claims)
	if err != nil {
		return errors.Annotate(err, "failed to parse SPC token")
	}
	claims, ok := jtoken.Claims.(*acme.TkClaims)
	if !ok {
		return errors.Errorf("invalid SPC token")
	}

	if claims.ATC.TKType != string(v2acme.IdentifierTNAuthList) {
		return errors.Errorf("invalid SPC token type: %s", claims.ATC.TKType)
	}

	// ACME Client
	// don't specify keyID for Account
	ac, err := NewClient(client, account.AccountURL, privateKey)
	if err != nil {
		return errors.Trace(err)
	}

	ctx := context.Background()
	// Order
	orderReq := &v2acme.OrderRequest{
		Identifiers: []v2acme.Identifier{
			{
				Type:  v2acme.IdentifierType(claims.ATC.TKType),
				Value: claims.ATC.TKValue,
			},
		},
	}

	logger.Debugf("status=submitting_order, url=%s", account.Registration.OrdersURL)
	order, orderURL, err := ac.Order(ctx, account.Registration.OrdersURL, orderReq)
	if err != nil {
		return errors.Trace(err)
	}

	var ch *v2acme.Challenge
	for _, authz := range order.Authorizations {
		logger.Debugf("status=get_authorization, url=%s", authz)
		a, err := ac.Authorization(ctx, authz)
		if err != nil {
			return errors.Annotatef(err, "failed to get Authorization")
		}
		logger.KV(xlog.DEBUG, "url", authz, "authorization", a)
		if a.Status != v2acme.StatusPending ||
			a.Identifier.Type != v2acme.IdentifierTNAuthList {
			logger.KV(xlog.DEBUG, "reason", "skip", "url", authz)
			continue
		}

		for _, chall := range a.Challenges {
			if chall.Type == "tkauth-01" {
				ch = &chall
				break
			}
		}
		if ch != nil {
			break
		}
	}

	if ch == nil {
		return errors.New("did not receive tkauth-01 challenge")
	}

	logger.Debugf("status=post_challenge")
	m := map[string]string{
		"atc": strings.TrimSpace(string(spc)),
	}
	_, err = ac.PostChallenge(ctx, ch.URL, m)
	if err != nil {
		return errors.Annotatef(err, "failed to post challenge")
	}

	// start pooling
	for {
		logger.Debugf("status=get_order")
		order, orderURL, err = ac.GetOrder(ctx, orderURL)
		if err != nil {
			return errors.Trace(err)
		}
		if order.Status == v2acme.StatusReady {
			break
		}
		time.Sleep(3 * time.Second)
	}

	logger.Debugf("status=csr")

	prov := csr.NewProvider(inmemcrypto.NewProvider())
	req := &csr.CertificateRequest{
		KeyRequest: prov.NewKeyRequest(*flags.OrgID, "ECDSA", 256, csr.SigningKey),
		Extensions: []csr.X509Extension{
			{
				ID:    csr.OID{1, 3, 6, 1, 5, 5, 7, 1, 26},
				Value: claims.ATC.TKValue,
			},
		},
	}
	csrPEM, key, _, _, err := prov.CreateRequestAndExportKey(req)
	if err != nil {
		return errors.Annotatef(err, "failed to create CSR")
	}

	block, _ := pem.Decode([]byte(csrPEM))
	certRequest, _ := x509.ParseCertificateRequest(block.Bytes)

	creq := v2acme.CertificateRequest{
		CSR: v2acme.JoseBuffer(base64.RawURLEncoding.EncodeToString(certRequest.Raw)),
	}

	order, err = ac.Finalize(ctx, order.FinalizeURL, creq)
	if err != nil {
		return errors.Annotatef(err, "failed to submit CSR")
	}

	if order.Status != v2acme.StatusValid {
		return errors.Errorf("unexpected order status: %s", order.Status)
	}

	certPEM, err := ac.Certificate(ctx, order.CertificateURL)
	if err != nil {
		return errors.Annotatef(err, "failed to download certificate")
	}
	crt, err := certutil.ParseFromPEM(certPEM)
	if err != nil {
		return errors.Annotatef(err, "failed to parse certificate")
	}
	basename := hex.EncodeToString(crt.SubjectKeyId)

	crtFile, keyFile, err := accountsStorage.SaveCert(basename, key, csrPEM, []byte(certPEM))
	if err != nil {
		return errors.Annotatef(err, "unable to save generated files")
	}

	fmt.Fprintf(cli.Writer(), "certificate: %s\nkey: %s\n", crtFile, keyFile)

	return nil
}

func createNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0o700)
	} else if err != nil {
		return err
	}
	return nil
}

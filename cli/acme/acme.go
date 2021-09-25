package acme

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-phorce/dolly/ctl"
	"github.com/go-phorce/dolly/xlog"
	"github.com/go-phorce/dolly/xpki/certutil"
	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/acme"
	acmemodel "github.com/martinisecurity/trusty/acme/model"
	"github.com/martinisecurity/trusty/api/v2acme"
	"github.com/martinisecurity/trusty/cli"
	"github.com/martinisecurity/trusty/pkg/csr"
	"github.com/martinisecurity/trusty/pkg/inmemcrypto"
)

var logger = xlog.NewPackageLogger("github.com/martinisecurity/trusty/cli", "acme")

// Directory returns ACME directory
func Directory(c ctl.Control, _ interface{}) error {
	cli := c.(*cli.Cli)

	client, err := cli.HTTPClient()
	if err != nil {
		return errors.Trace(err)
	}

	dir, err := client.Directory(context.Background())
	if err != nil {
		return errors.Trace(err)
	}

	ctl.WriteJSON(c.Writer(), dir)
	/* TODO
	if cli.IsJSON() {
		ctl.WriteJSON(c.Writer(), res)
		fmt.Fprint(c.Writer(), "\n")
	} else {
		print.Directory(c.Writer(), account)
	}
	*/
	return nil
}

// GetAccountFlags for GetAccount command
type GetAccountFlags struct {
	KeyID  *string
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

	accountsStorage, err := NewAccountsStorage("", client.CurrentHost(), *flags.KeyID)
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
			KeyID:       *flags.KeyID,
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

		reg, keyID, err := ac.Account(context.Background(), *flags.KeyID, hm, nil)
		if err != nil {
			return errors.Annotatef(err, "unable to retrieve account")
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
	KeyID   *string
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

	accountsStorage, err := NewAccountsStorage("", client.CurrentHost(), *flags.KeyID)
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
			//OrgID:       *flags.OrgID,
			KeyID:       *flags.KeyID,
			Fingerprint: fp,
			key:         privateKey,
		}
	}

	hm, err := base64.RawURLEncoding.DecodeString(*flags.EabMAC)
	if err != nil {
		return errors.Annotatef(err, "invalied EAB MAC key")
	}

	if account.Registration == nil {
		// don't specify keyID for Account
		ac, err := NewClient(client, "", privateKey)
		if err != nil {
			return errors.Trace(err)
		}

		reg, keyID, err := ac.Account(context.Background(), *flags.KeyID, hm, *flags.Contact)
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
	KeyID *string
	SPC   *string
	Days  *int
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

	accountsStorage, err := NewAccountsStorage("", client.CurrentHost(), *flags.KeyID)
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

	if flags.Days != nil && *flags.Days > 0 {
		notAfter := time.Now().Add(24 * time.Hour * time.Duration(*flags.Days)).UTC()
		orderReq.NotAfter = notAfter.Format(time.RFC3339)
	}

	order, orderURL, err := ac.Order(ctx, orderReq)
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
		KeyRequest: prov.NewKeyRequest(*flags.KeyID, "ECDSA", 256, csr.SigningKey),
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
		CSR: certRequest.Raw,
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
	basename := getCertBasename(crt)

	crtFile, keyFile, err := accountsStorage.SaveCert(basename, key, csrPEM, []byte(certPEM))
	if err != nil {
		return errors.Annotatef(err, "unable to save generated files")
	}

	fmt.Fprintf(cli.Writer(), "certificate: %s\nkey: %s\n", crtFile, keyFile)

	return nil
}

func getCertBasename(crt *x509.Certificate) string {
	sn := base64.RawURLEncoding.EncodeToString(crt.SerialNumber.Bytes()[:9])
	ikid := certutil.GetAuthorityKeyID(crt)

	return fmt.Sprintf("%s-%s", ikid[:4], sn)
}

func createNonExistingFolder(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0o700)
	} else if err != nil {
		return err
	}
	return nil
}

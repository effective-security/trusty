package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	acmemodel "github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/api/v2acme"
	"github.com/go-acme/lego/certcrypto"
	"github.com/juju/errors"
)

const (
	baseAccountsRootFolderName = "accounts"
	accountFileName            = "account.json"
	keyFileName                = "key"
	baseCertsRootFolderName    = "certificates"
)

// AccountsStorage A storage for account data.
//
// rootPath:
//
//     ~/.mrtsec/accounts/
//          │      └── root accounts directory
//          └── "path" option
//
// rootUserPath:
//
//     ~/.mrtsec/accounts/martini_443/83945903542501476/
//          │      │             │             └── orgID
//          │      │             └── CA server ("server" option)
//          │      └── root accounts directory
//          └── "path" option
//
// keyPath:
//
//     ~/.mrtsec/accounts/martini_443/83945903542501476/key
//          │      │             │             │
//          │      │             │             └── orgID
//          │      │             └── CA server ("server" option)
//          │      └── root accounts directory
//          └── "path" option
//
// accountFilePath:
//
//     ~/.mrtsec/accounts/martini_443/83945903542501476/account.json
//          │      │             │             │             └── account file
//          │      │             │             └── orgID ("keyID" option)
//          │      │             └── CA server ("server" option)
//          │      └── root accounts directory
//          └── "path" option
//
//     ~/.mrtsec/certificates/
//          │      └── root certificates directory
//          └── "path" option

// AccountsStorage represents Account storage
type AccountsStorage struct {
	server          string
	orgID           string
	rootPath        string
	orgPath         string
	certsPath       string
	keyFilePath     string
	accountFilePath string
}

const filePerm os.FileMode = 0o600

// NewAccountsStorage Creates a new AccountsStorage.
func NewAccountsStorage(folder, server string, orgID string) (*AccountsStorage, error) {
	if folder == "" {
		dirname, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Trace(err)
		}

		folder = path.Join(dirname, ".mrtsec")
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, errors.Trace(err)
	}

	rootPath := filepath.Join(folder, baseAccountsRootFolderName)
	certsPath := filepath.Join(folder, baseCertsRootFolderName)
	serverPath := strings.NewReplacer(":", "_", "/", string(os.PathSeparator)).Replace(serverURL.Host)
	accountsPath := filepath.Join(rootPath, serverPath)
	orgPath := filepath.Join(accountsPath, orgID)

	os.MkdirAll(orgPath, 0700)
	os.MkdirAll(certsPath, 0700)

	return &AccountsStorage{
		server:          server,
		orgID:           orgID,
		rootPath:        rootPath,
		orgPath:         orgPath,
		certsPath:       certsPath,
		keyFilePath:     filepath.Join(orgPath, keyFileName),
		accountFilePath: filepath.Join(orgPath, accountFileName),
	}, nil
}

// ExistsAccountFilePath returns error if account file does not exists
func (s *AccountsStorage) ExistsAccountFilePath() error {
	_, err := os.Stat(s.accountFilePath)
	return err
}

// GetRootPath returns Root path
func (s *AccountsStorage) GetRootPath() string {
	return s.rootPath
}

// GetCertificatesPath returns Certificates path
func (s *AccountsStorage) GetCertificatesPath() string {
	return s.certsPath
}

// GetOrgID returns OrgID
func (s *AccountsStorage) GetOrgID() string {
	return s.orgID
}

// Save account
func (s *AccountsStorage) Save(account *Account) error {
	jsonBytes, err := json.MarshalIndent(account, "", "\t")
	if err != nil {
		return errors.Trace(err)
	}

	return ioutil.WriteFile(s.accountFilePath, jsonBytes, filePerm)
}

// LoadAccount returns Account
func (s *AccountsStorage) LoadAccount(privateKey crypto.PrivateKey) (*Account, error) {
	fileBytes, err := ioutil.ReadFile(s.accountFilePath)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to load file for account %s", s.orgID)
	}

	var account Account
	err = json.Unmarshal(fileBytes, &account)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to parse file for account %s", s.orgID)
	}

	account.key = privateKey
	if account.Fingerprint == "" {
		account.Fingerprint, err = acmemodel.GetKeyFingerprint(privateKey)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to fingerprint key for account %s", s.orgID)
		}
	}

	return &account, nil
}

// GetPrivateKey returns PrivateKey
func (s *AccountsStorage) GetPrivateKey() (crypto.PrivateKey, error) {
	if _, err := os.Stat(s.keyFilePath); os.IsNotExist(err) {
		logger.Infof("no key found for account %s. Generating a key.", s.orgID)
		err = createNonExistingFolder(s.orgPath)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to create folder %s", s.keyFilePath)
		}

		privateKey, err := generatePrivateKey(s.keyFilePath)
		if err != nil {
			return nil, errors.Annotatef(err, "unable to generate private key for account %s", s.orgID)
		}

		logger.Infof("saved key to %s", s.keyFilePath)
		return privateKey, nil
	}

	privateKey, err := loadPrivateKey(s.keyFilePath)
	if err != nil {
		return nil, errors.Annotatef(err, "unable to load private key from file %s", s.keyFilePath)
	}

	return privateKey, nil
}

// SaveCert to file
func (s *AccountsStorage) SaveCert(baseName string, key, csrPEM, certPEM []byte) (string, string, error) {
	path := path.Join(s.certsPath, baseName)
	var err error

	certName := path + ".pem"
	keyName := path + ".key"

	if len(certPEM) > 0 {
		err = ioutil.WriteFile(certName, certPEM, 0664)
		if err != nil {
			return "", "", errors.Trace(err)
		}
	}
	if len(csrPEM) > 0 {
		err = ioutil.WriteFile(path+".csr", csrPEM, 0664)
		if err != nil {
			return "", "", errors.Trace(err)
		}
	}
	if len(key) > 0 {
		err = ioutil.WriteFile(keyName, key, 0600)
		if err != nil {
			return "", "", errors.Trace(err)
		}
	}
	return certName, keyName, nil
}

func generatePrivateKey(file string) (crypto.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, errors.Trace(err)
	}

	certOut, err := os.Create(file)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer certOut.Close()

	pemKey := certcrypto.PEMBlock(privateKey)
	err = pem.Encode(certOut, pemKey)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return privateKey, nil
}

func loadPrivateKey(file string) (crypto.PrivateKey, error) {
	keyBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, errors.Trace(err)
	}

	keyBlock, _ := pem.Decode(keyBytes)

	switch keyBlock.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(keyBlock.Bytes)
	}

	return nil, errors.New("unknown private key type")
}

// Account represents a users local saved credentials.
type Account struct {
	OrgID        string          `json:"org_id"`
	AccountURL   string          `json:"account_url"`
	Registration *v2acme.Account `json:"registration"`
	Fingerprint  string          `json:"fingerprint"`
	key          crypto.PrivateKey
}

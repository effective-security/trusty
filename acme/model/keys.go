package model

import (
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"sort"
	"strings"

	"github.com/ekspand/trusty/api/v2acme"
	"github.com/juju/errors"
	"gopkg.in/square/go-jose.v2"
)

// GetKeyID produces Base64-URL-encoded SHA256 digest of a provided public key.
func GetKeyID(key crypto.PublicKey) (string, error) {
	switch t := key.(type) {
	case *jose.JSONWebKey:
		if t == nil {
			return "", errors.Errorf("nil key")
		}
		return GetKeyID(t.Key)
	case jose.JSONWebKey:
		return GetKeyID(t.Key)
	default:
		keyDER, err := x509.MarshalPKIXPublicKey(key)
		if err != nil {
			return "", errors.Trace(err)
		}
		spkiDigest := sha256.Sum256(keyDER)

		return base64.RawURLEncoding.EncodeToString(spkiDigest[0:32]), nil
	}
}

// GetIDFromStrings produces Base64-URL-encoded SHA256 digest of a slice
func GetIDFromStrings(list []string) (string, error) {
	h := sha256.New()
	for _, str := range list {
		_, err := h.Write([]byte(str))
		if err != nil {
			return "", errors.Trace(err)
		}
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)), nil
}

// GetIDFromIdentifiers produces Base64-URL-encoded SHA256 digest of a slice
func GetIDFromIdentifiers(list []v2acme.Identifier) (string, error) {
	h := sha256.New()
	for _, idn := range list {
		_, err := h.Write([]byte(idn.Value))
		if err != nil {
			return "", errors.Trace(err)
		}
	}
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil)), nil
}

// UniqueLowerNames returns the set of all unique names in the input after all
// of them are lowercased. The returned names will be in their lowercased form
// and sorted alphabetically.
func UniqueLowerNames(names []string) (unique []string) {
	nameMap := make(map[string]int, len(names))
	for _, name := range names {
		nameMap[strings.ToLower(name)] = 1
	}

	unique = make([]string, 0, len(nameMap))
	for name := range nameMap {
		unique = append(unique, name)
	}
	sort.Strings(unique)
	return
}

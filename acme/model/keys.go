package model

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"sort"
	"strings"

	"github.com/juju/errors"
	"github.com/martinisecurity/trusty/api/v2acme"
	"gopkg.in/square/go-jose.v2"
)

// GetKeyID produces Base64-URL-encoded SHA256 digest of a provided public key.
func GetKeyID(key crypto.PublicKey) (string, error) {
	if key == nil {
		return "", errors.Errorf("nil key")
	}
	switch t := key.(type) {
	case *jose.JSONWebKey:
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

// GetKeyFingerprint produces hex-encoded SHA256 digest of a provided public key,
// prefixed with SHA256 and the hash value is represented as a sequence of
// uppercase hexadecimal bytes, separated by colons.
func GetKeyFingerprint(key crypto.PublicKey) (string, error) {
	if key == nil {
		return "", errors.Errorf("nil key")
	}
	switch t := key.(type) {
	case *jose.JSONWebKey:
		return GetKeyFingerprint(t.Key)
	case jose.JSONWebKey:
		return GetKeyFingerprint(t.Key)
	case *ecdsa.PrivateKey:
		return GetKeyFingerprint(&t.PublicKey)
	case *rsa.PrivateKey:
		return GetKeyFingerprint(&t.PublicKey)
	default:
		keyDER, err := x509.MarshalPKIXPublicKey(key)
		if err != nil {
			return "", errors.Trace(err)
		}
		spkiDigest := sha256.Sum256(keyDER)

		const hextable = "0123456789ABCDEF"
		dst := make([]byte, 2*32+31)
		j := 0
		for _, v := range spkiDigest[0:32] {
			if j > 1 {
				dst[j] = ':'
				j++
			}
			dst[j] = hextable[v>>4]
			dst[j+1] = hextable[v&0x0f]
			j += 2
		}

		return "SHA256 " + string(dst), nil
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

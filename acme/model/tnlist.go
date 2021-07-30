package model

import (
	"encoding/asn1"
	"encoding/base64"

	"github.com/juju/errors"
)

/*
TNAuthorizationList ::= SEQUENCE SIZE (1..MAX) OF TNEntry

TNEntry ::= CHOICE {
  spc   [0] ServiceProviderCode,
  range [1] TelephoneNumberRange,
  one   [2] TelephoneNumber
  }

ServiceProviderCode ::= IA5String

-- SPCs may be OCNs, various SPIDs, or other SP identifiers
-- from the telephone network.
TelephoneNumberRange ::= SEQUENCE {
  start TelephoneNumber,
  count INTEGER (2..MAX),
  ...
  }

TelephoneNumber ::= IA5String (SIZE (1..15)) (FROM ("0123456789#*"))
*/

// TNEntry proves TN Entry
type TNEntry struct {
	SPC ServiceProviderCode `asn1:"optional,tag:0"`
}

// ServiceProviderCode is IA5
type ServiceProviderCode struct {
	Code string `asn1:"ia5"`
}

// Base64 returns base64 encoded string
func (t *TNEntry) Base64() (string, error) {
	b, err := asn1.Marshal(*t)
	if err != nil {
		return "", errors.Trace(err)
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

// ParseTNEntry returns TNEntry
func ParseTNEntry(b64 string) (*TNEntry, error) {
	der, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, errors.Trace(err)
	}

	t := new(TNEntry)
	_, err = asn1.Unmarshal(der, t)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return t, nil
}

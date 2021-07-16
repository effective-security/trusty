package v2acme

import (
	"encoding/base64"
	"encoding/json"
)

// JoseBuffer fields get encoded and decoded JOSE-style, in base64url encoding
// with stripped padding.
type JoseBuffer []byte

// MarshalJSON encodes a JoseBuffer for transmission.
func (jb JoseBuffer) MarshalJSON() (result []byte, err error) {
	return json.Marshal(base64.RawURLEncoding.EncodeToString(jb))
}

// UnmarshalJSON decodes a JoseBuffer to an object.
func (jb *JoseBuffer) UnmarshalJSON(data []byte) (err error) {
	var str string
	err = json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	*jb, err = base64.RawURLEncoding.DecodeString(str)
	return
}

package authority

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ocsp"
)

func TestOCSPReasonStringToCode(t *testing.T) {
	tcases := []struct {
		reason string
		code   int
		err    string
	}{
		{"", ocsp.Unspecified, ""},
		{"unspecified", ocsp.Unspecified, ""},
		{"keycompromise", ocsp.KeyCompromise, ""},
		{"cacompromise", ocsp.CACompromise, ""},
		{"affiliationchanged", ocsp.AffiliationChanged, ""},
		{"superseded", ocsp.Superseded, ""},
		{"cessationofoperation", ocsp.CessationOfOperation, ""},
		{"certificatehold", ocsp.CertificateHold, ""},
		{"removefromcrl", ocsp.RemoveFromCRL, ""},
		{"privilegewithdrawn", ocsp.PrivilegeWithdrawn, ""},
		{"aacompromise", ocsp.AACompromise, ""},
		{"10", ocsp.AACompromise, ""},
		{"11", 11, "invalid status: 11"},
		{"-11", -11, "invalid status: -11"},
		{"xxx", 0, `strconv.Atoi: parsing "xxx": invalid syntax`},
	}

	for _, tc := range tcases {
		rc, err := OCSPReasonStringToCode(tc.reason)
		if tc.err != "" {
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error(), tc.reason)
		} else {
			require.NoError(t, err)
			assert.Equal(t, tc.code, rc, tc.reason)
		}
	}
}

package certutil

import (
	"crypto/x509"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_notAfter(t *testing.T) {
	now := time.Now()

	list := []*x509.Certificate{
		{
			NotAfter: now.Add(-1 * time.Hour),
		},
		{
			NotAfter: now.Add(time.Hour),
		},
	}
	l1 := filterByNotAfter(list, now)
	assert.Len(t, l1, 1)

	l2 := filterByAfter(list, now)
	assert.Len(t, l2, 1)
}

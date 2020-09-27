package print_test

import (
	"bytes"
	"testing"

	"github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/go-phorce/trusty/pkg/print"
	"github.com/stretchr/testify/assert"
)

func TestPrintServerVersion(t *testing.T) {
	r := &trustypb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}
	w := bytes.NewBuffer([]byte{})

	print.ServerVersion(w, r)

	out := string(w.Bytes())
	assert.Equal(t, "1.1.1 (go1.15.1)\n", out)
}

package main

import (
	"bytes"
	"testing"

	"github.com/effective-security/trusty/internal/version"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	out := bytes.NewBuffer([]byte{})
	errout := bytes.NewBuffer([]byte{})
	rc := 0
	exit := func(c int) {
		rc = c
	}

	realMain([]string{"trustyctl", "ver"}, out, errout, exit)
	assert.Equal(t, 1, rc)
	assert.Equal(t, "trustyctl: error: unexpected argument ver, did you mean \"version\"?\n", errout.String())
	assert.Empty(t, out.String())

	out = bytes.NewBuffer([]byte{})
	errout = bytes.NewBuffer([]byte{})
	rc = 0
	realMain([]string{"trustyctl", "--version"}, out, errout, exit)
	assert.Equal(t, version.Current().String()+"\n", out.String())
	// since our exit func does not call os.Exit, the next parser will fail
	assert.Equal(t, 1, rc)
	assert.NotEmpty(t, errout.String())
}

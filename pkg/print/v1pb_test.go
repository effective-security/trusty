package print_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

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

func TestServerStatusResponse(t *testing.T) {
	ver := &trustypb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}

	now := time.Now()

	r := &trustypb.ServerStatusResponse{
		Status: &trustypb.ServerStatus{
			Hostname:   "dissoupov",
			ListenUrls: []string{"https://0.0.0.0:7891"},
			Name:       "Trusty",
			Nodename:   "local",
			StartedAt:  now.Unix(),
		},
		Version: ver,
	}

	w := bytes.NewBuffer([]byte{})

	print.ServerStatusResponse(w, r)

	out := string(w.Bytes())
	assert.Contains(t, out, "  Name        | Trusty ")
	assert.Contains(t, out, "  Node        | local  ")
	assert.Contains(t, out, "  Host        | dissoupov ")
	assert.Contains(t, out, "  Listen URLs | https://0.0.0.0:7891")
	assert.Contains(t, out, "  Version     | 1.1.1    ")
	assert.Contains(t, out, "  Runtime     | go1.15.1 ")
	assert.Contains(t, out, fmt.Sprintf("  Started     | %s ", now.Format(time.RFC3339)))
	assert.Contains(t, out, "  Uptime      | 0s ")
}

func TestCallerStatusResponse(t *testing.T) {
	r := &trustypb.CallerStatusResponse{
		Id:   "12341234-1234124",
		Name: "local",
		Role: "trustry",
	}

	w := bytes.NewBuffer([]byte{})

	print.CallerStatusResponse(w, r)

	out := string(w.Bytes())
	assert.Equal(t, "  Name | local             \n"+
		"  ID   | 12341234-1234124  \n"+
		"  Role | trustry           \n\n", out)
}

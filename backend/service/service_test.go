package service_test

import (
	"net/http"
	"testing"

	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/martinisecurity/trusty/backend/service"
	"github.com/martinisecurity/trusty/backend/service/ca"
	"github.com/martinisecurity/trusty/backend/service/status"
	"github.com/martinisecurity/trusty/backend/service/swagger"
	"github.com/martinisecurity/trusty/pkg/gserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var serviceFactories = map[string]gserver.ServiceFactory{
	ca.ServiceName:      ca.Factory,
	status.ServiceName:  status.Factory,
	swagger.ServiceName: swagger.Factory,
}

func Test_invalidArgs(t *testing.T) {
	for _, f := range serviceFactories {
		testInvalidServiceArgs(t, f)
	}
}

func TestGetPublicServerURL(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, v1.PathForAuthURL, nil)
	require.NoError(t, err)

	u := service.GetPublicServerURL(r, "/v1").String()
	assert.Equal(t, "https:///v1", u)

	r.URL.Scheme = "https"
	r.Host = "mrtsec.io:8443"
	u = service.GetPublicServerURL(r, "/v1").String()
	assert.Equal(t, "https://mrtsec.io:8443/v1", u)
}

func testInvalidServiceArgs(t *testing.T, f gserver.ServiceFactory) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Expected panic but didn't get one")
		}
	}()
	f(nil)
}

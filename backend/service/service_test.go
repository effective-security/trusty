package service_test

import (
	"testing"

	"github.com/go-phorce/trusty/backend/service/status"
	"github.com/go-phorce/trusty/backend/trustyserver"
)

var serviceFactories = map[string]trustyserver.ServiceFactory{
	status.ServiceName: status.Factory,
}

func Test_invalidArgs(t *testing.T) {
	for _, f := range serviceFactories {
		testInvalidServiceArgs(t, f)
	}
}

func testInvalidServiceArgs(t *testing.T, f trustyserver.ServiceFactory) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Expected panic but didn't get one")
		}
	}()
	f(nil)
}

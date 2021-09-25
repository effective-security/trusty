package auth_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/martinisecurity/trusty/cli/auth"
	"github.com/martinisecurity/trusty/cli/testsuite"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	testsuite.Suite
}

func TestCtlSuite(t *testing.T) {
	s := new(testSuite)
	s.WithGRPC()
	suite.Run(t, s)
}

func (s *testSuite) TestAuth() {
	h := makeTestHandler(s.T(), `{
		"url": "eyJhbGciOiJIUzI1NiIsImtpZCI6IjAiLCJ0eXAiOiJKV1QifQ.eyJ0cnVzdHkiOnsiaWQiOiI3MjQzNTE1NDkzMjQ2NTc2NCIsImV4dGVybl9pZCI6IjI1NTg5MjAiLCJwcm92aWRlciI6ImdpdGh1YiIsImxvZ2luIjoiZGlzc291cG92IiwibmFtZSI6IkRlbmlzIElzc291cG92IiwiZW1haWwiOiJkZW5pc0Bla3NwYW5kLmNvbSIsImNvbXBhbnkiOiJnby1waG9yY2UiLCJhdmF0YXJfdXJsIjoiaHR0cHM6Ly9hdmF0YXJzLmdpdGh1YnVzZXJjb250ZW50LmNvbS91LzI1NTg5MjA_dj00In0sImRldmljZV9pZCI6ImRpc3NvdXBvdi13c2wyIiwiYXVkIjoidHJ1c3R5LWRldiIsImV4cCI6MTYyMTA0NTM0NCwiaWF0IjoxNjIxMDE2NTQ0LCJpc3MiOiJ0cnVzdHktZGV2Iiwic3ViIjoiZGVuaXNAZWtzcGFuZC5jb20ifQ.g2Z6QT7E-MvSQACYGpMWynNXNAIWLNrOcKD7HJrBj_U"
	}`)
	server := httptest.NewServer(h)
	defer server.Close()

	s.Cli.WithServer(server.URL)

	disableBrowser := true
	provider := "github"
	err := s.Run(auth.Authenticate, &auth.AuthenticateFlags{NoBrowser: &disableBrowser, Provider: &provider})
	s.Require().NoError(err)
	s.HasText("open auth URL in browser:\n")
}

func makeTestHandler(t *testing.T, responseBody string) http.Handler {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, responseBody)
	}
	return http.HandlerFunc(h)
}

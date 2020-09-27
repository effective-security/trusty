package status_test

import (
	"testing"

	"github.com/go-phorce/trusty/api/v1/trustypb"
	"github.com/go-phorce/trusty/cli/status"
	"github.com/go-phorce/trusty/cli/testsuite"
	"github.com/go-phorce/trusty/tests/mockpb"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
)

const (
	projFolder = "../../"
	loopBackIP = "127.0.0.1"
)

type testSuite struct {
	testsuite.Suite
}

func TestCtlSuite(t *testing.T) {
	s := new(testSuite)
	s.WithGRPC()
	suite.Run(t, s)
}

func TestCtlSuiteWithJSON(t *testing.T) {
	s := new(testSuite)
	s.WithGRPC().WithAppFlags([]string{"--json"})
	suite.Run(t, s)
}

func (s *testSuite) TestVersion() {
	expectedResponse := &trustypb.ServerVersion{
		Build:   "1.2.3",
		Runtime: "go1.15",
	}

	s.MockStatus = &mockpb.MockStatusServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	err := s.Run(status.Version, nil)
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText("{\n\t\"Build\": \"1.2.3\",\n\t\"Runtime\": \"go1.15\"\n}\n")
	} else {
		s.HasText("1.2.3\n")
	}
}

func (s *testSuite) TestServer() {
	expectedResponse := &trustypb.ServerStatusResponse{
		Status: &trustypb.ServerStatus{
			Name:       "mock",
			ListenURLs: []string{"host1:123"},
		},
		Version: &trustypb.ServerVersion{
			Build:   "1.2.3",
			Runtime: "go1.15",
		},
	}

	s.MockStatus = &mockpb.MockStatusServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	err := s.Run(status.Server, nil)
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText("{\n\t\"Status\": {\n\t\t\"ListenURLs\": [\n\t\t\t\"host1:123\"\n\t\t],\n\t\t\"Name\": \"mock\"\n\t},\n\t\"Version\": {\n\t\t\"Build\": \"1.2.3\",\n\t\t\"Runtime\": \"go1.15\"\n\t}\n}\n")
	} else {
		// TODO
		s.HasText("{\n\t\"Status\": {\n\t\t\"ListenURLs\": [\n\t\t\t\"host1:123\"\n\t\t],\n\t\t\"Name\": \"mock\"\n\t},\n\t\"Version\": {\n\t\t\"Build\": \"1.2.3\",\n\t\t\"Runtime\": \"go1.15\"\n\t}\n}\n")
	}
}

func (s *testSuite) TestCaller() {
	expectedResponse := &trustypb.CallerStatusResponse{
		Role: "test_role",
	}

	s.MockStatus = &mockpb.MockStatusServer{
		Err:   nil,
		Resps: []proto.Message{expectedResponse},
	}
	srv := s.SetupMockGRPC()
	defer srv.Stop()

	err := s.Run(status.Caller, nil)
	s.Require().NoError(err)

	if s.Cli.IsJSON() {
		s.HasText("{\n\t\"Role\": \"test_role\"\n}\n")
	} else {
		// TODO
		s.HasText("{\n\t\"Role\": \"test_role\"\n}\n")
	}
}

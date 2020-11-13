package status_test

import (
	"testing"

	"github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/ekspand/trusty/cli/status"
	"github.com/ekspand/trusty/cli/testsuite"
	"github.com/ekspand/trusty/tests/mockpb"
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
		s.HasText("{\n\t\"build\": \"1.2.3\",\n\t\"runtime\": \"go1.15\"\n}\n")
	} else {
		s.HasText("1.2.3")
	}
}

func (s *testSuite) TestServer() {
	expectedResponse := &trustypb.ServerStatusResponse{
		Status: &trustypb.ServerStatus{
			Name:       "mock",
			ListenUrls: []string{"host1:123"},
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
		s.HasText("{\n\t\"status\": {\n\t\t\"listen_urls\": [\n\t\t\t\"host1:123\"\n\t\t],\n\t\t\"name\": \"mock\"\n\t},\n\t\"version\": {\n\t\t\"build\": \"1.2.3\",\n\t\t\"runtime\": \"go1.15\"\n\t}\n}\n")
	} else {
		s.HasText("  Name        | mock ",
			"  Node        |            ",
			"  Host        |            ",
			"  Listen URLs | host1:123  ",
			"  Version     | 1.2.3      ",
			"  Runtime     | go1.15     ",
			"  Started     |")
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
		s.HasText("{\n\t\"role\": \"test_role\"\n}\n")
	} else {
		s.HasText("  Name |            \n  ID   |            \n  Role | test_role  \n\n")
	}
}

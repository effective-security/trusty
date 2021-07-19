package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ekspand/trusty/internal/config"
	"github.com/ekspand/trusty/internal/version"
	"github.com/go-phorce/dolly/algorithms/guid"
	"github.com/go-phorce/dolly/ctl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var testDirPath = filepath.Join(os.TempDir(), "/tests/trusty/cmd", "martinictl-"+guid.MustCreate())

const projFolder = "../../"

type testSuite struct {
	suite.Suite
	baseArgs []string
	out      bytes.Buffer
}

func (s *testSuite) run(additionalFlags ...string) ctl.ReturnCode {
	s.out.Reset()
	rc := realMain(append(s.baseArgs, additionalFlags...), &s.out, &s.out)
	return rc
}

// hasText is a helper method to assert that the out stream contains the supplied
// text somewhere
func (s *testSuite) hasText(t string) {
	outStr := s.out.String()
	s.True(strings.Index(outStr, t) >= 0, "Expecting to find text %q in value %q", t, outStr)
}

func (s *testSuite) hasNoText(t string) {
	outStr := s.out.String()
	s.True(strings.Index(outStr, t) < 0, "Expecting to NOT find text %q in value %q", t, outStr)
}

func (s *testSuite) hasFile(file string) {
	stat, err := os.Stat(file)
	s.Require().NoError(err, "File must exist: %s", file)
	s.Require().False(stat.IsDir())
}

func (s *testSuite) hasTextInFile(file, t string) {
	f, err := ioutil.ReadFile(file)
	s.Require().NoError(err, "Unable to read: %s", file)

	s.True(strings.Index(string(f), t) >= 0, "Expecting to find text %q in file %q", t, file)
}

func TestCtlSuite(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) SetupTest() {
	cfg, err := filepath.Abs(projFolder + "etc/dev/" + config.ConfigFileName)
	s.Require().NoError(err)

	s.baseArgs = []string{"martinictl", "-V", "-D", "--json", "--cfg", cfg}
}

func (s *testSuite) TearDownTest() {
}

func TestGoVersion(t *testing.T) {
	gv := runtime.Version()
	vsCheck := strings.HasPrefix(gv, "go1.15") || strings.HasPrefix(gv, "go1.16")
	assert.True(t, vsCheck, "should be built with go 1.16.+, got: %s", gv)

	v := version.Current()
	assert.True(t, v.Float() > 0)
	assert.NotEmpty(t, v.Runtime)
}

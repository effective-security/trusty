package trustypb_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/ekspand/trusty/api/v1/trustypb"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Decode_ValidDTO(t *testing.T) {
	var serverVersion trustypb.ServerVersion
	var serverStatus trustypb.ServerStatusResponse
	var callerStatus trustypb.CallerStatusResponse

	expVer := &trustypb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}

	expStatus := &trustypb.ServerStatusResponse{
		Status: &trustypb.ServerStatus{
			Hostname:   "dissoupov-ltl2",
			ListenUrls: []string{"https://0.0.0.0:7891"},
			Name:       "Trusty",
			StartedAt:  1601216552,
		},
		Version: expVer,
	}
	expCaller := &trustypb.CallerStatusResponse{
		Id:   "local",
		Name: "localhost",
		Role: "trusty-peer",
	}

	tcases := []struct {
		filename string
		dto      interface{}
		exp      interface{}
	}{
		{"testdata/ServerStatusResponse.json", &serverStatus, expStatus},
		{"testdata/CallerStatusResponse.json", &callerStatus, expCaller},
		{"testdata/ServerVersion.json", &serverVersion, expVer},
	}

	for _, tc := range tcases {
		t.Run(
			tc.filename,
			func(t *testing.T) {
				err := loadJson(tc.filename, tc.dto)
				require.NoError(t, err)
				require.NotNil(t, tc.dto)

				if tc.exp != nil {
					assert.Equal(t, tc.exp, tc.dto)
				}
			},
		)
	}
}

func loadJson(filename string, v interface{}) error {
	cfr, err := os.Open(filename)
	if err != nil {
		return errors.Trace(err)
	}
	defer cfr.Close()
	err = json.NewDecoder(cfr).Decode(v)
	if err != nil {
		return errors.Trace(err)
	}
	return nil
}

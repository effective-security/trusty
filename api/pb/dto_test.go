package pb_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/effective-security/trusty/api/pb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Decode_ValidDTO(t *testing.T) {
	var serverVersion pb.ServerVersion
	var serverStatus pb.ServerStatusResponse
	var callerStatus pb.CallerStatusResponse

	expVer := &pb.ServerVersion{
		Build:   "1.1.1",
		Runtime: "go1.15.1",
	}

	expStatus := &pb.ServerStatusResponse{
		Status: &pb.ServerStatus{
			Hostname:   "dissoupov",
			ListenURLs: []string{"https://0.0.0.0:7891"},
			Name:       "Trusty",
			StartedAt:  "2020-10-01T00:00:00Z",
		},
		Version: expVer,
	}
	expCaller := &pb.CallerStatusResponse{
		Subject: "localhost",
		Role:    "trusty-peer",
	}

	tcases := []struct {
		filename string
		dto      any
		exp      any
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

				if tc.exp != nil {
					js1, _ := json.Marshal(tc.exp)
					js2, _ := json.Marshal(tc.dto)
					assert.EqualValues(t, js1, js2, tc.filename)
				}
			},
		)
	}
}

func loadJson(filename string, v any) error {
	cfr, err := os.Open(filename)
	if err != nil {
		return errors.WithStack(err)
	}
	defer cfr.Close()
	err = json.NewDecoder(cfr).Decode(v)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

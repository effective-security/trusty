package appcontainer_test

import (
	"context"
	"testing"

	"github.com/ekspand/trusty/api/v1/pb"
	"github.com/ekspand/trusty/internal/appcontainer"
	"github.com/ekspand/trusty/tests/mockpb"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscovery(t *testing.T) {
	ca := &mockpb.MockCAServer{}
	ra := &mockpb.MockRAServer{}
	cis := &mockpb.MockCIServer{}

	srv := "TestDiscovery"
	d := appcontainer.NewDiscovery()
	err := d.Register(srv, ca)
	require.NoError(t, err)
	err = d.Register(srv, ra)
	require.NoError(t, err)

	err = d.Register(srv, cis)
	require.NoError(t, err)
	err = d.Register(srv, cis)
	require.Error(t, err)
	assert.Equal(t, "already registered: TestDiscovery/*mockpb.MockCIServer", err.Error())

	var pbCA pb.CAServiceServer
	err = d.Find(&pbCA)
	require.NoError(t, err)
	require.NotNil(t, pbCA)

	ca.SetResponse(&pb.IssuersInfoResponse{})
	issResp, err := pbCA.Issuers(context.Background(), &empty.Empty{})
	require.NoError(t, err)
	require.NotNil(t, issResp)

	count := 0
	err = d.ForEach(&pbCA, func(key string) error {
		count++
		return nil
	})
	require.NoError(t, err)
	require.NotNil(t, pbCA)
	assert.Equal(t, 1, count)

	var nonPointer pb.CAServiceServer
	err = d.Find(nonPointer)
	require.Error(t, err)
	assert.Equal(t, "a pointer to interface is required, invalid type: <invalid reflect.Value>", err.Error())

	err = d.Find(err)
	require.Error(t, err)
	assert.Equal(t, "non interface type: *errors.Err", err.Error())

	err = d.Find(&err)
	require.Error(t, err)
	assert.Equal(t, "not implemented: <error Value>", err.Error())

	err = d.ForEach(nonPointer, func(key string) error {
		return nil
	})
	require.Error(t, err)
	assert.Equal(t, "a pointer to interface is required, invalid type: <invalid reflect.Value>", err.Error())

	err = d.ForEach(err, func(key string) error {
		return nil
	})
	require.Error(t, err)
	assert.Equal(t, "non interface type: *errors.Err", err.Error())

	err = d.ForEach(&pbCA, func(key string) error {
		return errors.Errorf("unable to do something")
	})
	require.Error(t, err)
	assert.Equal(t, "failed to execute callback for *mockpb.MockCAServer: unable to do something", err.Error())
}

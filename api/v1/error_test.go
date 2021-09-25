package v1_test

import (
	"errors"
	"testing"

	v1 "github.com/martinisecurity/trusty/api/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestTrustyError(t *testing.T) {
	e1 := status.New(codes.PermissionDenied, "trusty: permission denied").Err()
	e2 := v1.ErrGRPCPermissionDenied
	e3 := v1.ErrPermissionDenied

	require.Equal(t, e1.Error(), e2.Error())
	assert.NotEqual(t, e1.Error(), e3.Error())

	ev1, ok := status.FromError(e1)
	assert.True(t, ok)
	assert.Equal(t, ev1.Code(), e3.(v1.TrustyError).Code())

	ev2, ok := status.FromError(e2)
	assert.True(t, ok)
	assert.Equal(t, ev2.Code(), e3.(v1.TrustyError).Code())

	assert.Nil(t, v1.Error(nil))

	someErr := errors.New("some error")
	assert.NotNil(t, v1.Error(someErr))
	assert.Equal(t, "some error", v1.ErrorDesc(someErr))

	assert.NotNil(t, v1.Error(e3))
	assert.Equal(t, "trusty: permission denied", v1.ErrorDesc(e3))
}

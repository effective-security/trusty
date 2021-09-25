package db_test

import (
	"errors"
	"testing"
	"time"

	model "github.com/martinisecurity/trusty/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NullTime(t *testing.T) {
	v := model.NullTime(nil)
	require.NotNil(t, v)
	assert.False(t, v.Valid)

	i := time.Now()
	v = model.NullTime(&i)
	require.NotNil(t, v)
	assert.True(t, v.Valid)
	assert.Equal(t, i, v.Time)
}

func TestString(t *testing.T) {
	v := model.String(nil)
	assert.Empty(t, v)

	s := "1234"
	v = model.String(&s)
	assert.Equal(t, s, v)
}

func TestID(t *testing.T) {
	_, err := model.ID("")
	require.Error(t, err)

	_, err = model.ID("@123")
	require.Error(t, err)

	v, err := model.ID("1234567")
	require.NoError(t, err)
	assert.Equal(t, uint64(1234567), v)
}

type validator struct {
	valid bool
}

func (t validator) Validate() error {
	if !t.valid {
		return errors.New("invalid")
	}
	return nil
}

func TestValidate(t *testing.T) {
	assert.Error(t, model.Validate(validator{false}))
	assert.NoError(t, model.Validate(validator{true}))
	assert.NoError(t, model.Validate(nil))
}

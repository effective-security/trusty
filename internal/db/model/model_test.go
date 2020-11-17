package model_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ekspand/trusty/internal/db/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	longVal = "hlvksjdhfvlkjsdfhlbkshjdflkjvhsldfkjvhlskdfjvhlakjfvhlakjfvhlakjhvlkajshvlkajshvlkjahsdlvkjahslkvjhalskdjvhaklsdvjhaklsjdvhalksjvhalkjsvhlakshvakljshvlkasjhvkajshvkajhvkajhlvkahlfkvj"
	longURL = "http://jfhsdjfghsjdfghsdfg.sdfhgslkfdhgslkjdfhglskjdfhgslkjdfhglskdjfhglskjdfhgslkdjfhglksjdfhgskjdfhglksjdfhglksjdfhg.com?hlvksjdhfvlkjsdfhlbkshjdflkjvhsldfkjvhlskdfjvhlakjfvhlakjfvhlakjhvlkajshvlkajshvlkjahsdlvkjahslkvjhalskdjvhaklsdvjhaklsjdvhalksjvhalkjsvhlakshvakljshvlkasjhvkajshvkajhvkajhlvkahlfkvj"
)

func TestUser(t *testing.T) {
	tcases := []struct {
		u   *model.User
		err string
	}{
		{&model.User{}, "invalid name: \"\""},
		{&model.User{Name: longVal}, fmt.Sprintf("invalid name: %q", longVal)},
		{&model.User{Name: "n1", Login: longVal}, fmt.Sprintf("invalid login: %q", longVal)},
		{&model.User{Name: "n1", Login: "l1", Email: longVal}, fmt.Sprintf("invalid email: %q", longVal)},
		{&model.User{Name: "n1", Login: "l1", Email: "e1", Company: longVal}, fmt.Sprintf("invalid company: %q", longVal)},
		{&model.User{Name: "n1", Login: "l1", Email: "e1", Company: "c1", AvatarURL: longURL}, fmt.Sprintf("invalid avatar: %q", longURL)},
	}
	for _, tc := range tcases {
		err := tc.u.Validate()
		if tc.err != "" {
			require.Error(t, err)
			assert.Equal(t, tc.err, err.Error())
		} else {
			assert.NoError(t, err)
		}
	}

	u := &model.User{ID: 1000, Name: "n1", Login: "l1", Email: "e1", Company: "c1", AvatarURL: "https://github.com/me"}
	dto := u.ToDto()
	assert.Equal(t, "1000", dto.ID)
	assert.Equal(t, u.Login, dto.Login)
	assert.Equal(t, u.Name, dto.Name)
	assert.Equal(t, u.Email, dto.Email)
	assert.Equal(t, u.Company, dto.Company)
	assert.Equal(t, u.AvatarURL, dto.AvatarURL)
}

func TestNullInt64(t *testing.T) {
	v := model.NullInt64(nil)
	require.NotNil(t, v)
	assert.False(t, v.Valid)

	i := int64(10000)
	v = model.NullInt64(&i)
	require.NotNil(t, v)
	assert.True(t, v.Valid)
	assert.Equal(t, i, v.Int64)
}

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
	v, err := model.ID("")
	require.Error(t, err)

	v, err = model.ID("@123")
	require.Error(t, err)

	v, err = model.ID("1234567")
	require.NoError(t, err)
	assert.Equal(t, int64(1234567), v)
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

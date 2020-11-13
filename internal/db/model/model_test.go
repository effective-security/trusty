package model_test

import (
	"fmt"
	"testing"

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

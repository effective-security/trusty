package model_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/ekspand/trusty/internal/db/orgsdb/model"
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

	u := &model.User{ID: 1000, Name: "n1", Login: "l1", Email: "e1", Company: "c1", ExternalID: "exID", AvatarURL: "https://github.com/me"}
	dto := u.ToDto()
	assert.Equal(t, "1000", dto.ID)
	assert.Equal(t, u.Login, dto.Login)
	assert.Equal(t, u.Name, dto.Name)
	assert.Equal(t, u.Email, dto.Email)
	assert.Equal(t, u.Company, dto.Company)
	assert.Equal(t, u.AvatarURL, dto.AvatarURL)
	assert.Equal(t, u.ExternalID, dto.ExternalID)
}

func TestOrganization(t *testing.T) {
	u := &model.Organization{
		ID:           1000,
		ExternalID:   "1001",
		Name:         "n1",
		Login:        "l1",
		Email:        "e1",
		BillingEmail: "b1",
		Company:      "c1",
		Location:     "wa",
		AvatarURL:    "https://github.com/me",
		Type:         "Organization",
	}
	dto := u.ToDto()
	assert.Equal(t, "1000", dto.ID)
	assert.Equal(t, "1001", dto.ExternalID)
	assert.Equal(t, u.Login, dto.Login)
	assert.Equal(t, u.Name, dto.Name)
	assert.Equal(t, u.Email, dto.Email)
	assert.Equal(t, u.BillingEmail, dto.BillingEmail)
	assert.Equal(t, u.Company, dto.Company)
	assert.Equal(t, u.Location, dto.Location)
	assert.Equal(t, u.AvatarURL, dto.AvatarURL)
	assert.Equal(t, u.Type, dto.Type)

	assert.NotEmpty(t, model.ToOrganizationsDto([]*model.Organization{u}))
}

func TestOrgMemberInfo(t *testing.T) {
	u := &model.OrgMemberInfo{
		MembershipID: 999,
		OrgID:        1000,
		OrgName:      "org_name",
		UserID:       1001,
		Name:         "name",
		Email:        "email",
		Role:         sql.NullString{String: "role", Valid: true},
		Source:       sql.NullString{String: "source", Valid: true},
	}

	assert.Equal(t, "role", u.GetRole())
	assert.Equal(t, "source", u.GetSource())

	dto := u.ToDto()
	assert.Equal(t, "999", dto.MembershipID)
	assert.Equal(t, "1000", dto.OrgID)
	assert.Equal(t, "1001", dto.UserID)
	assert.Equal(t, "org_name", dto.OrgName)
	assert.Equal(t, "name", dto.Name)
	assert.Equal(t, "email", dto.Email)
	assert.Equal(t, "role", dto.Role)
	assert.Equal(t, "source", dto.Source)
}

func TestOrgMembership(t *testing.T) {
	u := &model.OrgMembership{
		ID:      999,
		OrgID:   1000,
		OrgName: "org_name",
		UserID:  1001,
		Role:    sql.NullString{String: "role", Valid: true},
		Source:  sql.NullString{String: "source", Valid: true},
	}
	assert.Equal(t, "role", u.GetRole())
	assert.Equal(t, "source", u.GetSource())

	dto := u.ToDto()
	assert.Equal(t, "999", dto.ID)
	assert.Equal(t, "1000", dto.OrgID)
	assert.Equal(t, "1001", dto.UserID)
	assert.Equal(t, "org_name", dto.OrgName)
	assert.Equal(t, "role", dto.Role)
	assert.Equal(t, "source", dto.Source)
}

func TestRepository(t *testing.T) {

	u := &model.Repository{
		ID:         1000,
		OrgID:      2000,
		ExternalID: 1001,
		Name:       "n1",
		Email:      "e1",
		Company:    "c1",
		AvatarURL:  "https://github.com/me",
		Type:       "private",
	}
	dto := u.ToDto()
	assert.Equal(t, "1000", dto.ID)
	assert.Equal(t, "2000", dto.OrgID)
	assert.Equal(t, "1001", dto.ExternalID)
	assert.Equal(t, u.Type, dto.Type)
	assert.Equal(t, u.Name, dto.Name)
	assert.Equal(t, u.Email, dto.Email)
	assert.Equal(t, u.Company, dto.Company)
	assert.Equal(t, u.AvatarURL, dto.AvatarURL)

	u = &model.Repository{
		ID:         1000,
		OrgID:      2000,
		ExternalID: 0,
	}
	dto = u.ToDto()
	assert.Equal(t, "1000", dto.ID)
	assert.Equal(t, "2000", dto.OrgID)
	assert.Equal(t, "", dto.ExternalID)
}
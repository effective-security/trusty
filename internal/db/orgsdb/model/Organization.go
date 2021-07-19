package model

import (
	"strconv"
	"time"

	v1 "github.com/ekspand/trusty/api/v1"
)

// Organization represents an organization account.
type Organization struct {
	ID           uint64    `db:"id"`
	ExternalID   uint64    `db:"extern_id"`
	Provider     string    `db:"provider"`
	Login        string    `db:"login"`
	AvatarURL    string    `db:"avatar_url"`
	URL          string    `db:"html_url"`
	Name         string    `db:"name"`
	Email        string    `db:"email"`
	BillingEmail string    `db:"billing_email"`
	Company      string    `db:"company"`
	Location     string    `db:"location"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	Type         string    `db:"type"`
	Address      string    `db:"address"`
	Zip          string    `db:"zip"`
	State        string    `db:"state"`
	Country      string    `db:"country"`
	Phone        string    `db:"phone"`
}

// ToDto converts model to v1.Organization DTO
func (u *Organization) ToDto() *v1.Organization {
	user := &v1.Organization{
		ID:           strconv.FormatUint(u.ID, 10),
		ExternalID:   strconv.FormatUint(u.ExternalID, 10),
		Provider:     u.Provider,
		Login:        u.Login,
		Name:         u.Name,
		Email:        u.Email,
		BillingEmail: u.BillingEmail,
		Company:      u.Company,
		Location:     u.Location,
		AvatarURL:    u.AvatarURL,
		URL:          u.URL,
		Type:         u.Type,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
		Address:      u.Address,
		Zip:          u.Zip,
		State:        u.State,
		Country:      u.Country,
		Phone:        u.Phone,
	}

	return user
}

// ToOrganizationsDto returns Organizations
func ToOrganizationsDto(list []*Organization) []v1.Organization {
	res := make([]v1.Organization, len(list))
	for i, org := range list {
		res[i] = *org.ToDto()
	}
	return res
}

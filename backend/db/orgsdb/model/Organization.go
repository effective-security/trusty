package model

import (
	"strconv"
	"time"

	v1 "github.com/martinisecurity/trusty/api/v1"
)

// Organization represents an organization account.
type Organization struct {
	ID             uint64    `db:"id"`
	ExternalID     string    `db:"extern_id"`
	RegistrationID string    `db:"registration_id"`
	Provider       string    `db:"provider"`
	Login          string    `db:"login"`
	AvatarURL      string    `db:"avatar_url"`
	URL            string    `db:"html_url"`
	Name           string    `db:"name"`
	Email          string    `db:"email"`
	BillingEmail   string    `db:"billing_email"`
	Company        string    `db:"company"`
	Location       string    `db:"location"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
	Type           string    `db:"type"`
	Street         string    `db:"street_address"`
	City           string    `db:"city"`
	PostalCode     string    `db:"postal_code"`
	Region         string    `db:"region"`
	Country        string    `db:"country"`
	Phone          string    `db:"phone"`
	ApproverEmail  string    `db:"approver_email"`
	ApproverName   string    `db:"approver_name"`
	Status         string    `db:"status"`
	ExpiresAt      time.Time `db:"expires_at"`
}

// ToDto converts model to v1.Organization DTO
func (u *Organization) ToDto() *v1.Organization {
	org := &v1.Organization{
		ID:             strconv.FormatUint(u.ID, 10),
		ExternalID:     u.ExternalID,
		RegistrationID: u.RegistrationID,
		Provider:       u.Provider,
		Login:          u.Login,
		Name:           u.Name,
		Email:          u.Email,
		BillingEmail:   u.BillingEmail,
		Company:        u.Company,
		Location:       u.Location,
		AvatarURL:      u.AvatarURL,
		URL:            u.URL,
		Type:           u.Type,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		Street:         u.Street,
		City:           u.City,
		PostalCode:     u.PostalCode,
		Region:         u.Region,
		Country:        u.Country,
		Phone:          u.Phone,
		ApproverName:   u.ApproverName,
		ApproverEmail:  u.ApproverEmail,
		Status:         u.Status,
		ExpiresAt:      u.ExpiresAt,
	}

	return org
}

// ToOrganizationsDto returns Organizations
func ToOrganizationsDto(list []*Organization) []v1.Organization {
	res := make([]v1.Organization, len(list))
	for i, org := range list {
		res[i] = *org.ToDto()
	}
	return res
}

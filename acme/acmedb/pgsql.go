package acmedb

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ekspand/trusty/acme/model"
	"github.com/ekspand/trusty/internal/db"
	"github.com/juju/errors"
)

// SetRegistration registers account
func (p *SQLProvider) SetRegistration(ctx context.Context, reg *model.Registration) (*model.Registration, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = db.Validate(reg)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("external_id=%q, key_id=%s", reg.ExternalID, reg.KeyID)

	key, err := json.Marshal(reg.Key)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.Registration)
	var contact string
	err = p.db.QueryRowContext(ctx, `
			INSERT INTO registrations(id,external_id,key_id,key,contact,agreement,initial_ip,created_at,status)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
				ON CONFLICT (key_id)
			DO UPDATE
				SET external_id=$2,contact=$5,agreement=$6,initial_ip=$7,created_at=$8,status=$9
			RETURNING id,external_id,key_id,key,contact,agreement,initial_ip,created_at,status
			;`, id,
		reg.ExternalID,
		reg.KeyID,
		key,
		strings.Join(reg.Contact, ","),
		reg.Agreement,
		reg.InitialIP,
		reg.CreatedAt.UTC(),
		reg.Status,
	).Scan(&res.ID,
		&res.ExternalID,
		&res.KeyID,
		&key,
		&contact,
		&res.Agreement,
		&res.InitialIP,
		&res.CreatedAt,
		&res.Status,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()
	res.Key = reg.Key
	res.Contact = strings.Split(contact, ",")
	return res, nil
}

// GetRegistration returns account registration
func (p *SQLProvider) GetRegistration(ctx context.Context, id uint64) (*model.Registration, error) {
	res := new(model.Registration)
	var key string
	var contact string
	err := p.db.QueryRowContext(ctx, `
		SELECT
			id,external_id,key_id,key,contact,agreement,initial_ip,created_at,status
		FROM registrations
		WHERE id = $1
		;
		`, id).Scan(
		&res.ID,
		&res.ExternalID,
		&res.KeyID,
		&key,
		&contact,
		&res.Agreement,
		&res.InitialIP,
		&res.CreatedAt,
		&res.Status,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	res.CreatedAt = res.CreatedAt.UTC()

	err = json.Unmarshal([]byte(key), &res.Key)
	if err != nil {
		return nil, errors.Annotatef(err, "corrupted data")
	}
	res.Contact = strings.Split(contact, ",")

	return res, nil
}

// UpdateOrder updates Order
func (p *SQLProvider) UpdateOrder(ctx context.Context, order *model.Order) (*model.Order, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = db.Validate(order)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("reg_id=%d, dns=%v", order.RegistrationID, order.DNSNames)

	res := new(model.Order)
	var authorizations string
	var names string
	var jsError string

	if order.Error != nil {
		js, err := json.Marshal(order.Error)
		if err != nil {
			return nil, errors.Trace(err)
		}
		jsError = string(js)
	}

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO orders(id,reg_id,names_hash,created_at,status,expires_at,not_before,not_after,error,authorizations,cert_id,dns_names,binding_id,external_order_id)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
				ON CONFLICT (reg_id,names_hash)
			DO UPDATE
				SET created_at=$4,status=$5,expires_at=$6,not_before=$7,not_after=$8,error=$9,authorizations=$10,cert_id=$11,binding_id=$13,external_order_id=$14
			RETURNING id,reg_id,names_hash,created_at,status,expires_at,not_before,not_after,error,authorizations,cert_id,dns_names,binding_id,external_order_id
			;`, id,
		order.RegistrationID,
		order.NamesHash,
		order.CreatedAt.UTC(),
		order.Status,
		order.ExpiresAt.UTC(),
		order.NotBefore,
		order.NotAfter,
		jsError,
		strings.Join(order.Authorizations, ","),
		order.CertificateID,
		strings.Join(order.DNSNames, ","),
		order.ExternalBindingID,
		order.ExternalOrderID,
	).Scan(&res.ID,
		&res.RegistrationID,
		&res.NamesHash,
		&res.CreatedAt,
		&res.Status,
		&res.ExpiresAt,
		&res.NotBefore,
		&res.NotAfter,
		&jsError,
		&authorizations,
		&res.CertificateID,
		&names,
		&res.ExternalBindingID,
		&res.ExternalOrderID,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if jsError != "" {
		err = json.Unmarshal([]byte(jsError), &res.Error)
		if err != nil {
			return nil, errors.Annotatef(err, "corrupted data")
		}
	}
	res.DNSNames = strings.Split(names, ",")
	res.Authorizations = strings.Split(authorizations, ",")
	res.CreatedAt = res.CreatedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()

	return res, nil
}

// GetOrder returns Order by ID
func (p *SQLProvider) GetOrder(ctx context.Context, registrationID uint64, namesHash string) (*model.Order, error) {
	res := new(model.Order)

	var authorizations string
	var names string
	var jsError string

	err := p.db.QueryRowContext(ctx, `
		SELECT
			id,reg_id,names_hash,created_at,status,expires_at,not_before,not_after,error,authorizations,cert_id,dns_names,binding_id,external_order_id
		FROM orders
		WHERE reg_id = $1 AND names_hash = $2
		;
		`, registrationID, namesHash).Scan(
		&res.ID,
		&res.RegistrationID,
		&res.NamesHash,
		&res.CreatedAt,
		&res.Status,
		&res.ExpiresAt,
		&res.NotBefore,
		&res.NotAfter,
		&jsError,
		&authorizations,
		&res.CertificateID,
		&names,
		&res.ExternalBindingID,
		&res.ExternalOrderID,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if jsError != "" {
		err = json.Unmarshal([]byte(jsError), &res.Error)
		if err != nil {
			return nil, errors.Annotatef(err, "corrupted data")
		}
	}

	res.DNSNames = strings.Split(names, ",")
	res.Authorizations = strings.Split(authorizations, ",")
	res.CreatedAt = res.CreatedAt.UTC()
	res.ExpiresAt = res.ExpiresAt.UTC()

	return res, nil
}

// GetOrders returns all Orders for specified registration
func (p *SQLProvider) GetOrders(ctx context.Context, registrationID uint64) ([]*model.Order, error) {
	res, err := p.db.QueryContext(ctx, `
		SELECT
			id,reg_id,names_hash,created_at,status,expires_at,not_before,not_after,error,authorizations,cert_id,dns_names,binding_id,external_order_id
		FROM
			orders
		WHERE reg_id = $1
		;
		`, registrationID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Close()

	list := make([]*model.Order, 0, 100)

	for res.Next() {
		r := new(model.Order)
		var authorizations string
		var names string
		var jsError string

		err = res.Scan(
			&r.ID,
			&r.RegistrationID,
			&r.NamesHash,
			&r.CreatedAt,
			&r.Status,
			&r.ExpiresAt,
			&r.NotBefore,
			&r.NotAfter,
			&jsError,
			&authorizations,
			&r.CertificateID,
			&names,
			&r.ExternalBindingID,
			&r.ExternalOrderID,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}
		if jsError != "" {
			err = json.Unmarshal([]byte(jsError), &r.Error)
			if err != nil {
				return nil, errors.Annotatef(err, "corrupted data")
			}
		}
		r.DNSNames = strings.Split(names, ",")
		r.Authorizations = strings.Split(authorizations, ",")
		r.CreatedAt = r.CreatedAt.UTC()
		r.ExpiresAt = r.ExpiresAt.UTC()
		list = append(list, r)
	}

	return list, nil
}

// PutIssuedCertificate saves issued cert
func (p *SQLProvider) PutIssuedCertificate(ctx context.Context, cert *model.IssuedCertificate) (*model.IssuedCertificate, error) {
	id, err := p.NextID()
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = db.Validate(cert)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("reg_id=%d, order_id=%v", cert.RegistrationID, cert.OrderID)

	res := new(model.IssuedCertificate)

	err = p.db.QueryRowContext(ctx, `
			INSERT INTO acmecerts(id,reg_id,order_id,binding_id,external_id,pem)
				VALUES($1, $2, $3, $4, $5, $6)
				ON CONFLICT (reg_id,order_id)
			DO UPDATE
				SET binding_id=$4,external_id=$5,pem=$6
			RETURNING id,reg_id,order_id,binding_id,external_id,pem
			;`, id,
		cert.RegistrationID,
		cert.OrderID,
		cert.ExternalBindingID,
		cert.ExternalID,
		cert.Certificate,
	).Scan(&res.ID,
		&res.RegistrationID,
		&res.OrderID,
		&res.ExternalBindingID,
		&res.ExternalID,
		&res.Certificate,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}
	return res, nil
}

// GetIssuedCertificate returns IssuedCertificate by ID
func (p *SQLProvider) GetIssuedCertificate(ctx context.Context, certID uint64) (*model.IssuedCertificate, error) {
	res := new(model.IssuedCertificate)

	err := p.db.QueryRowContext(ctx, `
		SELECT
			id,reg_id,order_id,binding_id,external_id,pem
		FROM acmecerts
		WHERE id = $1
		;
		`, certID).Scan(
		&res.ID,
		&res.RegistrationID,
		&res.OrderID,
		&res.ExternalBindingID,
		&res.ExternalID,
		&res.Certificate,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	return res, nil
}

// InsertAuthorization will persist Authorization and all its Challenge objects
func (p *SQLProvider) InsertAuthorization(ctx context.Context, authz *model.Authorization) (*model.Authorization, error) {
	if authz.ID == 0 {
		return nil, errors.Errorf("authorization must have ID")
	}

	err := db.Validate(authz)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("id=%d, reg_id=%d", authz.ID, authz.RegistrationID)

	challenges, err := json.Marshal(authz.Challenges)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.Authorization)
	err = p.db.QueryRowContext(ctx, `
			INSERT INTO authorizations(id,reg_id,type,value,status,expires_at,challenges)
				VALUES($1, $2, $3, $4, $5, $6, $7)
			RETURNING id,reg_id,type,value,status,expires_at,challenges
			;`, authz.ID,
		authz.RegistrationID,
		authz.Identifier.Type,
		authz.Identifier.Value,
		authz.Status,
		authz.ExpiresAt.UTC(),
		challenges,
	).Scan(&res.ID,
		&res.RegistrationID,
		&res.Identifier.Type,
		&res.Identifier.Value,
		&res.Status,
		&res.ExpiresAt,
		&challenges)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = json.Unmarshal([]byte(challenges), &res.Challenges)
	if err != nil {
		return nil, errors.Annotatef(err, "corrupted data")
	}
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// UpdateAuthorization will update Authorization
func (p *SQLProvider) UpdateAuthorization(ctx context.Context, authz *model.Authorization) (*model.Authorization, error) {
	err := db.Validate(authz)
	if err != nil {
		return nil, errors.Trace(err)
	}

	logger.Debugf("id=%d, reg_id=%d", authz.ID, authz.RegistrationID)

	challenges, err := json.Marshal(authz.Challenges)
	if err != nil {
		return nil, errors.Trace(err)
	}

	res := new(model.Authorization)
	err = p.db.QueryRowContext(ctx, `
		UPDATE authorizations
		SET type=$2,value=$3,status=$4,expires_at=$5,challenges=$6
		WHERE id = $1
		RETURNING id,reg_id,type,value,status,expires_at,challenges
		;`,
		authz.ID,
		authz.Identifier.Type,
		authz.Identifier.Value,
		authz.Status,
		authz.ExpiresAt.UTC(),
		challenges,
	).Scan(
		&res.ID,
		&res.RegistrationID,
		&res.Identifier.Type,
		&res.Identifier.Value,
		&res.Status,
		&res.ExpiresAt,
		&challenges)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = json.Unmarshal([]byte(challenges), &res.Challenges)
	if err != nil {
		return nil, errors.Annotatef(err, "corrupted data")
	}
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// GetAuthorization returns Authorization by ID
func (p *SQLProvider) GetAuthorization(ctx context.Context, authzID uint64) (*model.Authorization, error) {
	res := new(model.Authorization)

	var challenges string
	err := p.db.QueryRowContext(ctx, `
		SELECT
			id,reg_id,type,value,status,expires_at,challenges
		FROM authorizations
		WHERE id = $1
		;
		`, authzID).Scan(
		&res.ID,
		&res.RegistrationID,
		&res.Identifier.Type,
		&res.Identifier.Value,
		&res.Status,
		&res.ExpiresAt,
		&challenges,
	)
	if err != nil {
		return nil, errors.Trace(err)
	}

	err = json.Unmarshal([]byte(challenges), &res.Challenges)
	if err != nil {
		return nil, errors.Annotatef(err, "corrupted data")
	}
	res.ExpiresAt = res.ExpiresAt.UTC()
	return res, nil
}

// GetAuthorizations returns all Authorizations for specified registration
func (p *SQLProvider) GetAuthorizations(ctx context.Context, regID uint64) ([]*model.Authorization, error) {
	res, err := p.db.QueryContext(ctx, `
		SELECT
			id,reg_id,type,value,status,expires_at,challenges
		FROM authorizations
		WHERE reg_id = $1
		;
		`, regID)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer res.Close()

	list := make([]*model.Authorization, 0, 100)

	for res.Next() {
		r := new(model.Authorization)
		var challenges string

		err = res.Scan(
			&r.ID,
			&r.RegistrationID,
			&r.Identifier.Type,
			&r.Identifier.Value,
			&r.Status,
			&r.ExpiresAt,
			&challenges,
		)
		if err != nil {
			return nil, errors.Trace(err)
		}
		err = json.Unmarshal([]byte(challenges), &r.Challenges)
		if err != nil {
			return nil, errors.Annotatef(err, "corrupted data")
		}

		r.ExpiresAt = r.ExpiresAt.UTC()
		list = append(list, r)
	}

	return list, nil
}
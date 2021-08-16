package pgsql

import (
	"context"

	"github.com/juju/errors"
)

// GetUsersCount returns number of users
func (p *Provider) GetUsersCount(ctx context.Context) (uint64, error) {
	count := uint64(0)
	err := p.db.QueryRowContext(ctx, `SELECT COUNT(id) FROM users;`).Scan(&count)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return count, nil
}

// GetOrgsCount returns number of organizations
func (p *Provider) GetOrgsCount(ctx context.Context) (uint64, error) {
	count := uint64(0)
	err := p.db.QueryRowContext(ctx, `SELECT COUNT(id) FROM orgs;`).Scan(&count)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return count, nil

}

package pgsql

import (
	"context"

	"github.com/pkg/errors"
)

// GetCertsCount returns number of certs
func (p *Provider) GetCertsCount(ctx context.Context) (uint64, error) {
	count := uint64(0)
	err := p.sql.QueryRowContext(ctx, `SELECT COUNT(id) FROM certificates;`).Scan(&count)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return count, nil
}

// GetRevokedCount returns number of revoked certs
func (p *Provider) GetRevokedCount(ctx context.Context) (uint64, error) {
	count := uint64(0)
	err := p.sql.QueryRowContext(ctx, `SELECT COUNT(id) FROM revoked;`).Scan(&count)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return count, nil
}

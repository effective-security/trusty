package pgsql

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// GetTableRowsCount returns number of rows
func (p *Provider) GetTableRowsCount(ctx context.Context, table string) (uint64, error) {
	count := uint64(0)
	err := p.sql.QueryRowContext(ctx,
		fmt.Sprintf("SELECT COUNT(id) FROM %s;", table)).
		Scan(&count)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	return count, nil
}

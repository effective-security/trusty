package cadb_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/effective-security/porto/pkg/flake"
	db "github.com/effective-security/trusty/backend/db/cadb"
	"github.com/effective-security/trusty/tests/testutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	provider db.Provider
	ctx      = context.Background()
)

const (
	projFolder = "../../../"
)

func TestMain(m *testing.M) {
	//xlog.SetGlobalLogLevel(xlog.TRACE)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.WithStack(err))
	}

	p, err := db.New(
		cfg.CaSQL.Driver,
		cfg.CaSQL.DataSource,
		cfg.CaSQL.MigrationsDir,
		0,
		flake.DefaultIDGenerator,
	)
	if err != nil {
		panic(err.Error())
	}
	defer p.Close()
	provider = p
	// Run the tests
	rc := m.Run()
	os.Exit(rc)
}
func Test_ListTables(t *testing.T) {
	expectedTables := []string{
		"'issuers'",
		"'cert_profiles'",
		"'nonces'",
		"'certificates'",
		"'revoked'",
		"'roots'",
		"'crls'",
	}
	require.NotNil(t, provider)
	require.NotNil(t, provider.DB())
	res, err := provider.DB().Query(fmt.Sprintf(`
	SELECT
		tablename
	FROM
		pg_catalog.pg_tables
	WHERE
		tablename IN (%s);
	`, strings.Join(expectedTables, ",")))
	require.NoError(t, err)
	defer res.Close()

	count := 0
	var table string
	for res.Next() {
		count++
		err = res.Scan(&table)
		require.NoError(t, err)
	}
	assert.Equal(t, len(expectedTables), count)
}

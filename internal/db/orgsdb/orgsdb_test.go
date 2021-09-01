package orgsdb_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/ekspand/trusty/internal/db/orgsdb"
	"github.com/ekspand/trusty/tests/testutils"
	"github.com/juju/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	provider orgsdb.Provider
	ctx      = context.Background()
)

const (
	projFolder = "../../../"
)

func TestMain(m *testing.M) {
	//xlog.SetGlobalLogLevel(xlog.TRACE)

	cfg, err := testutils.LoadConfig(projFolder, "UNIT_TEST")
	if err != nil {
		panic(errors.Trace(err))
	}

	p, err := orgsdb.New(
		cfg.OrgsSQL.Driver,
		cfg.OrgsSQL.DataSource,
		cfg.OrgsSQL.MigrationsDir,
		0,
		testutils.IDGenerator().NextID,
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
		"'users'",
		"'orgs'",
		"'orgmembers'",
		"'repos'",
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

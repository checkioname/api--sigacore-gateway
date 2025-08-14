package db

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

var (
	dsn         string
	testStore   Store
	testQueries *Queries
	teardown    func()
)

func TestMain(m *testing.M) {
	testQueries, dsn, teardown = StartDBWithTestContainer(nil)
	defer teardown()

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal(err)
	}

	testStore = NewStore(pool)
	code := m.Run()

	if pool != nil {
		pool.Close()
	}
	if teardown != nil {
		teardown()
	}
	os.Exit(code)
}

func clearTables(ctx context.Context, t *testing.T) {
	t.Helper()
	tables := []string{"transfers", "entries", "accounts"}

	query := "TRUNCATE TABLE " + strings.Join(tables, ", ") + " RESTART IDENTITY CASCADE;"
	_, err := testQueries.db.Exec(ctx, query)
	require.NoError(t, err)
}

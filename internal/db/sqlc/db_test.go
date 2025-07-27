package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestContainer(t *testing.T) {
	_, dsn, teardown := StartDBWithTestContainer(t)
	defer teardown()

	require.NotEmpty(t, dsn)
}

func StartDBWithTestContainer(t *testing.T) (*Queries, string, func()) {
	fmt.Println("Iniciando DB testContainers")
	if t != nil {
		t.Helper()
	}

	dsn, terminateContainer, err := startTestContainer()
	if err != nil {
		log.Printf("start test container: %v", err)
	}

	db, teardown, err := initializeDatabase(dsn)
	if err != nil {
		log.Printf("initialize database: %v", err)
		return nil, "", nil
	}

	return db, dsn, func() {
		teardown()
		terminateContainer()
	}
}

func startTestContainer() (string, func(), error) {
	_ = setColimaEnvVars()
	ctx := context.Background()

	// Disable ryuk until it is fixed in a future release of testcontainers-go
	if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
		panic(err)
	}
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		return "", nil, fmt.Errorf("starting PostgreSQL container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("getting PostrgreSQL container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return "", nil, fmt.Errorf("getting PostrgreSQL container port: %w", err)
	}

	host = fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		"postgres",
		"password",
		host,
		port.Int(),
		"test",
	)
	terminate := func() {
		if err := container.Terminate(context.Background()); err != nil {
			panic(err)
		}
	}
	return host, terminate, nil
}

func setColimaEnvVars() error {
	err := os.Setenv("TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE", "/var/run/docker.sock")
	if err != nil {
		return fmt.Errorf("setting TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE: %w", err)
	}
	err = os.Setenv("DOCKER_HOST", fmt.Sprintf("unix://%s/.colima/docker.sock", os.Getenv("HOME")))
	if err != nil {
		return fmt.Errorf("setting DOCKER_HOST: %w", err)
	}
	return nil
}

func initializeDatabase(dsn string) (*Queries, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		cancel()
		return nil, nil, fmt.Errorf("initializeDatabase: %v", err)
	}

	migrationPath := "file://../migration"
	m, err := migrate.New(migrationPath, dsn)
	if err != nil {
		conn.Close(ctx)
		cancel()
		return nil, nil, fmt.Errorf("initializeDatabase: %v", err)
	}

	if err := m.Up(); err != nil {
		conn.Close(ctx)
		cancel()
		return nil, nil, fmt.Errorf("initializeDatabase: %v", err)
	}

	db := New(conn)

	teardown := func() {
		_ = conn.Close(ctx)
		cancel()
	}

	return db, teardown, nil
}

//func RandomCreateAccount(balance int64) *Account {
//	var b int64
//	if balance > 0 {
//		b = balance
//	} else {
//		b = balance
//	}
//
//	return &Account{
//		util.RandomInt(0, 200),
//		util.RandomOwner(),
//		b,
//		util.RandomCurrency(),
//		time.Now(),
//	}

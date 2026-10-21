// Package testdb gives integration tests a real, migrated Postgres.
//
// One container starts per test binary. It holds a migrated template database;
// every New call clones it into a fresh database, so tests are isolated and can
// run in parallel without cleaning up after each other.
package testdb

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/postgres"
)

const (
	image    = "postgres:18.6-alpine"
	template = "spilno_template"
)

var (
	once    sync.Once
	baseURL string
	initErr error
	counter atomic.Int64
)

// New returns a pool connected to a fresh, fully migrated database. The
// database is dropped when the test ends.
func New(t testing.TB) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	once.Do(func() { baseURL, initErr = start(ctx) })
	if initErr != nil {
		t.Fatalf("testdb: %v", initErr)
	}

	name := fmt.Sprintf("test_%d", counter.Add(1))
	admin, err := pgx.Connect(ctx, baseURL)
	if err != nil {
		t.Fatalf("testdb: connect admin: %v", err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s TEMPLATE %s", name, template)); err != nil {
		t.Fatalf("testdb: create database: %v", err)
	}

	pool, err := pgxpool.New(ctx, withDatabase(baseURL, name))
	if err != nil {
		t.Fatalf("testdb: open pool: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		admin, err := pgx.Connect(ctx, baseURL)
		if err != nil {
			t.Errorf("testdb: connect admin for cleanup: %v", err)
			return
		}
		defer admin.Close(ctx)
		if _, err := admin.Exec(ctx, fmt.Sprintf("DROP DATABASE %s WITH (FORCE)", name)); err != nil {
			t.Errorf("testdb: drop database: %v", err)
		}
	})
	return pool
}

// start boots the container and migrates the template database. The container
// is reaped by testcontainers when the test binary exits.
func start(ctx context.Context) (string, error) {
	c, err := tcpostgres.Run(ctx, image,
		tcpostgres.WithDatabase("postgres"),
		tcpostgres.WithUsername("spilno"),
		tcpostgres.WithPassword("spilno"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", fmt.Errorf("start container: %w", err)
	}
	base, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return "", fmt.Errorf("connection string: %w", err)
	}

	admin, err := pgx.Connect(ctx, base)
	if err != nil {
		return "", fmt.Errorf("connect admin: %w", err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+template); err != nil {
		return "", fmt.Errorf("create template: %w", err)
	}

	pool, err := pgxpool.New(ctx, withDatabase(base, template))
	if err != nil {
		return "", fmt.Errorf("open template pool: %w", err)
	}
	// The template must have no open connections when it is cloned.
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool); err != nil {
		return "", err
	}
	return base, nil
}

func withDatabase(raw, db string) string {
	u, err := url.Parse(raw)
	if err != nil {
		panic(fmt.Sprintf("testdb: bad connection string: %v", err))
	}
	u.Path = "/" + db
	return u.String()
}

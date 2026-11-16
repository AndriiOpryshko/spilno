package postgres_test

import (
	"context"
	"strings"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// AC-04d, AC-07 (ADR-0003): the API refuses to serve as a role that bypasses row-level security.
func TestRequireRLSEnforcedRefusesOwnerRole(t *testing.T) {
	ctx := context.Background()
	owner := testdb.New(t) // connects as the owner (superuser)
	err := postgres.RequireRLSEnforced(ctx, owner)
	if err == nil {
		t.Fatal("RequireRLSEnforced(owner pool) = nil, want an error naming the bypass")
	}
	if !strings.Contains(err.Error(), "DATABASE_URL") || !strings.Contains(strings.ToLower(err.Error()), "row-level security") {
		t.Errorf("error %q should name DATABASE_URL and row-level security", err)
	}
}

func TestRequireRLSEnforcedAcceptsAppRole(t *testing.T) {
	app := testdb.NewApp(t)
	if err := postgres.RequireRLSEnforced(context.Background(), app); err != nil {
		t.Fatalf("RequireRLSEnforced(app pool) = %v, want nil", err)
	}
}

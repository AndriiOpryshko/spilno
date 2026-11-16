package config

import "testing"

// AC-04d: DATABASE_URL has no owner default; MIGRATIONS_DATABASE_URL does not borrow it either.
func TestDatabaseURLHasNoDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("MIGRATIONS_DATABASE_URL", "")
	c := Load()
	if c.DatabaseURL != "" {
		t.Errorf("DatabaseURL = %q, want empty without DATABASE_URL", c.DatabaseURL)
	}
	if c.MigrationsDatabaseURL != "" {
		t.Errorf("MigrationsDatabaseURL = %q, want empty without MIGRATIONS_DATABASE_URL", c.MigrationsDatabaseURL)
	}
}

package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// AC-01: a fresh contract carries no hryvnia equivalent; both new columns read as NULL.
func TestContractsUAHEquivalentColumnsStartNull(t *testing.T) {
	_, owner, assoc, _ := testdb.NewTwoAssociationsWithOwner(t)
	ctx := context.Background()

	var id string
	err := owner.QueryRow(ctx, `INSERT INTO contracts (counterparty, amount_cents, currency, signed_at, association_id)
		VALUES ('Test Counterparty', 10000, 'USD', DATE '2026-11-20', $1) RETURNING id::text`, assoc).Scan(&id)
	if err != nil {
		t.Fatalf("insert contract: %v", err)
	}

	var amount, rate pgtype.Int8
	err = owner.QueryRow(ctx, `SELECT uah_amount_kopiyky, uah_rate_micros FROM contracts WHERE id = $1`, id).Scan(&amount, &rate)
	if err != nil {
		t.Fatalf("read uah equivalent columns: %v", err)
	}
	if amount.Valid || rate.Valid {
		t.Fatalf("uah_amount_kopiyky valid=%v, uah_rate_micros valid=%v; want both NULL", amount.Valid, rate.Valid)
	}
}

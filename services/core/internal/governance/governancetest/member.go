// Package governancetest holds fixtures for tests that need a roster.
package governancetest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Member puts email on the roster of the association with a term covering today.
func Member(t testing.TB, tx db.DBTX, association uuid.UUID, email, position string) {
	t.Helper()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	MemberTerm(t, tx, association, email, position, today.AddDate(-1, 0, 0), time.Time{})
}

// MemberTerm is Member with an explicit term; a zero end leaves the term open.
func MemberTerm(t testing.TB, tx db.DBTX, association uuid.UUID, email, position string, start, end time.Time) {
	t.Helper()
	ctx := context.Background()
	memberID, revID := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	if _, err := tx.Exec(ctx, `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, memberID, association); err != nil {
		t.Fatalf("governancetest: insert board member: %v", err)
	}
	var endArg any
	if !end.IsZero() {
		endArg = end
	}
	if _, err := tx.Exec(ctx, `INSERT INTO board_member_revisions
		(id, association_id, board_member_id, full_name, sign_in_email, term_start, term_end, position)
		VALUES ($1, $2, $3, 'Test User', $4, $5, $6, $7)`,
		revID, association, memberID, email, start, endArg, position); err != nil {
		t.Fatalf("governancetest: insert revision: %v", err)
	}
}

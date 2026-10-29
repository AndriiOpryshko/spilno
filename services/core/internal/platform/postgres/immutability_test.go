package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const (
	assoc    = "00000000-0000-7000-8000-000000000001"
	member   = "00000000-0000-7000-8000-0000000000a1"
	meeting  = "00000000-0000-7000-8000-0000000000b1"
	protocol = "00000000-0000-7000-8000-0000000000c1"
	snapshot = "00000000-0000-7000-8000-0000000000d1"
)

// ADR-0007: a signed snapshot is append-only and no protocol is ever deleted, whatever the app does.
func TestProtocolRecordsAreImmutable(t *testing.T) {
	ctx := context.Background()
	pool := testdb.New(t)

	seed := []string{
		`INSERT INTO board_members (id, association_id) VALUES ('` + member + `', '` + assoc + `')`,
		`INSERT INTO meetings (id, title, held_at, association_id) VALUES ('` + meeting + `', 'Test meeting', now(), '` + assoc + `')`,
		`INSERT INTO protocols (id, association_id, meeting_id, state, revision, created_by_member_id)
			VALUES ('` + protocol + `', '` + assoc + `', '` + meeting + `', 'locked', 1, '` + member + `')`,
		`INSERT INTO protocol_snapshots (id, association_id, protocol_id, kind, number, meeting_date,
			schema_version, snapshot, fingerprint, created_by_member_id)
			VALUES ('` + snapshot + `', '` + assoc + `', '` + protocol + `', 'locked', 1, current_date,
			1, '{}', '\x00', '` + member + `')`,
	}
	for _, q := range seed {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("seed: %v\n%s", err, q)
		}
	}

	for name, q := range map[string]string{
		"update snapshot":   `UPDATE protocol_snapshots SET number = 2 WHERE id = '` + snapshot + `'`,
		"delete snapshot":   `DELETE FROM protocol_snapshots WHERE id = '` + snapshot + `'`,
		"truncate snapshot": `TRUNCATE protocol_snapshots CASCADE`,
		"truncate protocol": `TRUNCATE protocols CASCADE`,
		"delete protocol":   `DELETE FROM protocols WHERE id = '` + protocol + `'`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := pool.Exec(ctx, q)
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "23001" {
				t.Fatalf("want restrict_violation (23001), got %v", err)
			}
		})
	}
}

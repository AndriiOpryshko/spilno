package governance_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// querySQL returns the named statement of a queries file with every sqlc.arg/sqlc.narg turned into the
// $n of its first appearance; names lists them in $n order.
func querySQL(t *testing.T, file, name string) (sql string, names []string) {
	t.Helper()
	raw, err := os.ReadFile("adapters/postgres/queries/" + file)
	if err != nil {
		t.Fatal(err)
	}
	_, rest, ok := strings.Cut(string(raw), "-- name: "+name+" ")
	if !ok {
		t.Fatalf("%s has no %s query", file, name)
	}
	_, rest, _ = strings.Cut(rest, "\n")
	body, _, _ := strings.Cut(rest, "-- name:")
	pos := map[string]int{}
	body = regexp.MustCompile(`sqlc\.n?arg\((\w+)\)`).ReplaceAllStringFunc(body, func(m string) string {
		arg := regexp.MustCompile(`\((\w+)\)`).FindStringSubmatch(m)[1]
		if _, seen := pos[arg]; !seen {
			names = append(names, arg)
			pos[arg] = len(names)
		}
		return fmt.Sprintf("$%d", pos[arg])
	})
	return strings.TrimSuffix(strings.TrimSpace(body), ";"), names
}

// T63 (board-protocol T31): on a registry of many decisions the search is served by the newest-meeting
// index of registry_entries, and the withdrawn numbers by the (association, kind, number) index of the
// snapshots, never by a sequential scan of either table.
func TestRegistryQueriesUseTheirIndexes(t *testing.T) {
	pool, owner, as, _ := testdb.NewTwoAssociationsWithOwner(t)
	ctx := context.Background()
	a := uuid.MustParse(as)

	member := uuid.Must(uuid.NewV7())
	if _, err := owner.Exec(ctx, `INSERT INTO board_members (id, association_id) VALUES ($1, $2)`, member, a); err != nil {
		t.Fatal(err)
	}
	day := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	err := db.InAssociation(ctx, pool, a, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		for n := 1; n <= 400; n++ {
			d := day.AddDate(0, 0, n)
			if n%4 == 0 {
				governancetest.SeedWithdrawn(t, tx, a, n, d, member)
				continue
			}
			governancetest.SeedSigned(t, tx, a, governancetest.SignedSpec{Number: n, MeetingDate: d, Subject: "Про тест", LockedBy: member})
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, table := range []string{"registry_entries", "protocol_snapshots"} {
		if _, err := owner.Exec(ctx, "ANALYZE "+table); err != nil {
			t.Fatalf("analyze %s: %v", table, err)
		}
	}

	for _, tc := range []struct {
		query, index, seqScan string
		args                  map[string]any
	}{
		{"SearchRegistryDecisions", "idx_registry_entries_association_meeting_date", "Seq Scan on registry_entries", map[string]any{
			"association_id": a, "row_limit": int32(21),
		}},
		{"WithdrawnNumbersBetween", "idx_protocol_snapshots_association_kind_number", "Seq Scan on protocol_snapshots", map[string]any{
			"association_id": a, "lo": int32(100), "hi": int32(140),
		}},
	} {
		t.Run(tc.query, func(t *testing.T) {
			sql, names := querySQL(t, "registry.sql", tc.query)
			args := make([]any, len(names))
			for i, n := range names {
				args[i] = tc.args[n] // a filter left out stays NULL
			}
			var plan []string
			err := db.InAssociation(ctx, pool, a, "seed@example.test", func(ctx context.Context) error {
				tx, err := db.From(ctx)
				if err != nil {
					return err
				}
				rows, err := tx.Query(ctx, "EXPLAIN "+sql, args...)
				if err != nil {
					return err
				}
				defer rows.Close()
				for rows.Next() {
					var line string
					if err := rows.Scan(&line); err != nil {
						return err
					}
					plan = append(plan, line)
				}
				return rows.Err()
			})
			if err != nil {
				t.Fatal(err)
			}
			text := strings.Join(plan, "\n")
			if !strings.Contains(text, tc.index) {
				t.Errorf("plan does not use %s:\n%s", tc.index, text)
			}
			if strings.Contains(text, tc.seqScan) {
				t.Errorf("plan has %q:\n%s", tc.seqScan, text)
			}
		})
	}
}

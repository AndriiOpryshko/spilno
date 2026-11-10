package postgres_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
	"github.com/AndriiOpryshko/spilno/services/core/migrations"
)

// liveRelations are the tables and indexes the migrations must create; every Down must drop
// them again.
var liveRelations = []string{
	"contracts", "meetings", "associations",
	"accounts", "sessions", "google_connections",
	"board_members", "board_member_revisions", "association_settings", "protocol_registers",
	"idx_meetings_chair_board_member_id",
	"files", "cases", "case_documents", "protocols", "protocol_transitions",
	"agenda_items", "agenda_item_versions", "assistant_turns", "assistant_turn_documents",
	"decision_slots", "slot_values", "slot_confirmations",
	"poll_rounds", "poll_votes",
	"print_copies", "meeting_facts", "meeting_attendees", "meeting_item_votes",
	"protocol_snapshots", "registry_entries", "integrity_runs", "integrity_findings",
	"application_extracts", "applications", "application_filings", "application_checks",
	"application_decisions", "association_state_changes",
	"uq_applications_account_edrpou_pending", "idx_association_state_changes_association_created",
	"uq_associations_edrpou", "idx_associations_confirmed_by_decision",
	"idx_board_member_revisions_sign_in_email",
}

// liveColumns are columns the migrations add to tables that existed before them.
var liveColumns = [][2]string{
	{"meetings", "chair_board_member_id"},
	{"associations", "edrpou"}, {"associations", "address"}, {"associations", "state"},
	{"associations", "confirmed_by_decision_id"},
	{"board_member_revisions", "position"}, {"board_member_revisions", "decision_id"},
	{"contracts", "association_id"},
}

func TestMigrationsRoundtrip(t *testing.T) {
	ctx := context.Background()
	pool := testdb.New(t)
	db := stdlib.OpenDBFromPool(pool)
	t.Cleanup(func() { _ = db.Close() })
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		t.Fatalf("goose provider: %v", err)
	}

	assertSchema(t, pool)

	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatalf("down to 0: %v", err)
	}
	var left []string
	rows, err := pool.Query(ctx, `SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name <> 'goose_db_version' ORDER BY 1`)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		left = append(left, name)
	}
	if len(left) > 0 {
		t.Errorf("tables left after down to 0: %v", left)
	}

	if _, err := provider.Up(ctx); err != nil {
		t.Fatalf("re-apply: %v", err)
	}
	assertSchema(t, pool)
}

func assertSchema(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		t.Fatalf("river migrator: %v", err)
	}
	res, err := migrator.Validate(context.Background(), nil)
	if err != nil {
		t.Fatalf("river validate: %v", err)
	}
	if !res.OK {
		t.Errorf("river schema is not current: %v", res.Messages)
	}
	for _, name := range liveRelations {
		var found *string
		if err := pool.QueryRow(context.Background(), `SELECT to_regclass($1)::text`, "public."+name).Scan(&found); err != nil {
			t.Fatalf("to_regclass %s: %v", name, err)
		}
		if found == nil {
			t.Errorf("relation %s does not exist", name)
		}
	}
	assertAppRole(t, pool)
	for _, c := range liveColumns {
		var n int
		if err := pool.QueryRow(context.Background(), `SELECT count(*) FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`, c[0], c[1]).Scan(&n); err != nil {
			t.Fatalf("column %s.%s: %v", c[0], c[1], err)
		}
		if n != 1 {
			t.Errorf("column %s.%s does not exist", c[0], c[1])
		}
	}
}

// assertAppRole checks the spilno_app role exists without BYPASSRLS and that the audit tables are
// insert-only for it.
func assertAppRole(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	var bypass bool
	if err := pool.QueryRow(ctx, `SELECT rolbypassrls FROM pg_roles WHERE rolname = 'spilno_app'`).Scan(&bypass); err != nil {
		t.Errorf("role spilno_app: %v", err)
		return
	}
	if bypass {
		t.Errorf("role spilno_app must not bypass row-level security")
	}
	for _, table := range []string{"application_extracts", "application_filings", "application_checks",
		"application_decisions", "association_state_changes"} {
		var ins, upd, del bool
		if err := pool.QueryRow(ctx, `SELECT has_table_privilege('spilno_app', $1, 'INSERT'),
			has_table_privilege('spilno_app', $1, 'UPDATE'), has_table_privilege('spilno_app', $1, 'DELETE')`,
			"public."+table).Scan(&ins, &upd, &del); err != nil {
			t.Errorf("privileges on %s: %v", table, err)
			continue
		}
		if !ins || upd || del {
			t.Errorf("spilno_app on %s: insert=%v update=%v delete=%v, want insert only", table, ins, upd, del)
		}
	}
}

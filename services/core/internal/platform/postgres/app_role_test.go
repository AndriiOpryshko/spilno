package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// AC-06, AC-15: the history of a заявка is insert-only, enforced by privileges of the app role
// (ADR-0005), not by a trigger. The pool of NewTwoAssociations connects as a login role in
// spilno_app, never as the owner.
func TestAppRoleCannotRewriteInsertOnlyTables(t *testing.T) {
	ctx := context.Background()
	pool, assocA, assocB := testdb.NewTwoAssociations(t)

	var super, bypass, member bool
	if err := pool.QueryRow(ctx, `SELECT rolsuper, rolbypassrls, pg_has_role(current_user, 'spilno_app', 'MEMBER')
		FROM pg_roles WHERE rolname = current_user`).Scan(&super, &bypass, &member); err != nil {
		t.Fatalf("inspect role: %v", err)
	}
	if super || bypass || !member {
		t.Fatalf("pool role: superuser=%v bypassrls=%v member of spilno_app=%v, want false false true", super, bypass, member)
	}

	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM associations WHERE id IN ($1, $2)`, assocA, assocB).Scan(&n); err != nil || n != 2 {
		t.Fatalf("seeded associations A and B visible: got %d, err %v, want 2", n, err)
	}

	const (
		account  = "00000000-0000-7000-8000-0000000000e1"
		appl     = "00000000-0000-7000-8000-0000000000e2"
		extract  = "00000000-0000-7000-8000-0000000000e3"
		filing   = "00000000-0000-7000-8000-0000000000e4"
		check    = "00000000-0000-7000-8000-0000000000e5"
		decision = "00000000-0000-7000-8000-0000000000e6"
		change   = "00000000-0000-7000-8000-0000000000e7"
	)
	inserts := []struct{ table, q string }{
		{"accounts", `INSERT INTO accounts (id, oidc_issuer, oidc_subject, email)
			VALUES ('` + account + `', 'https://issuer.example.test', 'sub-1', 'user-1@example.test')`},
		{"applications", `INSERT INTO applications (id, kind, account_id, edrpou, state)
			VALUES ('` + appl + `', 'new_association', '` + account + `', '12345678', 'pending')`},
		{"application_extracts", `INSERT INTO application_extracts (id, account_id, content, sha256, size_bytes, media_type, original_name)
			VALUES ('` + extract + `', '` + account + `', '\x01', '\x02', 1, 'application/pdf', 'extract.pdf')`},
		{"application_filings", `INSERT INTO application_filings (id, application_id, extract_id, association_name, association_address,
			applicant_full_name, position, term_start)
			VALUES ('` + filing + `', '` + appl + `', '` + extract + `', 'Test OSBB', 'Test Street 1', 'Test User', 'head', current_date)`},
		{"application_checks", `INSERT INTO application_checks (id, filing_id, signature_valid, signer_full_name, signed_on, head_full_name,
			extract_edrpou, extract_dated_on, checked_by_email)
			VALUES ('` + check + `', '` + filing + `', true, 'Test User', current_date, 'Test User', '12345678', current_date, 'operator@example.test')`},
		{"application_decisions", `INSERT INTO application_decisions (id, application_id, filing_id, check_id, outcome, decided_by_email)
			VALUES ('` + decision + `', '` + appl + `', '` + filing + `', '` + check + `', 'rejected', 'operator@example.test')`},
		{"association_state_changes", `INSERT INTO association_state_changes (id, association_id, to_state, changed_by_email)
			VALUES ('` + change + `', '` + assocA + `', 'active', 'operator@example.test')`},
	}
	for _, in := range inserts {
		if _, err := pool.Exec(ctx, in.q); err != nil {
			t.Fatalf("app role INSERT into %s must work: %v", in.table, err)
		}
	}

	for _, tbl := range []string{"application_extracts", "application_filings", "application_checks", "application_decisions", "association_state_changes"} {
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM `+tbl).Scan(&n); err != nil || n != 1 {
			t.Fatalf("app role SELECT from %s: count %d, err %v, want 1", tbl, n, err)
		}
		for verb, q := range map[string]string{
			"update": `UPDATE ` + tbl + ` SET created_at = now()`,
			"delete": `DELETE FROM ` + tbl,
		} {
			t.Run(verb+" "+tbl, func(t *testing.T) {
				_, err := pool.Exec(ctx, q)
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) || pgErr.Code != "42501" {
					t.Fatalf("want insufficient_privilege (42501), got %v", err)
				}
			})
		}
	}
}

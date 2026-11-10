package identitytest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Pending is the rows PendingApplication seeds.
type Pending struct {
	AccountID, ApplicationID, FilingID, ExtractID uuid.UUID
}

// PendingApplication seeds a person, a pending connection application for edrpou
// and its filing with an extract, all stamped filedAt.
func PendingApplication(t testing.TB, tx db.DBTX, edrpou string, filedAt time.Time) Pending {
	t.Helper()
	ctx := context.Background()
	p := Pending{uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())}
	exec := func(what, q string, args ...any) {
		if _, err := tx.Exec(ctx, q, args...); err != nil {
			t.Fatalf("identitytest: insert %s: %v", what, err)
		}
	}
	exec("account", `INSERT INTO accounts (id, oidc_issuer, oidc_subject, email)
		VALUES ($1, 'https://issuer.example.test', $2, $3)`,
		p.AccountID, "sub-"+p.AccountID.String(), "user-"+p.AccountID.String()+"@example.test")
	content := []byte("test extract")
	exec("extract", `INSERT INTO application_extracts
		(id, account_id, content, sha256, size_bytes, media_type, original_name, created_at)
		VALUES ($1, $2, $3, $4, $5, 'application/pdf', 'extract.pdf', $6)`,
		p.ExtractID, p.AccountID, content, []byte("test-sha256"), len(content), filedAt)
	exec("application", `INSERT INTO applications (id, kind, account_id, edrpou, state, created_at)
		VALUES ($1, 'connection', $2, $3, 'pending', $4)`, p.ApplicationID, p.AccountID, edrpou, filedAt)
	exec("filing", `INSERT INTO application_filings
		(id, application_id, extract_id, association_name, association_address, applicant_full_name, position, term_start, created_at)
		VALUES ($1, $2, $3, 'Test Association', 'Test Address 1', 'Test User', 'head', ($4::timestamptz)::date, $4)`,
		p.FilingID, p.ApplicationID, p.ExtractID, filedAt)
	return p
}

// Check records a check of the filing's extract and returns its id.
func Check(t testing.TB, tx db.DBTX, filingID uuid.UUID, signer, head string, extractDated time.Time) uuid.UUID {
	t.Helper()
	id := uuid.Must(uuid.NewV7())
	_, err := tx.Exec(context.Background(), `INSERT INTO application_checks
		(id, filing_id, signature_valid, signer_full_name, signed_on, head_full_name, extract_edrpou, extract_dated_on, checked_by_email)
		SELECT $1, f.id, true, $3, $5, $4, a.edrpou, $5, 'operator@example.test'
		FROM application_filings f JOIN applications a ON a.id = f.application_id WHERE f.id = $2`,
		id, filingID, signer, head, extractDated)
	if err != nil {
		t.Fatalf("identitytest: insert check: %v", err)
	}
	return id
}

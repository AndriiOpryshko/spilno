package identitytest_test

import (
	"context"
	"testing"
	"time"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/identitytest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

func TestPendingApplicationSeedsAFilingWithItsExtract(t *testing.T) {
	pool := testdb.New(t)
	ctx := context.Background()
	filed := time.Now().UTC().AddDate(0, 0, -2)

	p := identitytest.PendingApplication(t, pool, "12345678", filed)

	var state, edrpou, kind string
	var extractID, filingExtract any
	var size int
	err := pool.QueryRow(ctx, `SELECT a.state, a.edrpou, a.kind, f.extract_id, e.id, e.size_bytes
		FROM applications a
		JOIN application_filings f ON f.application_id = a.id
		JOIN application_extracts e ON e.id = f.extract_id
		WHERE a.id = $1 AND f.id = $2`, p.ApplicationID, p.FilingID).
		Scan(&state, &edrpou, &kind, &filingExtract, &extractID, &size)
	if err != nil {
		t.Fatalf("read seeded application back: %v", err)
	}
	if state != "pending" || edrpou != "12345678" {
		t.Errorf("application = %s/%s, want pending/12345678", state, edrpou)
	}
	if size <= 0 || filingExtract != extractID {
		t.Errorf("extract size %d, filing extract %v vs %v", size, filingExtract, extractID)
	}
	var age int
	if err := pool.QueryRow(ctx, `SELECT (now()::date - created_at::date) FROM applications WHERE id = $1`, p.ApplicationID).Scan(&age); err != nil {
		t.Fatal(err)
	}
	if age != 2 {
		t.Errorf("application created %d days ago, want 2", age)
	}
}

func TestCheckRecordsARowOnTheFiling(t *testing.T) {
	pool := testdb.New(t)
	ctx := context.Background()
	p := identitytest.PendingApplication(t, pool, "12345678", time.Now())
	dated := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	id := identitytest.Check(t, pool, p.FilingID, "Test Signer", "Test Head", dated)

	var signer, head string
	var on time.Time
	err := pool.QueryRow(ctx, `SELECT signer_full_name, head_full_name, extract_dated_on
		FROM application_checks WHERE id = $1 AND filing_id = $2`, id, p.FilingID).Scan(&signer, &head, &on)
	if err != nil {
		t.Fatalf("read seeded check back: %v", err)
	}
	if signer != "Test Signer" || head != "Test Head" || !on.Equal(dated) {
		t.Errorf("check = %q/%q/%v", signer, head, on)
	}
}

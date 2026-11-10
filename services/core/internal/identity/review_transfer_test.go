package identity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

const reviewT28AssocA = "00000000-0000-7000-8000-0000000000a1"

// reviewT28Transfer files a запит на передачу for association A from the card and returns its id.
func reviewT28Transfer(t *testing.T, e *reviewEnv, as string) string {
	t.Helper()
	reviewT28SetCode(t, e)
	return e.fileAs(t, as, transferFields("55667788"), fileAppPDF(256))
}

// reviewT28Storage returns the raw storage_account_email of the review answer.
func reviewT28Storage(t *testing.T, e *reviewEnv, id string) (json.RawMessage, bool) {
	t.Helper()
	status, _, body := e.do(t, "operator", http.MethodGet, "/identity/operator/applications/"+id, "")
	if status != http.StatusOK {
		t.Fatalf("getApplicationForReview = %d: %s", status, body)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	raw, ok := m["storage_account_email"]
	return raw, ok
}

// reviewT28SetCode gives association A its code, as the owner of the row would.
func reviewT28SetCode(t *testing.T, e *reviewEnv) {
	t.Helper()
	aid := uuid.MustParse(reviewT28AssocA)
	err := db.InAssociation(context.Background(), e.pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE associations SET edrpou = '55667788' WHERE id = $1`, aid)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func reviewT28Connect(t *testing.T, e *reviewEnv, email string) {
	t.Helper()
	aid := uuid.MustParse(reviewT28AssocA)
	err := db.InAssociation(context.Background(), e.pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		var acc uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT id FROM accounts LIMIT 1`).Scan(&acc); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO google_connections
			(id, association_id, google_account_email, refresh_token_ciphertext, refresh_token_nonce, encryption_key_id, scopes, authorized_by_account_id)
			VALUES ($1, $2, $3, '\x00', '\x00', 'k1', 'drive.file', $4)`, uuid.New(), aid, email, acc)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

// AC-19b: two pending запити for one association each list the other.
func TestReview_TransferListsTheOtherPendingRequestsOfTheAssociation(t *testing.T) {
	e := newReviewEnv(t)
	one := reviewT28Transfer(t, e, "first")
	two := reviewT28Transfer(t, e, "second")
	if got := e.review(t, one).Competing; len(got) != 1 || got[0] != two {
		t.Fatalf("competing of first transfer = %v, want [%s]", got, two)
	}
	if got := e.review(t, two).Competing; len(got) != 1 || got[0] != one {
		t.Fatalf("competing of second transfer = %v, want [%s]", got, one)
	}
}

// AC-28b: the review carries the address of the account holding the files; null when none is
// connected and for a заявка.
func TestReview_TransferCarriesTheStorageAccount(t *testing.T) {
	e := newReviewEnv(t)
	id := reviewT28Transfer(t, e, "first")
	raw, ok := reviewT28Storage(t, e, id)
	if !ok || string(raw) != "null" {
		t.Fatalf("storage_account_email without storage = %s (present %v), want null", raw, ok)
	}
	reviewT28Connect(t, e, "files@example.test")
	raw, _ = reviewT28Storage(t, e, id)
	if string(raw) != `"files@example.test"` {
		t.Fatalf("storage_account_email = %s, want \"files@example.test\"", raw)
	}
	app := e.fileAs(t, "second", fileAppFields(), fileAppPDF(128))
	if raw, _ = reviewT28Storage(t, e, app); string(raw) != "null" {
		t.Fatalf("storage_account_email of a заявка = %s, want null", raw)
	}
}

// The same оператор, off every roster, is still refused on the association paths.
func TestReview_OperatorStaysRefusedOnAssociationPathsAfterReview(t *testing.T) {
	e := newReviewEnv(t)
	id := reviewT28Transfer(t, e, "first")
	reviewT28Connect(t, e, "files@example.test")
	e.review(t, id)
	for _, p := range []string{
		"/identity/associations/" + reviewT28AssocA + "/storage-connection",
	} {
		status, _, body := e.do(t, "operator", http.MethodGet, p, "")
		if status != http.StatusForbidden {
			t.Errorf("GET %s as operator = %d %s, want 403", p, status, body)
		}
	}
}

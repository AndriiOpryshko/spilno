package identity_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

func transferFields(code string) map[string]string {
	f := fileAppFields()
	f["kind"] = "transfer"
	f["edrpou"] = code
	return f
}

// transferSetAssociation updates association A's code and state as the owner of the row would.
func transferSetAssociation(t *testing.T, e *fileAppEnv, edrpou *string, state string) {
	t.Helper()
	aid := uuid.MustParse(e.assocA)
	err := db.InAssociation(context.Background(), e.pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE associations SET edrpou = $2, state = $3 WHERE id = $1`, aid, edrpou, state)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
}

func transferAttachedTo(t *testing.T, e *fileAppEnv, body []byte) (kind, state string, assoc *uuid.UUID) {
	t.Helper()
	var app struct{ ID string }
	if err := json.Unmarshal(body, &app); err != nil {
		t.Fatal(err)
	}
	err := e.pool.QueryRow(context.Background(), `SELECT kind, state, association_id FROM applications WHERE id = $1`, app.ID).Scan(&kind, &state, &assoc)
	if err != nil {
		t.Fatal(err)
	}
	return
}

// AC-17: any signed-in person, on the roster or not, files a запит for a connected code: 201 pending,
// attached to that association, and the body names nobody of its roster.
func TestFileTransferForAConnectedCodeIsPendingAndRevealsNoRoster(t *testing.T) {
	for _, who := range []string{"first", "head"} {
		t.Run(who, func(t *testing.T) {
			e := newFileAppEnv(t)
			code := "11223344"
			transferSetAssociation(t, e, &code, "connected")
			got := e.file(t, who, transferFields(code), fileAppPDF(256))
			if got.status != http.StatusCreated {
				t.Fatalf("status = %d (%s), want 201", got.status, got.body)
			}
			var body struct{ Kind, State string }
			_ = json.Unmarshal(got.body, &body)
			if body.Kind != "transfer" || body.State != "pending" {
				t.Errorf("body = %s, want a pending transfer", got.body)
			}
			for _, leak := range []string{"head@example.test", "\"head\"", "Association A"} {
				if strings.Contains(string(got.body), leak) {
					t.Errorf("body reveals %q: %s", leak, got.body)
				}
			}
			kind, state, assoc := transferAttachedTo(t, e, got.body)
			if kind != "transfer" || state != "pending" || assoc == nil || assoc.String() != e.assocA {
				t.Errorf("stored (%s, %s, %v), want a pending transfer attached to %s", kind, state, assoc, e.assocA)
			}
			if c := e.counts(t); c != [3]int{1, 1, 1} {
				t.Errorf("rows = %v, want [1 1 1]", c)
			}
		})
	}
}

// AC-18: a code no association has is 409 identity.association_not_connected and nothing is stored.
func TestFileTransferForAnUnknownCodeIsRefused(t *testing.T) {
	e := newFileAppEnv(t)
	got := e.file(t, "first", transferFields("99999999"), fileAppPDF(64))
	if got.status != http.StatusConflict || got.code() != "identity.association_not_connected" {
		t.Fatalf("status = %d code = %q (%s), want 409 identity.association_not_connected", got.status, got.code(), got.body)
	}
	if c := e.counts(t); c != [3]int{} {
		t.Errorf("rows = %v, want none stored", c)
	}
}

// AC-19c: a член правління files from the card with the association_id and a code the association
// does not have yet: attached to that association.
func TestFileTransferFromTheCardAttachesByAssociationID(t *testing.T) {
	e := newFileAppEnv(t)
	f := transferFields("55667788")
	f["association_id"] = e.assocA
	got := e.file(t, "head", f, fileAppPDF(64))
	if got.status != http.StatusCreated {
		t.Fatalf("status = %d (%s), want 201", got.status, got.body)
	}
	_, state, assoc := transferAttachedTo(t, e, got.body)
	if state != "pending" || assoc == nil || assoc.String() != e.assocA {
		t.Errorf("stored (%s, %v), want pending attached to %s", state, assoc, e.assocA)
	}
}

// AC-19c default: an association_id sent by someone who is not a current member of it answers
// identity.association_not_connected, so the id reveals nothing; nothing is stored.
func TestFileTransferWithAnAssociationIDOfANonMemberIsRefused(t *testing.T) {
	e := newFileAppEnv(t)
	f := transferFields("55667788")
	f["association_id"] = e.assocA
	got := e.file(t, "first", f, fileAppPDF(64))
	if got.status != http.StatusConflict || got.code() != "identity.association_not_connected" {
		t.Fatalf("status = %d code = %q (%s), want 409 identity.association_not_connected", got.status, got.code(), got.body)
	}
	if c := e.counts(t); c != [3]int{} {
		t.Errorf("rows = %v, want none stored", c)
	}
}

// AC-24b: a disabled association is accepted.
func TestFileTransferForADisabledAssociationIsAccepted(t *testing.T) {
	e := newFileAppEnv(t)
	code := "11223344"
	transferSetAssociation(t, e, &code, "disabled")
	got := e.file(t, "first", transferFields(code), fileAppPDF(64))
	if got.status != http.StatusCreated {
		t.Fatalf("status = %d (%s), want 201", got.status, got.body)
	}
	if _, _, assoc := transferAttachedTo(t, e, got.body); assoc == nil || assoc.String() != e.assocA {
		t.Errorf("attached to %v, want %s", assoc, e.assocA)
	}
}

// AC-02 and AC-30 hold for the запит's own fields; AC-04b: a second open запит of the same person
// for the same code is refused.
func TestFileTransferValidatesItsFieldsAndRefusesADuplicate(t *testing.T) {
	e := newFileAppEnv(t)
	code := "11223344"
	transferSetAssociation(t, e, &code, "connected")

	bad := transferFields(code)
	bad["position"] = "member"
	got := e.file(t, "first", bad, fileAppPDF(64))
	if got.status != http.StatusUnprocessableEntity || got.code() != "identity.application_invalid" || !strings.Contains(string(got.body), `"/position"`) {
		t.Fatalf("invalid: status = %d code = %q (%s), want 422 identity.application_invalid naming /position", got.status, got.code(), got.body)
	}

	ended := transferFields(code)
	ended["term_end"] = "2020-01-01"
	got = e.file(t, "first", ended, fileAppPDF(64))
	if got.status != http.StatusConflict || got.code() != "identity.term_ended" {
		t.Fatalf("ended: status = %d code = %q (%s), want 409 identity.term_ended", got.status, got.code(), got.body)
	}
	if c := e.counts(t); c != [3]int{} {
		t.Fatalf("rows = %v, want none stored", c)
	}

	if first := e.file(t, "first", transferFields(code), fileAppPDF(64)); first.status != http.StatusCreated {
		t.Fatalf("first: status = %d (%s), want 201", first.status, first.body)
	}
	dup := e.file(t, "first", transferFields(code), fileAppPDF(64))
	if dup.status != http.StatusConflict || dup.code() != "identity.application_already_open" {
		t.Fatalf("duplicate: status = %d code = %q (%s), want 409 identity.application_already_open", dup.status, dup.code(), dup.body)
	}
}

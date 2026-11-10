package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Contract this test pins for the wiring (like ApprovalConfig): identity cannot import governance,
// so listMemberships receives the roster reads as functions. cmd/api passes
// governance.Module.MembershipsOf and CurrentMemberID, adapted to these shapes.
//
//	func (m *Module) EnableMemberships(cfg MembershipsConfig)
//	type RosterSeat struct { AssociationID uuid.UUID; Position string } // "head" | "member"
//	type MembershipsConfig struct {
//		MembershipsOf   func(ctx context.Context, email string, date time.Time) ([]RosterSeat, error)
//		CurrentMemberID func(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error)
//	}

type cardEnv struct {
	*reviewEnv
	a, b string
}

// newCardEnv is newApprovalEnv with memberships enabled and these extra people: multi (head of A,
// member of B), solo (member of A), nobody (no roster). Association A is the hand-made one.
func newCardEnv(t *testing.T) *cardEnv {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	gov := governance.New(pool)
	idm := identity.New(pool)
	idm.EnableApprovals(identity.ApprovalConfig{SeatHead: func(ctx context.Context, association uuid.UUID, fullName, signInEmail string, termStart time.Time, decisionID uuid.UUID) (uuid.UUID, error) {
		return gov.SeatHead(ctx, association, governance.Person{FullName: fullName, SignInEmail: signInEmail}, governance.Term{Start: termStart}, decisionID)
	}})
	idm.EnableMemberships(identity.MembershipsConfig{
		MembershipsOf: func(ctx context.Context, email string, date time.Time) ([]identity.RosterSeat, error) {
			ms, err := gov.MembershipsOf(ctx, email, date)
			out := make([]identity.RosterSeat, 0, len(ms))
			for _, m := range ms {
				out = append(out, identity.RosterSeat{AssociationID: m.AssociationID, Position: m.Position})
			}
			return out, err
		},
		CurrentMemberID: gov.CurrentMemberID,
	})
	people := map[string]identity.ProviderIdentity{}
	for p, email := range map[string]string{"first": "first@example.test", "multi": "multi@example.test", "solo": "solo@example.test", "nobody": "nobody@example.test", "operator": guardOperatorEmail} {
		people[p] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + p, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: &signInFake{people: people}, OperatorEmail: guardOperatorEmail})
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardRoster{gov}))
		if err := idm.Routes(r); err != nil {
			t.Fatal(err)
		}
	})
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)

	seed := func(assoc, email, position string) {
		id := uuid.MustParse(assoc)
		if err := db.InAssociation(context.Background(), pool, id, "seed@example.test", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			governancetest.Member(t, tx, id, email, position)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	seed(a, "multi@example.test", "head")
	seed(b, "multi@example.test", "member")
	seed(a, "solo@example.test", "member")

	env := &reviewEnv{srv: srv, pool: pool, cookies: map[string]*http.Cookie{}}
	for p := range people {
		c := sessionCookieOf(signInAs(t, srv, p))
		if c == nil {
			t.Fatalf("no session cookie for %s", p)
		}
		env.cookies[p] = c
	}
	return &cardEnv{reviewEnv: env, a: a, b: b}
}

type cardEvent struct {
	Kind           string  `json:"kind"`
	OccurredAt     string  `json:"occurred_at"`
	DecisionID     *string `json:"decision_id"`
	FiledAt        *string `json:"filed_at"`
	Reason         *string `json:"reason"`
	HeadFullName   *string `json:"head_full_name"`
	SignerFullName *string `json:"signer_full_name"`
	ExtractID      *string `json:"extract_id"`
	ExtractDatedOn *string `json:"extract_dated_on"`
}

type cardView struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	State           string      `json:"state"`
	HeadFullName    *string     `json:"head_full_name"`
	ConnectedByHand bool        `json:"connected_by_hand"`
	History         []cardEvent `json:"history"`
}

type cardMemberships struct {
	Items []struct {
		AssociationID    string `json:"association_id"`
		AssociationName  string `json:"association_name"`
		AssociationState string `json:"association_state"`
		BoardMemberID    string `json:"board_member_id"`
		Position         string `json:"position"`
	} `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

func cardGet(t *testing.T, e *cardEnv, as, assoc string) cardView {
	t.Helper()
	status, _, body := e.do(t, as, http.MethodGet, "/identity/associations/"+assoc, "")
	if status != http.StatusOK {
		t.Fatalf("getAssociationCard as %s = %d, want 200: %s", as, status, body)
	}
	var v cardView
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func cardMembershipsOf(t *testing.T, e *cardEnv, as string) cardMemberships {
	t.Helper()
	status, _, body := e.do(t, as, http.MethodGet, "/identity/memberships", "")
	if status != http.StatusOK {
		t.Fatalf("listMemberships as %s = %d, want 200: %s", as, status, body)
	}
	var v cardMemberships
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

// cardApproved files, checks and approves a заявка by "first"; returns the association, the
// decision, the extract id and the filing day.
func cardApproved(t *testing.T, e *cardEnv) (assoc, decision, extract string) {
	t.Helper()
	f := approvalFile(t, e.reviewEnv, "first")
	approvalGoodCheck(t, e.reviewEnv, f)
	status, body := approvalApprove(e.reviewEnv, t, "operator", f, "")
	if status != http.StatusOK {
		t.Fatalf("approve = %d: %s", status, body)
	}
	var d struct {
		ID            string `json:"id"`
		AssociationID string `json:"association_id"`
	}
	_ = json.Unmarshal(body, &d)
	extract = approvalScalar[string](t, e.reviewEnv, `SELECT extract_id::text FROM application_filings WHERE id = $1`, f.filing)
	return d.AssociationID, d.ID, extract
}

// AC-04d: a person on A and B gets both with their position and the association's state; a person
// on no roster gets an empty page; the board_member_id is the roster entry's id.
func TestMemberships_ListEveryAssociationOfThePerson(t *testing.T) {
	e := newCardEnv(t)
	if _, _, body := e.do(t, "operator", http.MethodPost, "/identity/operator/associations/"+e.b+"/disabling", `{"reason":"test reason"}`); body == nil {
		t.Fatal("no body")
	}

	got := cardMembershipsOf(t, e, "multi")
	if len(got.Items) != 2 || got.NextCursor != nil {
		t.Fatalf("memberships of multi = %+v, want exactly A and B and no next page", got)
	}
	want := map[string][2]string{e.a: {"head", "connected"}, e.b: {"member", "disabled"}}
	for _, it := range got.Items {
		w, ok := want[it.AssociationID]
		if !ok || it.Position != w[0] || it.AssociationState != w[1] || it.AssociationName == "" {
			t.Errorf("membership %+v, want position %s and state %s with a name", it, w[0], w[1])
		}
		// The roster is under row-level security: read the expected entry inside the association's scope.
		var wantID string
		if err := db.InAssociation(context.Background(), e.pool, uuid.MustParse(it.AssociationID), "", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			return tx.QueryRow(ctx, `SELECT board_member_id::text FROM board_member_revisions WHERE association_id = $1 AND sign_in_email = 'multi@example.test'`, it.AssociationID).Scan(&wantID)
		}); err != nil {
			t.Fatalf("read the roster entry of multi in %s: %v", it.AssociationID, err)
		}
		if it.BoardMemberID != wantID {
			t.Errorf("board_member_id = %s, want the roster entry %s", it.BoardMemberID, wantID)
		}
	}

	none := cardMembershipsOf(t, e, "nobody")
	if none.Items == nil || len(none.Items) != 0 || none.NextCursor != nil {
		t.Errorf("memberships of a person on no roster = %+v, want an empty items array", none)
	}
}

// AC-29: the card of an approved association has the голова, the extract date, the signer, the
// decision date; a disabling and an enabling appear newest first with their reason.
func TestCard_ApprovedAssociationShowsTheConfirmationChain(t *testing.T) {
	e := newCardEnv(t)
	assoc, decision, extract := cardApproved(t, e)

	got := cardMembershipsOf(t, e, "first")
	if len(got.Items) != 1 || got.Items[0].AssociationID != assoc || got.Items[0].Position != "head" {
		t.Fatalf("memberships of the new голова = %+v, want the new association as head", got)
	}

	card := cardGet(t, e, "first", assoc)
	if card.ID != assoc || card.Name != "Test Association" || card.State != "connected" || card.ConnectedByHand {
		t.Fatalf("card = %+v, want the connected association, not by hand", card)
	}
	if card.HeadFullName == nil || *card.HeadFullName != "Test User" {
		t.Fatalf("head_full_name = %v, want Test User", card.HeadFullName)
	}
	if len(card.History) != 1 {
		t.Fatalf("history = %+v, want the one approval", card.History)
	}
	ev := card.History[0]
	decidedAt := approvalScalar[time.Time](t, e.reviewEnv, `SELECT created_at FROM application_decisions WHERE id = $1`, decision)
	filedAt := approvalScalar[time.Time](t, e.reviewEnv, `SELECT f.created_at FROM application_filings f JOIN application_decisions d ON d.filing_id = f.id WHERE d.id = $1`, decision)
	occurred, _ := time.Parse(time.RFC3339, ev.OccurredAt)
	if ev.Kind != "connected" || ev.DecisionID == nil || *ev.DecisionID != decision || !occurred.Equal(decidedAt.Truncate(time.Microsecond)) && occurred.Sub(decidedAt).Abs() > time.Millisecond {
		t.Errorf("approval event = %+v, want connected by decision %s at %s", ev, decision, decidedAt)
	}
	if ev.FiledAt == nil {
		t.Errorf("filed_at is null, want %s", filedAt)
	} else if f, _ := time.Parse(time.RFC3339, *ev.FiledAt); f.Sub(filedAt).Abs() > time.Millisecond {
		t.Errorf("filed_at = %s, want %s", *ev.FiledAt, filedAt)
	}
	if ev.SignerFullName == nil || *ev.SignerFullName != "Test User" || ev.HeadFullName == nil || *ev.HeadFullName != "Test User" ||
		ev.ExtractID == nil || *ev.ExtractID != extract || ev.ExtractDatedOn == nil || *ev.ExtractDatedOn != reviewKyivDay(1) {
		t.Errorf("approval event = %+v, want signer and head Test User, extract %s dated %s", ev, extract, reviewKyivDay(1))
	}

	if status, _, body := e.do(t, "operator", http.MethodPost, "/identity/operator/associations/"+assoc+"/disabling", `{"reason":"Unpaid dues"}`); status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("disable = %d: %s", status, body)
	}
	if status, _, body := e.do(t, "operator", http.MethodPost, "/identity/operator/associations/"+assoc+"/enabling", ""); status != http.StatusOK && status != http.StatusCreated {
		t.Fatalf("enable = %d: %s", status, body)
	}
	card = cardGet(t, e, "first", assoc)
	kinds := []string{}
	for _, h := range card.History {
		kinds = append(kinds, h.Kind)
	}
	if len(kinds) != 3 || kinds[0] != "enabled" || kinds[1] != "disabled" || kinds[2] != "connected" {
		t.Fatalf("history kinds = %v, want [enabled disabled connected] newest first", kinds)
	}
	if r := card.History[1].Reason; r == nil || *r != "Unpaid dues" || card.History[1].DecisionID != nil || card.History[1].ExtractID != nil {
		t.Errorf("disabling event = %+v, want reason Unpaid dues and no decision or extract", card.History[1])
	}
}

// AC-21 (ADR-0007): the card is not on the allow-list of a disabled association.
func TestCard_ADisabledAssociationRefusesItsMembers(t *testing.T) {
	e := newCardEnv(t)
	assoc, _, extract := cardApproved(t, e)
	e.do(t, "operator", http.MethodPost, "/identity/operator/associations/"+assoc+"/disabling", `{"reason":"Unpaid dues"}`)
	for _, path := range []string{"", "/extracts/" + extract} {
		status, _, body := e.do(t, "first", http.MethodGet, "/identity/associations/"+assoc+path, "")
		if status != http.StatusForbidden || assocStateProblemCode(body) != "identity.association_disabled" {
			t.Errorf("GET card%s of a disabled association = %d %s, want 403 identity.association_disabled", path, status, body)
		}
	}
}

// AC-19c: the hand-made association has no confirmation: marked by hand, no head, no extract.
func TestCard_HandMadeAssociationHasNoConfirmation(t *testing.T) {
	e := newCardEnv(t)
	card := cardGet(t, e, "solo", e.a)
	if !card.ConnectedByHand || card.State != "connected" {
		t.Errorf("card = %+v, want connected_by_hand", card)
	}
	if card.HeadFullName == nil {
		// multi is seeded as head of A: the голова comes from the roster, not from a confirmation.
		t.Errorf("head_full_name is null, want the roster's head")
	}
	for _, h := range card.History {
		if h.Kind == "connected" || h.ExtractID != nil || h.DecisionID != nil {
			t.Errorf("history has a confirmation %+v, want none before a запит is approved", h)
		}
	}
	if card.History == nil {
		t.Error("history is null, want an array")
	}
}

// AC-29, AC-07: the extract downloads for members only, byte for byte; any other path is refused.
func TestCard_ExtractDownloadsForMembersOnly(t *testing.T) {
	e := newCardEnv(t)
	assoc, _, extract := cardApproved(t, e)
	path := "/identity/associations/" + assoc + "/extracts/" + extract

	status, hdr, body := e.do(t, "first", http.MethodGet, path, "")
	if status != http.StatusOK || !bytes.Equal(body, fileAppPDF(512)) {
		t.Fatalf("download as a member = %d with %d bytes, want 200 and the uploaded 512 bytes", status, len(body))
	}
	if d := hdr.Get("Content-Disposition"); len(d) < 10 || d[:10] != "attachment" {
		t.Errorf("Content-Disposition = %q, want attachment", d)
	}

	for _, who := range []string{"nobody", "solo", "multi"} {
		status, _, body := e.do(t, who, http.MethodGet, path, "")
		if status != http.StatusForbidden || assocStateProblemCode(body) != "identity.not_current_member" {
			t.Errorf("download as %s = %d %s, want 403 identity.not_current_member", who, status, body)
		}
		status, _, body = e.do(t, who, http.MethodGet, "/identity/associations/"+assoc, "")
		if status != http.StatusForbidden || assocStateProblemCode(body) != "identity.not_current_member" {
			t.Errorf("card as %s = %d %s, want 403 identity.not_current_member", who, status, body)
		}
	}

	// A member of A asks A for the extract of another association: not found, never the file.
	if status, _, body := e.do(t, "solo", http.MethodGet, "/identity/associations/"+e.a+"/extracts/"+extract, ""); status != http.StatusNotFound {
		t.Errorf("another association's extract through A = %d %s, want 404", status, body)
	}
	if status, _, body := e.do(t, "first", http.MethodGet, "/identity/associations/"+assoc+"/extracts/"+uuid.NewString(), ""); status != http.StatusNotFound {
		t.Errorf("unknown extract = %d %s, want 404", status, body)
	}
}

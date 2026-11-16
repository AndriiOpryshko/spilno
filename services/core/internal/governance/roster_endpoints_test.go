package governance_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Handler tests of listBoardMembers, createBoardMember and updateBoardMember (association-onboarding T62:
// AC-09, AC-16b, AC-20, AC-25), served behind the identity guard exactly as cmd/api does, on two seeded
// associations. Roster of A: a previous голова (term ended, ended under prevDecision), the голова
// seated under headDecision, and a plain член. Roster of B: the оператор as a member and one more person.

const (
	rosterHeadEmail     = "head@example.test"
	rosterPrevEmail     = "prev-head@example.test"
	rosterMemberEmail   = "member@example.test"
	rosterNobodyEmail   = "nobody@example.test"
	rosterOperatorEmail = "operator@example.test"
	rosterBOnlyEmail    = "b-only@example.test"
	rosterDisabledWhy   = "спір щодо складу правління"
)

type rosterProvider struct {
	people map[string]identity.ProviderIdentity
}

func (rosterProvider) AuthCodeURL(state, nonce, challenge string) string {
	return "https://accounts.example.test/auth?" + url.Values{"state": {state}}.Encode()
}

func (p rosterProvider) Exchange(_ context.Context, code, _, _ string) (identity.ProviderIdentity, error) {
	return p.people[code], nil
}

type rosterEntry struct {
	ID                string  `json:"id"`
	FullName          string  `json:"full_name"`
	SignInEmail       *string `json:"sign_in_email"`
	TermStart         string  `json:"term_start"`
	TermEnd           *string `json:"term_end"`
	RevisedByMemberID *string `json:"revised_by_member_id"`
	RevisedAt         string  `json:"revised_at"`
	Position          string  `json:"position"`
	DecisionID        *string `json:"decision_id"`
}

type rosterPage struct {
	Items      []rosterEntry `json:"items"`
	NextCursor *string       `json:"next_cursor"`
}

type rosterProblem struct {
	Status int    `json:"status"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
	Errors []struct {
		Pointer string `json:"pointer"`
	} `json:"errors"`
}

type rosterEnv struct {
	srv                       *httptest.Server
	pool                      *pgxpool.Pool
	a, b                      string
	cookies                   map[string]*http.Cookie
	headDecision, prevDecison uuid.UUID
}

func rosterDay(n int) string { return today().AddDate(0, 0, n).Format("2006-01-02") }

func newRosterEnv(t *testing.T) *rosterEnv {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	people := map[string]identity.ProviderIdentity{}
	for name, email := range map[string]string{"head": rosterHeadEmail, "nobody": rosterNobodyEmail, "operator": rosterOperatorEmail} {
		people[name] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + name, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: rosterProvider{people}, OperatorEmail: rosterOperatorEmail})
	gov := governance.New(pool, idm.HasStorageConnection)
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardedRoster{gov}))
		for _, err := range []error{idm.Routes(r), gov.Routes(r)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	e := &rosterEnv{pool: pool, a: a, b: b, cookies: map[string]*http.Cookie{},
		headDecision: uuid.Must(uuid.NewV7()), prevDecison: uuid.Must(uuid.NewV7())}
	aid, bid := uuid.MustParse(a), uuid.MustParse(b)
	in := func(assoc uuid.UUID, fn func(ctx context.Context, tx db.DBTX) error) {
		t.Helper()
		err := db.InAssociation(context.Background(), pool, assoc, "seed@example.test", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			return fn(ctx, tx)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	// Entries are created one transaction apart so that "newest entry first" is unambiguous:
	// previous head, then the approved head, then a plain member.
	in(aid, func(ctx context.Context, tx db.DBTX) error {
		// A term that runs past the transfer, so the transfer below is what ends it (EndTerm never
		// moves an earlier end, AC-19).
		governancetest.MemberTerm(t, tx, aid, rosterPrevEmail, "head", today().AddDate(0, 0, -400), today().AddDate(0, 0, 400))
		return nil
	})
	in(aid, func(ctx context.Context, tx db.DBTX) error {
		var prev uuid.UUID
		if err := tx.QueryRow(ctx, `SELECT board_member_id FROM board_member_revisions WHERE association_id = $1 AND sign_in_email = $2`, aid, rosterPrevEmail).Scan(&prev); err != nil {
			return err
		}
		return gov.EndTerm(ctx, aid, prev, today().AddDate(0, 0, -31), e.prevDecison)
	})
	in(aid, func(ctx context.Context, tx db.DBTX) error {
		_, err := gov.SeatHead(ctx, aid, governance.Person{FullName: "Test User", SignInEmail: rosterHeadEmail}, governance.Term{Start: today().AddDate(0, 0, -30)}, e.headDecision)
		return err
	})
	in(aid, func(ctx context.Context, tx db.DBTX) error {
		governancetest.MemberTerm(t, tx, aid, rosterMemberEmail, "member", today().AddDate(0, 0, -20), today().AddDate(0, 0, 400))
		return nil
	})
	in(bid, func(ctx context.Context, tx db.DBTX) error {
		governancetest.Member(t, tx, bid, rosterOperatorEmail, "member")
		governancetest.Member(t, tx, bid, rosterBOnlyEmail, "head")
		return nil
	})

	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	e.srv = srv
	for name := range people {
		e.cookies[name] = rosterSignIn(t, r, name)
	}
	return e
}

func rosterSignIn(t *testing.T, h http.Handler, code string) *http.Cookie {
	t.Helper()
	start := httptest.NewRecorder()
	h.ServeHTTP(start, httptest.NewRequest("GET", "/identity/sign-in", nil))
	loc, err := url.Parse(start.Header().Get("Location"))
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest("GET", "/identity/sign-in/callback?code="+code+"&state="+url.QueryEscape(loc.Query().Get("state")), nil)
	for _, c := range start.Result().Cookies() {
		req.AddCookie(c)
	}
	cb := httptest.NewRecorder()
	h.ServeHTTP(cb, req)
	for _, c := range cb.Result().Cookies() {
		if c.Name == "spilno_session" && c.Value != "" {
			return c
		}
	}
	t.Fatalf("sign-in as %s set no session cookie", code)
	return nil
}

// call sends the request as the named person and returns the status and the raw body.
func (e *rosterEnv) call(t *testing.T, as, method, path string, body any) (int, []byte) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, e.srv.URL+path, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.AddCookie(e.cookies[as])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

func (e *rosterEnv) board(assoc string) string {
	return "/governance/associations/" + assoc + "/board-members"
}

func (e *rosterEnv) list(t *testing.T, as, assoc, query string) rosterPage {
	t.Helper()
	status, raw := e.call(t, as, "GET", e.board(assoc)+query, nil)
	if status != http.StatusOK {
		t.Fatalf("listBoardMembers as %s on %s%s: status = %d, want 200; body %s", as, assoc, query, status, raw)
	}
	var p rosterPage
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return p
}

func (p rosterPage) byEmail(email string) (rosterEntry, bool) {
	for _, it := range p.Items {
		if it.SignInEmail != nil && *it.SignInEmail == email {
			return it, true
		}
	}
	return rosterEntry{}, false
}

func (p rosterPage) emails() []string {
	var out []string
	for _, it := range p.Items {
		if it.SignInEmail != nil {
			out = append(out, *it.SignInEmail)
		}
	}
	return out
}

func rosterProblemOf(t *testing.T, raw []byte) rosterProblem {
	t.Helper()
	var p rosterProblem
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("decode problem %s: %v", raw, err)
	}
	return p
}

func (e *rosterEnv) revisions(t *testing.T, assoc string) int {
	t.Helper()
	var n int
	id := uuid.MustParse(assoc)
	err := db.InAssociation(context.Background(), e.pool, id, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT count(*) FROM board_member_revisions WHERE association_id = $1`, id).Scan(&n)
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

// AC-09, AC-20: the approved голова lists the roster; every entry carries the position, the decision
// behind it, the term and the sign-in address; the entries of the new and the previous голова name
// their decisions.
func TestListBoardMembersShowsPositionAndDecision(t *testing.T) {
	e := newRosterEnv(t)
	page := e.list(t, "head", e.a, "")

	if got, want := page.emails(), []string{rosterMemberEmail, rosterHeadEmail, rosterPrevEmail}; !slices.Equal(got, want) {
		t.Fatalf("roster emails = %v, want %v (newest entry first)", got, want)
	}
	if page.NextCursor != nil {
		t.Errorf("next_cursor = %q, want null on the only page", *page.NextCursor)
	}

	head, _ := page.byEmail(rosterHeadEmail)
	if head.Position != "head" || head.DecisionID == nil || *head.DecisionID != e.headDecision.String() {
		t.Errorf("head entry position/decision_id = %q/%v, want head/%s", head.Position, head.DecisionID, e.headDecision)
	}
	if head.TermStart != rosterDay(-30) || head.TermEnd != nil || head.FullName != "Test User" || head.ID == "" || head.RevisedAt == "" {
		t.Errorf("head entry = %+v, want term_start %s, open term, full name, id and revised_at", head, rosterDay(-30))
	}

	prev, _ := page.byEmail(rosterPrevEmail)
	if prev.Position != "head" || prev.DecisionID == nil || *prev.DecisionID != e.prevDecison.String() {
		t.Errorf("previous head position/decision_id = %q/%v, want head/%s", prev.Position, prev.DecisionID, e.prevDecison)
	}
	if prev.TermEnd == nil || *prev.TermEnd != rosterDay(-31) {
		t.Errorf("previous head term_end = %v, want %s", prev.TermEnd, rosterDay(-31))
	}

	member, _ := page.byEmail(rosterMemberEmail)
	if member.Position != "member" || member.DecisionID != nil {
		t.Errorf("member entry position/decision_id = %q/%v, want member/null", member.Position, member.DecisionID)
	}
}

// The on parameter keeps the roster in force on that date.
func TestListBoardMembersFiltersByDate(t *testing.T) {
	e := newRosterEnv(t)
	got := e.list(t, "head", e.a, "?on="+rosterDay(0)).emails()
	if want := []string{rosterMemberEmail, rosterHeadEmail}; !slices.Equal(got, want) {
		t.Errorf("roster on today = %v, want %v (the previous head's term ended)", got, want)
	}
	got = e.list(t, "head", e.a, "?on="+rosterDay(-25)).emails()
	if want := []string{rosterHeadEmail}; !slices.Equal(got, want) {
		t.Errorf("roster 25 days ago = %v, want %v (the member's term had not started)", got, want)
	}
}

// AC-09: the approved голова adds another член правління; the entry is a plain member with no decision,
// authored by the голова, and the next list shows it.
func TestCreateBoardMemberAddsAMember(t *testing.T) {
	e := newRosterEnv(t)
	headID := mustEntry(t, e.list(t, "head", e.a, ""), rosterHeadEmail).ID
	body := map[string]any{"full_name": "Another Member", "sign_in_email": "another@example.test", "term_start": rosterDay(-1), "term_end": nil}

	status, raw := e.call(t, "head", "POST", e.board(e.a), body)
	if status != http.StatusCreated {
		t.Fatalf("createBoardMember: status = %d, want 201; body %s", status, raw)
	}
	var created rosterEntry
	if err := json.Unmarshal(raw, &created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.FullName != "Another Member" || created.TermStart != rosterDay(-1) || created.TermEnd != nil {
		t.Errorf("created = %+v, want the submitted name and term", created)
	}
	if created.SignInEmail == nil || *created.SignInEmail != "another@example.test" {
		t.Errorf("created sign_in_email = %v, want another@example.test", created.SignInEmail)
	}
	if created.Position != "member" || created.DecisionID != nil {
		t.Errorf("created position/decision_id = %q/%v, want member/null", created.Position, created.DecisionID)
	}
	if created.RevisedByMemberID == nil || *created.RevisedByMemberID != headID {
		t.Errorf("created revised_by_member_id = %v, want the acting голова's entry %s", created.RevisedByMemberID, headID)
	}

	page := e.list(t, "head", e.a, "")
	got, ok := page.byEmail("another@example.test")
	if !ok || got.ID != created.ID {
		t.Fatalf("roster after create = %v, want the new entry %s", page.emails(), created.ID)
	}
	if page.Items[0].ID != created.ID {
		t.Errorf("first entry = %s, want the new one (newest first)", page.Items[0].ID)
	}
}

// A person who never signs in has no sign-in address.
func TestCreateBoardMemberWithoutSignInAddress(t *testing.T) {
	e := newRosterEnv(t)
	status, raw := e.call(t, "head", "POST", e.board(e.a), map[string]any{"full_name": "No Login", "term_start": rosterDay(0)})
	if status != http.StatusCreated {
		t.Fatalf("createBoardMember: status = %d, want 201; body %s", status, raw)
	}
	var created rosterEntry
	if err := json.Unmarshal(raw, &created); err != nil {
		t.Fatal(err)
	}
	if created.SignInEmail != nil {
		t.Errorf("sign_in_email = %q, want null", *created.SignInEmail)
	}
}

// Changing a член правління saves a new revision of the same entry; a seat made by an approval keeps its
// position and decision (the input carries neither, so they cannot be rewritten here).
func TestUpdateBoardMemberKeepsPositionAndDecision(t *testing.T) {
	e := newRosterEnv(t)
	page := e.list(t, "head", e.a, "")
	head := mustEntry(t, page, rosterHeadEmail)
	member := mustEntry(t, page, rosterMemberEmail)

	t.Run("the seat made by an approval", func(t *testing.T) {
		body := map[string]any{"full_name": "Test User Renamed", "sign_in_email": rosterHeadEmail, "term_start": head.TermStart, "term_end": rosterDay(200)}
		status, raw := e.call(t, "head", "PUT", e.board(e.a)+"/"+head.ID, body)
		if status != http.StatusOK {
			t.Fatalf("updateBoardMember: status = %d, want 200; body %s", status, raw)
		}
		var got rosterEntry
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		if got.ID != head.ID || got.FullName != "Test User Renamed" || got.TermEnd == nil || *got.TermEnd != rosterDay(200) {
			t.Errorf("updated = %+v, want the same entry renamed with term_end %s", got, rosterDay(200))
		}
		if got.Position != "head" || got.DecisionID == nil || *got.DecisionID != e.headDecision.String() {
			t.Errorf("updated position/decision_id = %q/%v, want head/%s kept", got.Position, got.DecisionID, e.headDecision)
		}
		if got.RevisedByMemberID == nil || *got.RevisedByMemberID != head.ID {
			t.Errorf("updated revised_by_member_id = %v, want the acting голова %s", got.RevisedByMemberID, head.ID)
		}
		after := e.list(t, "head", e.a, "")
		again := mustEntry(t, after, rosterHeadEmail)
		if len(after.Items) != 3 || again.ID != head.ID || again.FullName != "Test User Renamed" || again.Position != "head" {
			t.Errorf("roster after update = %+v, want still 3 entries with the head's latest revision", after.Items)
		}
	})

	t.Run("the term of a plain member ends", func(t *testing.T) {
		body := map[string]any{"full_name": member.FullName, "sign_in_email": rosterMemberEmail, "term_start": member.TermStart, "term_end": rosterDay(-1)}
		status, raw := e.call(t, "head", "PUT", e.board(e.a)+"/"+member.ID, body)
		if status != http.StatusOK {
			t.Fatalf("updateBoardMember: status = %d, want 200; body %s", status, raw)
		}
		var got rosterEntry
		if err := json.Unmarshal(raw, &got); err != nil {
			t.Fatal(err)
		}
		if got.ID != member.ID || got.TermEnd == nil || *got.TermEnd != rosterDay(-1) || got.Position != "member" || got.DecisionID != nil {
			t.Errorf("updated = %+v, want the same member with term_end %s, position member, no decision", got, rosterDay(-1))
		}
		if now := e.list(t, "head", e.a, "?on="+rosterDay(0)).emails(); slices.Contains(now, rosterMemberEmail) {
			t.Errorf("roster today = %v, want the member gone after the term ended", now)
		}
	})

	t.Run("an unknown entry", func(t *testing.T) {
		body := map[string]any{"full_name": "Nobody", "term_start": rosterDay(0)}
		status, raw := e.call(t, "head", "PUT", e.board(e.a)+"/"+uuid.NewString(), body)
		if status != http.StatusNotFound {
			t.Errorf("updateBoardMember of an unknown id: status = %d, want 404; body %s", status, raw)
		}
	})
}

// AC-02 (carried): the 422 names every broken field at once and nothing is saved.
func TestBoardMemberWritesNameEveryBrokenField(t *testing.T) {
	e := newRosterEnv(t)
	member := mustEntry(t, e.list(t, "head", e.a, ""), rosterMemberEmail)
	bad := map[string]any{"full_name": "   ", "term_start": rosterDay(0), "term_end": rosterDay(-1)}
	before := e.revisions(t, e.a)

	for _, tc := range []struct{ name, method, path string }{
		{"create", "POST", e.board(e.a)},
		{"update", "PUT", e.board(e.a) + "/" + member.ID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, raw := e.call(t, "head", tc.method, tc.path, bad)
			if status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422; body %s", status, raw)
			}
			p := rosterProblemOf(t, raw)
			if p.Code != "governance.roster_entry_invalid" {
				t.Errorf("code = %q, want governance.roster_entry_invalid", p.Code)
			}
			var pointers []string
			for _, f := range p.Errors {
				pointers = append(pointers, f.Pointer)
			}
			slices.Sort(pointers)
			if want := []string{"/full_name", "/term_end"}; !slices.Equal(pointers, want) {
				t.Errorf("error pointers = %v, want every broken field %v", pointers, want)
			}
		})
	}
	if after := e.revisions(t, e.a); after != before {
		t.Errorf("revisions after refused writes = %d, want %d (nothing saved)", after, before)
	}
}

// AC-16b, AC-07: a person off the roster, the оператор included, is refused on all three operations with
// the same body whether the entry exists or not; an оператор who is on a roster lists it as its member.
func TestBoardMemberOperationsRefuseAPersonOffTheRoster(t *testing.T) {
	e := newRosterEnv(t)
	member := mustEntry(t, e.list(t, "head", e.a, ""), rosterMemberEmail)
	input := map[string]any{"full_name": "Test User", "term_start": rosterDay(0)}
	ops := []struct {
		name, method, path string
		body               any
	}{
		{"listBoardMembers", "GET", e.board(e.a), nil},
		{"createBoardMember", "POST", e.board(e.a), input},
		{"updateBoardMember existing", "PUT", e.board(e.a) + "/" + member.ID, input},
		{"updateBoardMember missing", "PUT", e.board(e.a) + "/" + uuid.NewString(), input},
	}
	before := e.revisions(t, e.a)
	for _, who := range []string{"nobody", "operator"} {
		var first []byte
		for _, op := range ops {
			status, raw := e.call(t, who, op.method, op.path, op.body)
			p := rosterProblemOf(t, raw)
			if status != http.StatusForbidden || p.Code != "identity.not_current_member" {
				t.Errorf("%s as %s off the roster: status = %d code = %q, want 403 identity.not_current_member", op.name, who, status, p.Code)
			}
			if first == nil {
				first = raw
			} else if !bytes.Equal(first, raw) {
				t.Errorf("%s as %s: body %s differs from the first refusal %s", op.name, who, raw, first)
			}
		}
	}
	if after := e.revisions(t, e.a); after != before {
		t.Errorf("revisions after refused calls = %d, want %d", after, before)
	}

	// The оператор is on B's roster: a member's access there, through the handler.
	page := e.list(t, "operator", e.b, "")
	if got, want := page.emails(), []string{rosterBOnlyEmail, rosterOperatorEmail}; !sameSet(got, want) {
		t.Errorf("roster of B as its member the оператор = %v, want %v", got, want)
	}
}

// AC-25: while the association is disabled, list, create and update answer 403
// identity.association_disabled with the reason and change nothing.
func TestBoardMemberOperationsRefuseADisabledAssociation(t *testing.T) {
	e := newRosterEnv(t)
	other := mustEntry(t, e.list(t, "operator", e.b, ""), rosterBOnlyEmail)
	status, raw := e.call(t, "operator", "POST", "/identity/operator/associations/"+e.b+"/disabling", map[string]any{"reason": rosterDisabledWhy})
	if status != http.StatusOK {
		t.Fatalf("disableAssociation: status = %d, want 200; body %s", status, raw)
	}
	input := map[string]any{"full_name": "Test User", "term_start": rosterDay(0)}
	before := e.revisions(t, e.b)
	for _, op := range []struct {
		name, method, path string
		body               any
	}{
		{"listBoardMembers", "GET", e.board(e.b), nil},
		{"createBoardMember", "POST", e.board(e.b), input},
		{"updateBoardMember", "PUT", e.board(e.b) + "/" + other.ID, input},
	} {
		status, raw := e.call(t, "operator", op.method, op.path, op.body)
		p := rosterProblemOf(t, raw)
		if status != http.StatusForbidden || p.Code != "identity.association_disabled" || !strings.Contains(p.Detail, rosterDisabledWhy) {
			t.Errorf("%s on a disabled association: status = %d code = %q detail = %q, want 403 identity.association_disabled carrying %q",
				op.name, status, p.Code, p.Detail, rosterDisabledWhy)
		}
	}
	if after := e.revisions(t, e.b); after != before {
		t.Errorf("revisions of the disabled association = %d, want %d unchanged", after, before)
	}
	// Another association is untouched by it.
	if got := e.list(t, "head", e.a, "").emails(); len(got) != 3 {
		t.Errorf("roster of A = %v, want its 3 entries", got)
	}
}

// ADR-0015: rows of the second association are never returned, listed, updated or counted into a page.
func TestBoardMemberOperationsAreScopedToOneAssociation(t *testing.T) {
	e := newRosterEnv(t)
	a := e.list(t, "head", e.a, "")
	for _, foreign := range []string{rosterBOnlyEmail, rosterOperatorEmail} {
		if _, ok := a.byEmail(foreign); ok {
			t.Errorf("roster of A lists %s of association B", foreign)
		}
	}
	if got := len(a.Items); got != 3 {
		t.Errorf("roster of A has %d entries, want 3", got)
	}
	b := e.list(t, "operator", e.b, "")
	for _, foreign := range []string{rosterHeadEmail, rosterPrevEmail, rosterMemberEmail} {
		if _, ok := b.byEmail(foreign); ok {
			t.Errorf("roster of B lists %s of association A", foreign)
		}
	}

	// B's entry id under A's path is a missing entry, and A's roster is not changed by it.
	bOnly := mustEntry(t, b, rosterBOnlyEmail)
	before := e.revisions(t, e.a)
	status, raw := e.call(t, "head", "PUT", e.board(e.a)+"/"+bOnly.ID, map[string]any{"full_name": "Hijack", "term_start": rosterDay(0)})
	if status != http.StatusNotFound {
		t.Errorf("updateBoardMember of B's entry under A: status = %d, want 404; body %s", status, raw)
	}
	if after := e.revisions(t, e.a); after != before {
		t.Errorf("revisions of A = %d, want %d unchanged", after, before)
	}
	if again := mustEntry(t, e.list(t, "operator", e.b, ""), rosterBOnlyEmail); again.FullName != bOnly.FullName {
		t.Errorf("B's entry name = %q, want %q unchanged", again.FullName, bOnly.FullName)
	}
}

func mustEntry(t *testing.T, p rosterPage, email string) rosterEntry {
	t.Helper()
	it, ok := p.byEmail(email)
	if !ok {
		t.Fatalf("roster %v has no entry for %s", p.emails(), email)
	}
	return it
}

func sameSet(a, b []string) bool {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	return slices.Equal(a, b)
}

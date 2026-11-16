package governance_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// Handler tests of searchRegistry and getSignedProtocol (association-onboarding T63, carrying board-protocol
// T31: AC-09, AC-26 here; AC-18, AC-41, AC-15 are the registry behaviour they serve), behind the identity
// guard on the two seeded associations of newRosterEnv.
//
// Registry of A, as signed protocols (number, meeting date, counterparty, amount in kopiyky):
//
//	1  2026-01-15  -                     -          «Про затвердження кошторису на 2026 рік»
//	2  2026-02-10  ТОВ Тест Покрівля     18450000   «Про ремонт покрівлі» (handwritten corrections noted)
//	3  2026-03-05  ПП Тест Ліфт          25000000   «Про заміну ліфта»
//	4  2026-04-20  ТОВ Тест Покрівля      9000000   «Про гідроізоляцію підвалу»
//	5  2026-06-01  ФОП Тест Прибирання   12000000   «Про прибирання території»
//	7  2026-09-10  ТОВ Тест Фарба        18450000   «Про фарбування фасаду»
//
// plus withdrawn protocols holding numbers 6 and 8. Registry of B: signed 1 (2026-05-05, the same
// counterparty and amount as A's 2) and withdrawn 3.

const (
	cpRoof  = "ТОВ Тест Покрівля"
	cpLift  = "ПП Тест Ліфт"
	cpClean = "ФОП Тест Прибирання"
	cpPaint = "ТОВ Тест Фарба"

	correctionsNote = "У пункті 1 виправлено номер рахунку від руки, підписано всіма присутніми"
)

type regDecision struct {
	ProtocolID    string  `json:"protocol_id"`
	AgendaItemID  string  `json:"agenda_item_id"`
	Number        int     `json:"number"`
	MeetingDate   string  `json:"meeting_date"`
	Subject       string  `json:"subject"`
	Counterparty  *string `json:"counterparty"`
	AmountKopiyky *int64  `json:"amount_kopiyky"`
	Currency      *string `json:"currency"`
}

type regPage struct {
	Items            []regDecision `json:"items"`
	WithdrawnNumbers []int         `json:"withdrawn_numbers"`
	NextCursor       *string       `json:"next_cursor"`
}

func (p regPage) numbers() []int {
	out := []int{}
	for _, it := range p.Items {
		out = append(out, it.Number)
	}
	return out
}

type regFile struct {
	ID                 string  `json:"id"`
	OriginalName       string  `json:"original_name"`
	MediaType          string  `json:"media_type"`
	SizeBytes          int64   `json:"size_bytes"`
	SHA256             string  `json:"sha256"`
	UploadedByMemberID *string `json:"uploaded_by_member_id"`
}

type signedProtocolBody struct {
	ProtocolID  string `json:"protocol_id"`
	Number      int    `json:"number"`
	MeetingDate string `json:"meeting_date"`
	Meeting     struct {
		ID                 string  `json:"id"`
		HeldAt             string  `json:"held_at"`
		MeetingDate        string  `json:"meeting_date"`
		ChairBoardMemberID *string `json:"chair_board_member_id"`
	} `json:"meeting"`
	AgendaItems []struct {
		ID       string  `json:"id"`
		Position int     `json:"position"`
		Subject  *string `json:"subject"`
	} `json:"agenda_items"`
	Documents []struct {
		ID   string  `json:"id"`
		File regFile `json:"file"`
	} `json:"documents"`
	PrintCopy struct {
		ID       string  `json:"id"`
		Number   int     `json:"number"`
		File     regFile `json:"file"`
		VoidedAt *string `json:"voided_at"`
	} `json:"print_copy"`
	MeetingFacts struct {
		SignedScan       regFile  `json:"signed_scan"`
		PresentMemberIDs []string `json:"present_member_ids"`
		ItemVotes        []struct {
			AgendaItemID   string `json:"agenda_item_id"`
			VotesFor       int    `json:"votes_for"`
			VotesAgainst   int    `json:"votes_against"`
			VotesAbstained int    `json:"votes_abstained"`
		} `json:"item_votes"`
		HasHandwrittenCorrections bool    `json:"has_handwritten_corrections"`
		CorrectionsNote           *string `json:"corrections_note"`
		EnteredByMemberID         string  `json:"entered_by_member_id"`
	} `json:"meeting_facts"`
	PollHistory []struct {
		ID          string `json:"id"`
		RoundNumber int    `json:"round_number"`
		Label       string `json:"label"`
		Items       []struct {
			AgendaItemID string `json:"agenda_item_id"`
			Tally        struct {
				For     int `json:"for"`
				Against int `json:"against"`
				Abstain int `json:"abstain"`
			} `json:"tally"`
		} `json:"items"`
	} `json:"poll_history"`
	LockedByMemberID string `json:"locked_by_member_id"`
	LockedAt         string `json:"locked_at"`
}

type registryEnv struct {
	*rosterEnv
	headID           string // board_members.id of A's голова
	signedA, signedB map[int]governancetest.Signed
	withdrawnA       map[int]uuid.UUID
	withdrawnB       map[int]uuid.UUID
}

func date(s string) time.Time {
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return d
}

func kop(n int64) *int64 { return &n }

func newRegistryEnv(t *testing.T) *registryEnv {
	t.Helper()
	e := &registryEnv{
		rosterEnv: newRosterEnv(t),
		signedA:   map[int]governancetest.Signed{}, signedB: map[int]governancetest.Signed{},
		withdrawnA: map[int]uuid.UUID{}, withdrawnB: map[int]uuid.UUID{},
	}
	e.headID = mustEntry(t, e.list(t, "head", e.a, ""), rosterHeadEmail).ID
	operatorB := mustEntry(t, e.list(t, "operator", e.b, ""), rosterOperatorEmail).ID

	seed := func(assoc string, fn func(tx db.DBTX)) {
		t.Helper()
		id := uuid.MustParse(assoc)
		err := db.InAssociation(context.Background(), e.pool, id, "seed@example.test", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			fn(tx)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	a, b := uuid.MustParse(e.a), uuid.MustParse(e.b)
	head := uuid.MustParse(e.headID)
	seed(e.a, func(tx db.DBTX) {
		for _, s := range []governancetest.SignedSpec{
			{Number: 1, MeetingDate: date("2026-01-15"), Subject: "Про затвердження кошторису на 2026 рік"},
			{Number: 2, MeetingDate: date("2026-02-10"), Subject: "Про ремонт покрівлі", Counterparty: cpRoof, AmountKopiyky: kop(18450000), CorrectionsNote: correctionsNote},
			{Number: 3, MeetingDate: date("2026-03-05"), Subject: "Про заміну ліфта", Counterparty: cpLift, AmountKopiyky: kop(25000000)},
			{Number: 4, MeetingDate: date("2026-04-20"), Subject: "Про гідроізоляцію підвалу", Counterparty: cpRoof, AmountKopiyky: kop(9000000)},
			{Number: 5, MeetingDate: date("2026-06-01"), Subject: "Про прибирання території", Counterparty: cpClean, AmountKopiyky: kop(12000000)},
			{Number: 7, MeetingDate: date("2026-09-10"), Subject: "Про фарбування фасаду", Counterparty: cpPaint, AmountKopiyky: kop(18450000)},
		} {
			s.LockedBy = head
			e.signedA[s.Number] = governancetest.SeedSigned(t, tx, a, s)
		}
		e.withdrawnA[6] = governancetest.SeedWithdrawn(t, tx, a, 6, date("2026-07-01"), head)
		e.withdrawnA[8] = governancetest.SeedWithdrawn(t, tx, a, 8, date("2026-10-01"), head)
	})
	seed(e.b, func(tx db.DBTX) {
		bm := uuid.MustParse(operatorB)
		e.signedB[1] = governancetest.SeedSigned(t, tx, b, governancetest.SignedSpec{
			Number: 1, MeetingDate: date("2026-05-05"), Subject: "Про ремонт покрівлі", Counterparty: cpRoof, AmountKopiyky: kop(18450000), LockedBy: bm,
		})
		e.withdrawnB[3] = governancetest.SeedWithdrawn(t, tx, b, 3, date("2026-05-20"), bm)
	})
	return e
}

func (e *registryEnv) registry(assoc string) string {
	return "/governance/associations/" + assoc + "/registry"
}

// searchRaw sends searchRegistry and returns the status and the raw body.
func (e *registryEnv) searchRaw(t *testing.T, as, assoc string, q url.Values) (int, []byte) {
	t.Helper()
	path := e.registry(assoc) + "/decisions"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	return e.call(t, as, "GET", path, nil)
}

// search sends searchRegistry as the named person and requires 200.
func (e *registryEnv) search(t *testing.T, as, assoc string, q url.Values) regPage {
	t.Helper()
	status, raw := e.searchRaw(t, as, assoc, q)
	if status != http.StatusOK {
		t.Fatalf("searchRegistry as %s on %s?%s: status = %d, want 200; body %s", as, assoc, q.Encode(), status, raw)
	}
	var p regPage
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return p
}

func (e *registryEnv) getRaw(t *testing.T, as, assoc string, protocol uuid.UUID) (int, []byte) {
	t.Helper()
	return e.call(t, as, "GET", e.registry(assoc)+"/protocols/"+protocol.String(), nil)
}

func (e *registryEnv) get(t *testing.T, as, assoc string, protocol uuid.UUID) (signedProtocolBody, []byte) {
	t.Helper()
	status, raw := e.getRaw(t, as, assoc, protocol)
	if status != http.StatusOK {
		t.Fatalf("getSignedProtocol %s as %s on %s: status = %d, want 200; body %s", protocol, as, assoc, status, raw)
	}
	var p signedProtocolBody
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	return p, raw
}

func q(kv ...string) url.Values {
	v := url.Values{}
	for i := 0; i < len(kv); i += 2 {
		v.Set(kv[i], kv[i+1])
	}
	return v
}

func sameInstant(t *testing.T, field, got string, want time.Time) {
	t.Helper()
	g, err := time.Parse(time.RFC3339, got)
	if err != nil || !g.Equal(want) {
		t.Errorf("%s = %q, want the instant %s", field, got, want.UTC().Format(time.RFC3339))
	}
}

// AC-09, AC-18: the approved голова opens the реєстр of a new association: an empty page, not a 501 and not
// a refusal; the lists are arrays, never null.
func TestSearchRegistryAnswersAnEmptyPageForANewAssociation(t *testing.T) {
	e := newRosterEnv(t)
	status, raw := e.call(t, "head", "GET", "/governance/associations/"+e.a+"/registry/decisions", nil)
	if status != http.StatusOK {
		t.Fatalf("searchRegistry as the approved голова of a new association: status = %d, want 200; body %s", status, raw)
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}
	for field, want := range map[string]string{"items": "[]", "withdrawn_numbers": "[]", "next_cursor": "null"} {
		if got := strings.TrimSpace(string(body[field])); got != want {
			t.Errorf("%s = %s, want %s on an empty registry; body %s", field, got, want, raw)
		}
	}
}

// AC-18: the registry lists the decisions of signed protocols newest meeting first, each with its number, date,
// subject, counterparty, amount and currency; a withdrawn protocol is only a gap in the numbers.
func TestSearchRegistryListsSignedDecisionsNewestFirst(t *testing.T) {
	e := newRegistryEnv(t)
	page := e.search(t, "head", e.a, nil)

	if got, want := page.numbers(), []int{7, 5, 4, 3, 2, 1}; !slices.Equal(got, want) {
		t.Fatalf("registry numbers = %v, want %v (newest meeting first, withdrawn protocols not listed)", got, want)
	}
	if page.NextCursor != nil {
		t.Errorf("next_cursor = %q, want null on the only page", *page.NextCursor)
	}
	if want := []int{6}; !slices.Equal(page.WithdrawnNumbers, want) {
		t.Errorf("withdrawn_numbers = %v, want %v (number 8 is above this page's range 1..7)", page.WithdrawnNumbers, want)
	}

	two := page.Items[4] // number 2
	s2 := e.signedA[2]
	if two.ProtocolID != s2.ProtocolID.String() || two.AgendaItemID != s2.AgendaItemID.String() || two.MeetingDate != "2026-02-10" ||
		two.Subject != "Про ремонт покрівлі" {
		t.Errorf("decision 2 = %+v, want protocol %s, agenda item %s, 2026-02-10, subject «Про ремонт покрівлі»", two, s2.ProtocolID, s2.AgendaItemID)
	}
	if two.Counterparty == nil || *two.Counterparty != cpRoof || two.AmountKopiyky == nil || *two.AmountKopiyky != 18450000 ||
		two.Currency == nil || *two.Currency != "UAH" {
		t.Errorf("decision 2 counterparty/amount/currency = %v/%v/%v, want %s/18450000/UAH", two.Counterparty, two.AmountKopiyky, two.Currency, cpRoof)
	}
	one := page.Items[5] // number 1 names no counterparty and no amount
	if one.Counterparty != nil || one.AmountKopiyky != nil || one.Currency != nil {
		t.Errorf("decision 1 counterparty/amount/currency = %v/%v/%v, want all null", one.Counterparty, one.AmountKopiyky, one.Currency)
	}
}

// AC-18: cursor pages over the whole registry, and over a filtered one; the withdrawn numbers follow each page's
// number range.
func TestSearchRegistryPagesWithACursor(t *testing.T) {
	e := newRegistryEnv(t)

	t.Run("whole registry, two a page", func(t *testing.T) {
		type step struct {
			numbers   []int
			withdrawn []int
			last      bool
		}
		steps := []step{{[]int{7, 5}, []int{6}, false}, {[]int{4, 3}, []int{}, false}, {[]int{2, 1}, []int{}, true}}
		cursor := ""
		for i, want := range steps {
			query := q("limit", "2")
			if cursor != "" {
				query.Set("cursor", cursor)
			}
			page := e.search(t, "head", e.a, query)
			if got := page.numbers(); !slices.Equal(got, want.numbers) {
				t.Fatalf("page %d numbers = %v, want %v", i+1, got, want.numbers)
			}
			if got := page.WithdrawnNumbers; !slices.Equal(got, want.withdrawn) && (len(got) != 0 || len(want.withdrawn) != 0) {
				t.Errorf("page %d withdrawn_numbers = %v, want %v", i+1, got, want.withdrawn)
			}
			if want.last != (page.NextCursor == nil) {
				t.Fatalf("page %d next_cursor = %v, want null only on the last page", i+1, page.NextCursor)
			}
			if page.NextCursor != nil {
				cursor = *page.NextCursor
			}
		}
	})

	t.Run("a filtered search", func(t *testing.T) {
		first := e.search(t, "head", e.a, q("counterparty", "тест покрівля", "limit", "1"))
		if got := first.numbers(); !slices.Equal(got, []int{4}) || first.NextCursor == nil {
			t.Fatalf("first filtered page = %v next_cursor %v, want [4] and a cursor", got, first.NextCursor)
		}
		second := e.search(t, "head", e.a, q("counterparty", "тест покрівля", "limit", "1", "cursor", *first.NextCursor))
		if got := second.numbers(); !slices.Equal(got, []int{2}) || second.NextCursor != nil {
			t.Errorf("second filtered page = %v next_cursor %v, want [2] and null", got, second.NextCursor)
		}
	})

	t.Run("a cursor this endpoint never issued", func(t *testing.T) {
		status, raw := e.searchRaw(t, "head", e.a, q("cursor", "bm90LWEtY3Vyc29y"))
		if status != http.StatusUnprocessableEntity {
			t.Errorf("searchRegistry with a foreign cursor: status = %d, want 422; body %s", status, raw)
		}
	})
}

// AC-18: one case per filter, and the filters combine with AND.
func TestSearchRegistryFilters(t *testing.T) {
	e := newRegistryEnv(t)
	for _, tc := range []struct {
		name  string
		query url.Values
		want  []int
	}{
		{"counterparty contains, ignoring case", q("counterparty", "тест покрівля"), []int{4, 2}},
		{"counterparty of a decision that names none is never matched", q("counterparty", "ТОВ"), []int{7, 4, 2}},
		{"subject contains, ignoring case", q("subject", "про РЕМОНТ"), []int{2}},
		{"subject, another word", q("subject", "ГІДРОІЗОЛЯЦ"), []int{4}},
		{"exact amount", q("amount_kopiyky", "18450000"), []int{7, 2}},
		{"amount range, both bounds inclusive", q("amount_from_kopiyky", "10000000", "amount_to_kopiyky", "20000000"), []int{7, 5, 2}},
		{"amount from", q("amount_from_kopiyky", "18450000"), []int{7, 3, 2}},
		{"amount to", q("amount_to_kopiyky", "9000000"), []int{4}},
		{"period, both ends inclusive", q("date_from", "2026-02-10", "date_to", "2026-04-20"), []int{4, 3, 2}},
		{"period from", q("date_from", "2026-06-01"), []int{7, 5}},
		{"period to", q("date_to", "2026-01-31"), []int{1}},
		{"counterparty and period", q("counterparty", "тест покрівля", "date_from", "2026-03-01"), []int{4}},
		{"counterparty and amount range", q("counterparty", "тест покрівля", "amount_from_kopiyky", "10000000", "amount_to_kopiyky", "20000000"), []int{2}},
		{"nothing matches", q("counterparty", "немає такого"), []int{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			page := e.search(t, "head", e.a, tc.query)
			if got := page.numbers(); !slices.Equal(got, tc.want) {
				t.Errorf("numbers for %s = %v, want %v", tc.query.Encode(), got, tc.want)
			}
		})
	}
}

// AC-18: a broken filter answers 422 governance.search_invalid naming every broken filter at once.
func TestSearchRegistryRejectsBrokenFilters(t *testing.T) {
	e := newRegistryEnv(t)
	const (
		amount = "/query/amount_from_kopiyky"
		period = "/query/date_from"
	)
	for _, tc := range []struct {
		name  string
		query url.Values
		want  []string
	}{
		{"range from above to", q("amount_from_kopiyky", "20000000", "amount_to_kopiyky", "10000000"), []string{amount}},
		{"period from after to", q("date_from", "2026-05-01", "date_to", "2026-04-01"), []string{period}},
		{"both at once", q("amount_from_kopiyky", "20000000", "amount_to_kopiyky", "10000000", "date_from", "2026-05-01", "date_to", "2026-04-01"), []string{amount, period}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			status, raw := e.searchRaw(t, "head", e.a, tc.query)
			if status != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422; body %s", status, raw)
			}
			p := rosterProblemOf(t, raw)
			if p.Code != "governance.search_invalid" {
				t.Errorf("code = %q, want governance.search_invalid; body %s", p.Code, raw)
			}
			var pointers []string
			for _, f := range p.Errors {
				pointers = append(pointers, f.Pointer)
			}
			slices.Sort(pointers)
			if !slices.Equal(pointers, tc.want) {
				t.Errorf("error pointers = %v, want every broken filter %v", pointers, tc.want)
			}
		})
	}
}

// ADR-0015: the second association's entries, withdrawn numbers and protocols are never returned.
func TestRegistryIsScopedToOneAssociation(t *testing.T) {
	e := newRegistryEnv(t)
	foreign := e.signedB[1].ProtocolID.String()

	// B's protocol has the counterparty and amount of A's number 2.
	for _, query := range []url.Values{nil, q("counterparty", "тест покрівля"), q("amount_kopiyky", "18450000"), q("subject", "ремонт")} {
		page := e.search(t, "head", e.a, query)
		for _, it := range page.Items {
			if it.ProtocolID == foreign {
				t.Errorf("search of A %q returned B's protocol %s", query.Encode(), foreign)
			}
		}
	}
	whole := e.search(t, "head", e.a, nil)
	if want := []int{6}; !slices.Equal(whole.WithdrawnNumbers, want) {
		t.Errorf("withdrawn_numbers of A = %v, want %v (B's withdrawn number 3 is not A's)", whole.WithdrawnNumbers, want)
	}

	b := e.search(t, "operator", e.b, nil)
	if got, want := b.numbers(), []int{1}; !slices.Equal(got, want) || b.Items[0].ProtocolID != foreign {
		t.Errorf("registry of B = %v, want only its number 1 (%s)", b.Items, foreign)
	}
	if want := []int{}; len(b.WithdrawnNumbers) != 0 {
		t.Errorf("withdrawn_numbers of B = %v, want %v (range 1..1 holds none of 3)", b.WithdrawnNumbers, want)
	}

	// A signed protocol opened under the other association's path is a missing one.
	for _, c := range []struct {
		as, assoc string
		protocol  uuid.UUID
	}{
		{"head", e.a, e.signedB[1].ProtocolID},
		{"operator", e.b, e.signedA[2].ProtocolID},
	} {
		status, raw := e.getRaw(t, c.as, c.assoc, c.protocol)
		if status != http.StatusNotFound {
			t.Errorf("getSignedProtocol %s under %s: status = %d, want 404; body %s", c.protocol, c.assoc, status, raw)
		}
	}
}

// AC-18, AC-41, AC-15: a signed protocol opens with its meeting, agenda items, cited documents, print copy and
// signed scan as file references, meeting facts with the corrections note, and the poll history labelled informal.
func TestGetSignedProtocolReturnsTheSignedProtocol(t *testing.T) {
	e := newRegistryEnv(t)
	s := e.signedA[2]
	got, _ := e.get(t, "head", e.a, s.ProtocolID)

	if got.ProtocolID != s.ProtocolID.String() || got.Number != 2 || got.MeetingDate != "2026-02-10" {
		t.Errorf("protocol_id/number/meeting_date = %s/%d/%s, want %s/2/2026-02-10", got.ProtocolID, got.Number, got.MeetingDate, s.ProtocolID)
	}
	if got.Meeting.ID != s.MeetingID.String() || got.Meeting.MeetingDate != "2026-02-10" ||
		got.Meeting.ChairBoardMemberID == nil || *got.Meeting.ChairBoardMemberID != e.headID {
		t.Errorf("meeting = %+v, want id %s, date 2026-02-10, chair %s", got.Meeting, s.MeetingID, e.headID)
	}
	sameInstant(t, "meeting.held_at", got.Meeting.HeldAt, date("2026-02-10").Add(16*time.Hour))

	if len(got.AgendaItems) != 1 || got.AgendaItems[0].ID != s.AgendaItemID.String() || got.AgendaItems[0].Position != 1 ||
		got.AgendaItems[0].Subject == nil || *got.AgendaItems[0].Subject != "Про ремонт покрівлі" {
		t.Errorf("agenda_items = %+v, want the one item %s «Про ремонт покрівлі»", got.AgendaItems, s.AgendaItemID)
	}
	if len(got.Documents) != 1 || got.Documents[0].ID != s.DocumentID.String() ||
		got.Documents[0].File.ID != s.DocumentFileID.String() || got.Documents[0].File.OriginalName != s.DocumentFileName ||
		got.Documents[0].File.SHA256 != s.DocumentSHA256 {
		t.Errorf("documents = %+v, want the cited document %s with file %s (%s)", got.Documents, s.DocumentID, s.DocumentFileID, s.DocumentFileName)
	}

	// The signed copy: the print copy and the scan, each with the file reference the download uses.
	pc := got.PrintCopy
	if pc.ID != s.PrintCopyID.String() || pc.Number != 2 || pc.VoidedAt != nil || pc.File.ID != s.PrintFileID.String() ||
		pc.File.OriginalName != s.PrintFileName || pc.File.MediaType != "application/pdf" || pc.File.SizeBytes != 91234 || pc.File.SHA256 != s.PrintSHA256 {
		t.Errorf("print_copy = %+v, want %s number 2 with file %s %s (91234 bytes, sha256 %s)", pc, s.PrintCopyID, s.PrintFileID, s.PrintFileName, s.PrintSHA256)
	}
	scan := got.MeetingFacts.SignedScan
	if scan.ID != s.ScanFileID.String() || scan.OriginalName != s.ScanFileName || scan.SizeBytes != 1834221 || scan.SHA256 != s.ScanSHA256 ||
		scan.UploadedByMemberID == nil || *scan.UploadedByMemberID != e.headID {
		t.Errorf("meeting_facts.signed_scan = %+v, want %s %s (1834221 bytes, sha256 %s) uploaded by %s", scan, s.ScanFileID, s.ScanFileName, s.ScanSHA256, e.headID)
	}

	f := got.MeetingFacts
	if !slices.Equal(f.PresentMemberIDs, []string{e.headID}) || f.EnteredByMemberID != e.headID {
		t.Errorf("meeting_facts present/entered_by = %v/%s, want [%s]/%s", f.PresentMemberIDs, f.EnteredByMemberID, e.headID, e.headID)
	}
	if len(f.ItemVotes) != 1 || f.ItemVotes[0].AgendaItemID != s.AgendaItemID.String() ||
		f.ItemVotes[0].VotesFor != 3 || f.ItemVotes[0].VotesAgainst != 0 || f.ItemVotes[0].VotesAbstained != 0 {
		t.Errorf("meeting_facts.item_votes = %+v, want 3-0-0 on %s", f.ItemVotes, s.AgendaItemID)
	}
	// AC-15: the note sits beside the frozen text.
	if !f.HasHandwrittenCorrections || f.CorrectionsNote == nil || *f.CorrectionsNote != correctionsNote {
		t.Errorf("meeting_facts corrections = %v/%v, want true/%q", f.HasHandwrittenCorrections, f.CorrectionsNote, correctionsNote)
	}

	// AC-41: the poll is history, labelled informal.
	if len(got.PollHistory) != 1 || got.PollHistory[0].ID != s.PollRoundID.String() || got.PollHistory[0].RoundNumber != 1 ||
		got.PollHistory[0].Label != governancetest.PollLabel {
		t.Fatalf("poll_history = %+v, want round 1 (%s) labelled %q", got.PollHistory, s.PollRoundID, governancetest.PollLabel)
	}
	if items := got.PollHistory[0].Items; len(items) != 1 || items[0].AgendaItemID != s.AgendaItemID.String() || items[0].Tally.For != 1 {
		t.Errorf("poll_history[0].items = %+v, want the tally 1-0-0 on %s", items, s.AgendaItemID)
	}

	if got.LockedByMemberID != e.headID {
		t.Errorf("locked_by_member_id = %s, want %s", got.LockedByMemberID, e.headID)
	}
	sameInstant(t, "locked_at", got.LockedAt, s.LockedAt)

	// A protocol with no handwritten corrections carries no note.
	clean, _ := e.get(t, "head", e.a, e.signedA[4].ProtocolID)
	if clean.MeetingFacts.HasHandwrittenCorrections || clean.MeetingFacts.CorrectionsNote != nil {
		t.Errorf("protocol 4 corrections = %v/%v, want false/null", clean.MeetingFacts.HasHandwrittenCorrections, clean.MeetingFacts.CorrectionsNote)
	}
}

// getSignedProtocol answers 404 for an unknown id and for a withdrawn protocol: the registry shows signed
// protocols only (ADR-0007).
func TestGetSignedProtocolRefusesWhatIsNotSigned(t *testing.T) {
	e := newRegistryEnv(t)
	for name, id := range map[string]uuid.UUID{"unknown id": uuid.New(), "withdrawn protocol": e.withdrawnA[6]} {
		status, raw := e.getRaw(t, "head", e.a, id)
		if status != http.StatusNotFound {
			t.Errorf("getSignedProtocol of a %s: status = %d, want 404; body %s", name, status, raw)
		}
	}
}

// ADR-0005: no Drive file or revision id, and no Drive field, in any response body of either operation.
func TestRegistryResponsesNeverCarryDriveIDs(t *testing.T) {
	e := newRegistryEnv(t)
	var bodies [][]byte
	_, raw := e.searchRaw(t, "head", e.a, nil)
	bodies = append(bodies, raw)
	for _, s := range e.signedA {
		status, raw := e.getRaw(t, "head", e.a, s.ProtocolID)
		if status != http.StatusOK {
			t.Fatalf("getSignedProtocol %s: status = %d, want 200; body %s", s.ProtocolID, status, raw)
		}
		bodies = append(bodies, raw)
	}
	var ids []string
	for _, s := range e.signedA {
		ids = append(ids, s.DriveIDs...)
	}
	if len(ids) == 0 {
		t.Fatal("fixture wrote no Drive ids")
	}
	for _, body := range bodies {
		for _, id := range ids {
			if bytes.Contains(body, []byte(id)) {
				t.Errorf("a response body contains the Drive id %s: %s", id, body)
			}
		}
		for _, field := range []string{"drive_file_id", "drive_revision_id", "gdrv-"} {
			if bytes.Contains(body, []byte(field)) {
				t.Errorf("a response body contains %q: %s", field, body)
			}
		}
	}
}

// AC-26: a member of a вимкнене ОСББ opens the registry and a signed protocol as before disabling, the same
// bodies; the other association is unaffected.
func TestRegistryStaysReadableWhileTheAssociationIsDisabled(t *testing.T) {
	e := newRegistryEnv(t)
	s := e.signedA[2]
	searchStatus, searchBefore := e.searchRaw(t, "head", e.a, nil)
	getStatus, getBefore := e.getRaw(t, "head", e.a, s.ProtocolID)
	if searchStatus != http.StatusOK || getStatus != http.StatusOK {
		t.Fatalf("before disabling: searchRegistry %d, getSignedProtocol %d, want 200 both; bodies %s / %s", searchStatus, getStatus, searchBefore, getBefore)
	}

	status, raw := e.call(t, "operator", "POST", "/identity/operator/associations/"+e.a+"/disabling", map[string]any{"reason": rosterDisabledWhy})
	if status != http.StatusOK {
		t.Fatalf("disableAssociation: status = %d, want 200; body %s", status, raw)
	}

	searchStatus, searchAfter := e.searchRaw(t, "head", e.a, nil)
	if searchStatus != http.StatusOK {
		t.Errorf("searchRegistry on a disabled association: status = %d, want 200; body %s", searchStatus, searchAfter)
	} else if !bytes.Equal(searchBefore, searchAfter) {
		t.Errorf("searchRegistry body changed by disabling:\n before %s\n after  %s", searchBefore, searchAfter)
	}
	getStatus, getAfter := e.getRaw(t, "head", e.a, s.ProtocolID)
	if getStatus != http.StatusOK {
		t.Errorf("getSignedProtocol on a disabled association: status = %d, want 200; body %s", getStatus, getAfter)
	} else if !bytes.Equal(getBefore, getAfter) {
		t.Errorf("getSignedProtocol body changed by disabling:\n before %s\n after  %s", getBefore, getAfter)
	}
	if page := e.search(t, "head", e.a, q("counterparty", "тест покрівля")); !slices.Equal(page.numbers(), []int{4, 2}) {
		t.Errorf("a filtered search on a disabled association = %v, want [4 2]", page.numbers())
	}
	if got := e.search(t, "operator", e.b, nil).numbers(); !slices.Equal(got, []int{1}) {
		t.Errorf("registry of the other, enabled association = %v, want [1]", got)
	}
}

// AC-07: a person off the roster, the оператор included, is refused with the same body whether the protocol
// exists or not; this read does not leak a registry.
func TestRegistryRefusesAPersonOffTheRoster(t *testing.T) {
	e := newRegistryEnv(t)
	for _, who := range []string{"nobody", "operator"} {
		var first []byte
		for name, call := range map[string]func() (int, []byte){
			"searchRegistry":                 func() (int, []byte) { return e.searchRaw(t, who, e.a, nil) },
			"getSignedProtocol existing":     func() (int, []byte) { return e.getRaw(t, who, e.a, e.signedA[2].ProtocolID) },
			"getSignedProtocol missing":      func() (int, []byte) { return e.getRaw(t, who, e.a, uuid.New()) },
			"getSignedProtocol of withdrawn": func() (int, []byte) { return e.getRaw(t, who, e.a, e.withdrawnA[6]) },
		} {
			status, raw := call()
			if p := rosterProblemOf(t, raw); status != http.StatusForbidden || p.Code != "identity.not_current_member" {
				t.Errorf("%s as %s off the roster: status = %d code = %q, want 403 identity.not_current_member", name, who, status, p.Code)
			}
			if first == nil {
				first = raw
			} else if !bytes.Equal(first, raw) {
				t.Errorf("%s as %s: body %s differs from the first refusal %s", name, who, raw, first)
			}
		}
	}
}

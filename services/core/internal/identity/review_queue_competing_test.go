package identity_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type queueItem struct {
	ID        string `json:"id"`
	State     string `json:"state"`
	Competing *bool  `json:"competing"`
}

func (e *reviewEnv) queue(t *testing.T, as, path string) (items []queueItem, raw []map[string]json.RawMessage, next *string) {
	t.Helper()
	status, _, body := e.do(t, as, http.MethodGet, path, "")
	if status != http.StatusOK {
		t.Fatalf("GET %s as %s = %d: %s", path, as, status, body)
	}
	var page struct {
		Items      []queueItem `json:"items"`
		NextCursor *string     `json:"next_cursor"`
	}
	if err := json.Unmarshal(body, &page); err != nil {
		t.Fatal(err)
	}
	var rawPage struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(body, &rawPage); err != nil {
		t.Fatal(err)
	}
	return page.Items, rawPage.Items, page.NextCursor
}

func queueCompeting(t *testing.T, items []queueItem, id string) bool {
	t.Helper()
	for _, it := range items {
		if it.ID == id {
			if it.Competing == nil {
				t.Fatalf("queue item %s has no boolean `competing`; items = %+v", id, items)
			}
			return *it.Competing
		}
	}
	t.Fatalf("queue has no item %s: %+v", id, items)
	return false
}

// AC-04c: the queue item itself says whether another pending заявка holds the same ЄДРПОУ, so the
// page the оператор sees does not have to hold the competitor to mark it.
func TestQueue_MarksTwoPendingApplicationsOfOneCodeAndNotALoneOne(t *testing.T) {
	e := newReviewEnv(t)
	one := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	two := e.fileAs(t, "second", fileAppFields(), fileAppPDF(512))
	lone := fileAppFields()
	lone["edrpou"] = "87654321"
	three := e.fileAs(t, "head", lone, fileAppPDF(512))

	items, _, _ := e.queue(t, "operator", "/identity/operator/applications?limit=100")
	if !queueCompeting(t, items, one) || !queueCompeting(t, items, two) {
		t.Fatalf("two pending заявки for one code must both be competing: %+v", items)
	}
	if queueCompeting(t, items, three) {
		t.Fatalf("a lone pending заявка must not be competing: %+v", items)
	}
}

// The marker is worked out by the server over the whole queue, not over the page: a competitor on
// the next page still marks the item on this one.
func TestQueue_MarksAnItemWhoseCompetitorIsOnAnotherPage(t *testing.T) {
	e := newReviewEnv(t)
	one := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	two := e.fileAs(t, "second", fileAppFields(), fileAppPDF(512))

	items, _, next := e.queue(t, "operator", "/identity/operator/applications?limit=1")
	if len(items) != 1 || items[0].ID != one || next == nil {
		t.Fatalf("first page = %+v next=%v, want only %s and a next cursor", items, next, one)
	}
	if !queueCompeting(t, items, one) {
		t.Fatalf("the item on the first page must be competing although %s is on the next page", two)
	}
}

// A decided заявка is not a competitor: once one is rejected the other pending one is alone.
func TestQueue_ADecidedCompetitorDoesNotMark(t *testing.T) {
	e := newReviewEnv(t)
	one := e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	two := e.fileAs(t, "second", fileAppFields(), fileAppPDF(512))
	filing := e.review(t, two).Filings[0].ID
	if status, _, body := rejectionPost(e, t, "operator", two, filing, "Витяг нечитабельний"); status != http.StatusOK {
		t.Fatalf("rejectApplication = %d: %s", status, body)
	}

	pending, _, _ := e.queue(t, "operator", "/identity/operator/applications?limit=100")
	if queueCompeting(t, pending, one) {
		t.Fatalf("the pending заявка has only a rejected one beside it and must not be competing: %+v", pending)
	}
	rejected, _, _ := e.queue(t, "operator", "/identity/operator/applications?state=rejected&limit=100")
	if queueCompeting(t, rejected, two) {
		t.Fatalf("a decided заявка is never competing: %+v", rejected)
	}
}

// The applicant's own list is not told that anyone else filed for the same code.
func TestQueue_TheApplicantsOwnListCarriesNoCompetingMarker(t *testing.T) {
	e := newReviewEnv(t)
	e.fileAs(t, "first", fileAppFields(), fileAppPDF(512))
	e.fileAs(t, "second", fileAppFields(), fileAppPDF(512))

	_, raw, _ := e.queue(t, "first", "/identity/applications?limit=100")
	if len(raw) != 1 {
		t.Fatalf("own list = %d items, want 1", len(raw))
	}
	if _, leaked := raw[0]["competing"]; leaked {
		t.Fatalf("listMyApplications item leaks `competing`: %v", raw[0])
	}
}

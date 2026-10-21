package governance_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// newServer serves the context through the same router, validator and
// generated handlers as production, on a real migrated database.
func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	r := httpapi.NewRouter()
	if err := governance.New(testdb.New(t)).Routes(r); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

// validBody is a request body that passes every rule.
func validBody() map[string]any {
	return map[string]any{
		"title":           "Sample title",
		"kind":            "REGULAR",
		"held_at":         "2026-01-15T10:00:00Z",
		"ends_at":         "2026-01-15T12:00:00Z",
		"location":        "Sample location",
		"quorum_required": false,
	}
}

type problem struct {
	Status int `json:"status"`
	Errors []struct {
		Pointer string `json:"pointer"`
	} `json:"errors"`
}

// do sends a request and decodes the JSON response into out, failing the test
// unless the status is want.
func do(t *testing.T, method, url string, body any, want int, out any) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var raw json.RawMessage
	_ = json.NewDecoder(resp.Body).Decode(&raw)
	if resp.StatusCode != want {
		t.Fatalf("%s %s: status = %d, want %d; body %s", method, url, resp.StatusCode, want, raw)
	}
	if want >= 400 {
		if ct := resp.Header.Get("Content-Type"); ct != "application/problem+json" {
			t.Errorf("Content-Type = %q, want application/problem+json", ct)
		}
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			t.Fatalf("decode %s: %v", raw, err)
		}
	}
}

func TestCreateAndGetMeeting(t *testing.T) {
	srv := newServer(t)
	sent := validBody()

	var created map[string]any
	do(t, "POST", srv.URL+"/governance/meetings", sent, http.StatusCreated, &created)
	for _, k := range []string{"id", "created_at", "updated_at"} {
		if s, _ := created[k].(string); s == "" {
			t.Errorf("created[%q] = %v, want a value", k, created[k])
		}
	}
	for k, want := range sent {
		if fmt.Sprint(created[k]) != fmt.Sprint(want) {
			t.Errorf("created[%q] = %v, want %v", k, created[k], want)
		}
	}

	var got map[string]any
	do(t, "GET", srv.URL+"/governance/meetings/"+created["id"].(string), nil, http.StatusOK, &got)
	if !maps.EqualFunc(got, created, func(a, b any) bool { return fmt.Sprint(a) == fmt.Sprint(b) }) {
		t.Errorf("got = %v, want the created meeting %v", got, created)
	}
}

func TestCreateMeetingRejectsInvalidBody(t *testing.T) {
	srv := newServer(t)

	tests := []struct {
		name        string
		edit        func(map[string]any)
		wantPointer string
	}{
		{"missing title", func(b map[string]any) { delete(b, "title") }, "/title"},
		{"title too short", func(b map[string]any) { b["title"] = strings.Repeat("x", 0) }, "/title"},
		{"title too long", func(b map[string]any) { b["title"] = strings.Repeat("x", 201) }, "/title"},
		{"kind unknown", func(b map[string]any) { b["kind"] = "UNKNOWN" }, "/kind"},
		{"missing held_at", func(b map[string]any) { delete(b, "held_at") }, "/held_at"},
		{"ends_at equal to held_at", func(b map[string]any) { b["ends_at"] = b["held_at"] }, "/ends_at"},
		{"ends_at before held_at", func(b map[string]any) { b["ends_at"] = "2026-01-15T09:59:00Z" }, "/ends_at"},
		{"location too long", func(b map[string]any) { b["location"] = strings.Repeat("x", 201) }, "/location"},
		{"unknown property", func(b map[string]any) { b["not_a_field"] = 1 }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := validBody()
			tt.edit(body)
			var p problem
			do(t, "POST", srv.URL+"/governance/meetings", body, http.StatusUnprocessableEntity, &p)
			if tt.wantPointer == "" {
				return
			}
			for _, e := range p.Errors {
				if e.Pointer == tt.wantPointer {
					return
				}
			}
			t.Errorf("errors = %+v, want pointer %s", p.Errors, tt.wantPointer)
		})
	}
}

func TestGetUnknownMeeting(t *testing.T) {
	srv := newServer(t)
	do(t, "GET", srv.URL+"/governance/meetings/00000000-0000-4000-8000-000000000000", nil, http.StatusNotFound, nil)
}

func TestListMeetingsPaginates(t *testing.T) {
	srv := newServer(t)
	for range 3 {
		do(t, "POST", srv.URL+"/governance/meetings", validBody(), http.StatusCreated, nil)
	}

	type page struct {
		Items      []map[string]any `json:"items"`
		NextCursor *string          `json:"next_cursor"`
	}
	var first page
	do(t, "GET", srv.URL+"/governance/meetings?limit=2", nil, http.StatusOK, &first)
	if len(first.Items) != 2 || first.NextCursor == nil {
		t.Fatalf("first page: %d items, cursor %v; want 2 items and a cursor", len(first.Items), first.NextCursor)
	}
	var second page
	do(t, "GET", srv.URL+"/governance/meetings?limit=2&cursor="+*first.NextCursor, nil, http.StatusOK, &second)
	if len(second.Items) != 1 || second.NextCursor != nil {
		t.Fatalf("second page: %d items, cursor %v; want 1 item and no cursor", len(second.Items), second.NextCursor)
	}
	if second.Items[0]["id"] == first.Items[0]["id"] || second.Items[0]["id"] == first.Items[1]["id"] {
		t.Errorf("second page repeats an item from the first")
	}
}

func TestListMeetingsRejectsForeignCursor(t *testing.T) {
	srv := newServer(t)
	do(t, "GET", srv.URL+"/governance/meetings?cursor=bm90LWEtY3Vyc29y", nil, http.StatusUnprocessableEntity, nil)
}

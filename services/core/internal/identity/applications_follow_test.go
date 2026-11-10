package identity_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

type followApp struct {
	ID              string     `json:"id"`
	State           string     `json:"state"`
	FiledAt         time.Time  `json:"filed_at"`
	RejectionReason *string    `json:"rejection_reason"`
	AssociationID   *uuid.UUID `json:"association_id"`
}

type followDetail struct {
	Application followApp `json:"application"`
	Filings     []struct {
		ID      string `json:"id"`
		Extract struct {
			ID     string `json:"id"`
			Sha256 string `json:"sha256"`
		} `json:"extract"`
		CreatedAt time.Time `json:"created_at"`
	} `json:"filings"`
	Decisions []struct {
		Outcome string  `json:"outcome"`
		Reason  *string `json:"reason"`
	} `json:"decisions"`
}

func (e *fileAppEnv) followDo(t *testing.T, as, method, path string, body []byte, ct string) (guardReply, http.Header) {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if ct != "" {
		req.Header.Set("Content-Type", ct)
	}
	if c := e.cookies[as]; c != nil {
		req.AddCookie(c)
	}
	resp, err := signInNoRedirect.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return guardReply{resp.StatusCode, b}, resp.Header
}

// followFile files a connection заявка as `as` and returns its id.
func (e *fileAppEnv) followFile(t *testing.T, as string, extract []byte) followApp {
	t.Helper()
	got := e.file(t, as, fileAppFields(), extract)
	if got.status != http.StatusCreated {
		t.Fatalf("file: status = %d (%s), want 201", got.status, got.body)
	}
	var a followApp
	if err := json.Unmarshal(got.body, &a); err != nil {
		t.Fatal(err)
	}
	return a
}

// followReject records a rejection of the latest filing the way the review does: a decision row
// and the state moved to rejected.
func (e *fileAppEnv) followReject(t *testing.T, appID, reason string) {
	t.Helper()
	ctx := context.Background()
	var filing string
	if err := e.pool.QueryRow(ctx, `SELECT id FROM application_filings WHERE application_id = $1 ORDER BY created_at DESC, id DESC LIMIT 1`, appID).Scan(&filing); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(ctx, `INSERT INTO application_decisions (id, application_id, filing_id, outcome, reason, decided_by_email)
		VALUES ($1, $2, $3, 'rejected', $4, 'operator@example.test')`, uuid.Must(uuid.NewV7()), appID, filing, reason); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(ctx, `UPDATE applications SET state = 'rejected' WHERE id = $1`, appID); err != nil {
		t.Fatal(err)
	}
}

func (e *fileAppEnv) followRefile(t *testing.T, as, appID string, extract []byte) guardReply {
	t.Helper()
	f := fileAppFields()
	delete(f, "kind")
	delete(f, "edrpou")
	f["association_name"] = "Fixed Association"
	body, ct := fileAppMultipart(t, f, extract)
	r, _ := e.followDo(t, as, "POST", "/identity/applications/"+appID+"/filings", body, ct)
	return r
}

func (e *fileAppEnv) followDetail(t *testing.T, as, appID string) followDetail {
	t.Helper()
	r, _ := e.followDo(t, as, "GET", "/identity/applications/"+appID, nil, "")
	if r.status != http.StatusOK {
		t.Fatalf("get: status = %d (%s), want 200", r.status, r.body)
	}
	var d followDetail
	if err := json.Unmarshal(r.body, &d); err != nil {
		t.Fatal(err)
	}
	return d
}

// AC-05: the applicant lists only their own applications with state and filing date; another
// account's application is absent and answers the same 404 as a missing id.
func TestApplicantListsAndOpensOnlyTheirOwnApplications(t *testing.T) {
	e := newFileAppEnv(t)
	mine := e.followFile(t, "first", fileAppPDF(64))

	r, _ := e.followDo(t, "first", "GET", "/identity/applications", nil, "")
	if r.status != http.StatusOK {
		t.Fatalf("list: status = %d (%s), want 200", r.status, r.body)
	}
	var page struct {
		Items []followApp `json:"items"`
	}
	if err := json.Unmarshal(r.body, &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0].ID != mine.ID || page.Items[0].State != "pending" || page.Items[0].FiledAt.IsZero() {
		t.Fatalf("list = %s, want my one pending application with filed_at", r.body)
	}
	if page.Items[0].AssociationID != nil {
		t.Errorf("a pending заявка names an association: %s", r.body)
	}

	other, _ := e.followDo(t, "second", "GET", "/identity/applications", nil, "")
	if other.status != http.StatusOK || strings.Contains(string(other.body), mine.ID) {
		t.Errorf("another account's list = %d %s, want 200 without %s", other.status, other.body, mine.ID)
	}

	foreign, _ := e.followDo(t, "second", "GET", "/identity/applications/"+mine.ID, nil, "")
	missing, _ := e.followDo(t, "second", "GET", "/identity/applications/"+uuid.NewString(), nil, "")
	if foreign.status != http.StatusNotFound || missing.status != http.StatusNotFound {
		t.Fatalf("foreign = %d, missing = %d, want both 404", foreign.status, missing.status)
	}
	own := e.followDetail(t, "first", mine.ID)
	if own.Application.State != "pending" || len(own.Filings) != 1 || len(own.Decisions) != 0 {
		t.Errorf("detail = %+v, want pending with one filing and no decision", own)
	}
}

// AC-06, AC-14: a rejected application is refiled into pending with a new filing date; the old
// filing, its extract and its reason stay in the history.
func TestApplicantRefilesARejectedApplicationKeepingTheHistory(t *testing.T) {
	e := newFileAppEnv(t)
	a := e.followFile(t, "first", fileAppPDF(64))
	e.followReject(t, a.ID, "wrong head in the extract")

	d := e.followDetail(t, "first", a.ID)
	if d.Application.State != "rejected" || d.Application.RejectionReason == nil || *d.Application.RejectionReason != "wrong head in the extract" {
		t.Fatalf("rejected application = %+v, want state rejected with the reason", d.Application)
	}

	if other := e.followRefile(t, "second", a.ID, fileAppPDF(80)); other.status != http.StatusNotFound {
		t.Fatalf("refile by another account: status = %d (%s), want 404", other.status, other.body)
	}

	got := e.followRefile(t, "first", a.ID, fileAppPDF(80))
	if got.status != http.StatusCreated {
		t.Fatalf("refile: status = %d (%s), want 201", got.status, got.body)
	}
	var re followApp
	if err := json.Unmarshal(got.body, &re); err != nil {
		t.Fatal(err)
	}
	if re.ID != a.ID || re.State != "pending" || !re.FiledAt.After(a.FiledAt) {
		t.Errorf("refiled = %+v, want same id, pending, filed_at after %s", re, a.FiledAt)
	}

	d = e.followDetail(t, "first", a.ID)
	if len(d.Filings) != 2 || len(d.Decisions) != 1 || d.Decisions[0].Outcome != "rejected" || d.Decisions[0].Reason == nil || *d.Decisions[0].Reason != "wrong head in the extract" {
		t.Fatalf("history = %+v, want two filings and the rejection with its reason", d)
	}
	if !d.Filings[0].CreatedAt.After(d.Filings[1].CreatedAt) || d.Filings[0].Extract.ID == d.Filings[1].Extract.ID {
		t.Errorf("filings = %+v, want newest first, each with its own extract", d.Filings)
	}
	if c := e.counts(t); c != [3]int{1, 2, 2} {
		t.Errorf("rows (applications, filings, extracts) = %v, want [1 2 2]", c)
	}
	if n := count(t, e.pool, "application_checks"); n != 0 {
		t.Errorf("checks = %d, none may be carried to the new filing", n)
	}
}

// AC-06: only a rejected application is refiled; a code connected meanwhile refuses the refile
// and the application stays rejected.
func TestRefileIsRefusedForPendingAndForACodeConnectedMeanwhile(t *testing.T) {
	e := newFileAppEnv(t)
	a := e.followFile(t, "first", fileAppPDF(64))
	got := e.followRefile(t, "first", a.ID, fileAppPDF(80))
	if got.status != http.StatusConflict || got.code() != "identity.application_not_pending" {
		t.Fatalf("pending refile: status = %d code = %q (%s), want 409 identity.application_not_pending", got.status, got.code(), got.body)
	}

	e.followReject(t, a.ID, "no")
	aid := uuid.MustParse(e.assocA)
	err := db.InAssociation(context.Background(), e.pool, aid, "seed@example.test", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE associations SET edrpou = $2 WHERE id = $1`, aid, fileAppEdrpou)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	got = e.followRefile(t, "first", a.ID, fileAppPDF(80))
	if got.status != http.StatusConflict || got.code() != "identity.association_already_connected" {
		t.Fatalf("connected refile: status = %d code = %q (%s), want 409 identity.association_already_connected", got.status, got.code(), got.body)
	}
	if d := e.followDetail(t, "first", a.ID); d.Application.State != "rejected" || len(d.Filings) != 1 {
		t.Errorf("after the refusal = %+v, want rejected with one filing", d)
	}
}

// ADR-0005: the applicant downloads their own extract byte for byte as an attachment; another
// account's extract answers the same 404 as a missing one.
func TestApplicantDownloadsOnlyTheirOwnExtract(t *testing.T) {
	e := newFileAppEnv(t)
	upload := fileAppPDF(1024)
	a := e.followFile(t, "first", upload)
	d := e.followDetail(t, "first", a.ID)
	path := "/identity/applications/" + a.ID + "/filings/" + d.Filings[0].ID + "/extract"

	r, h := e.followDo(t, "first", "GET", path, nil, "")
	if r.status != http.StatusOK || !bytes.Equal(r.body, upload) {
		t.Fatalf("download: status = %d, %d bytes, want 200 with the %d uploaded bytes", r.status, len(r.body), len(upload))
	}
	if cd := h.Get("Content-Disposition"); !strings.HasPrefix(cd, "attachment") {
		t.Errorf("Content-Disposition = %q, want attachment", cd)
	}
	foreign, _ := e.followDo(t, "second", "GET", path, nil, "")
	missing, _ := e.followDo(t, "second", "GET", "/identity/applications/"+a.ID+"/filings/"+uuid.NewString()+"/extract", nil, "")
	if foreign.status != http.StatusNotFound || missing.status != http.StatusNotFound {
		t.Errorf("foreign = %d, missing = %d, want both 404", foreign.status, missing.status)
	}
}

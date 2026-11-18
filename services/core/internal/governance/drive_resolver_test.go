package governance_test

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
)

// T113 (AC-26, AC-28, review 2026-11-17-2 S03): the Drive resolver reads the connection once, reuses the
// association's token and bounds Drive with a deadline; an unreadable token is a storage outage.

func (e *wiringEnv) uploadedDoc(t *testing.T, as, assoc string, data []byte) caseDoc {
	t.Helper()
	got := e.attach(t, as, assoc, data)
	if got.status != http.StatusCreated {
		t.Fatalf("upload: status = %d, want 201; body %s", got.status, got.body)
	}
	return decodeCase[caseDoc](t, got)
}

func (e *wiringEnv) download(t *testing.T, as, assoc, fileID string) caseResp {
	t.Helper()
	return e.json(t, as, "GET", "/governance/associations/"+assoc+"/files/"+fileID, "")
}

func TestTwoRequestsOfOneAssociationAskTheTokenEndpointOnce(t *testing.T) {
	e := newWiringEnv(t)
	e.connect(t, e.a, "1//token-A", "folder-A")
	pdf := append([]byte("%PDF-1.7 "), bytes.Repeat([]byte{7}, 64)...)
	doc := e.uploadedDoc(t, "head", e.a, pdf)

	for i := 0; i < 2; i++ {
		if dl := e.download(t, "head", e.a, doc.File.ID); dl.status != http.StatusOK || !bytes.Equal(dl.body, pdf) {
			t.Fatalf("download %d: status = %d, want 200 and the uploaded bytes", i+1, dl.status)
		}
	}
	if n := e.drive.tokenRequests(); n != 1 {
		t.Errorf("token endpoint calls for one upload and two downloads = %d, want 1 (token reused per association)", n)
	}
}

func TestAReconnectMakesTheNextRequestUseTheNewConnection(t *testing.T) {
	e := newWiringEnv(t)
	e.connect(t, e.a, "1//token-old", "folder-old")
	e.uploadedDoc(t, "head", e.a, []byte("%PDF-1.7 one"))
	e.uploadedDoc(t, "head", e.a, []byte("%PDF-1.7 two"))
	if n := e.drive.tokenRequests(); n != 1 {
		t.Errorf("token endpoint calls before the reconnect = %d, want 1", n)
	}

	e.connect(t, e.a, "1//token-new", "folder-new")
	e.uploadedDoc(t, "head", e.a, []byte("%PDF-1.7 three"))

	ups := e.drive.uploaded()
	last := ups[len(ups)-1]
	if last.Auth != "Bearer at-for-1//token-new" || last.Parent != "folder-new" {
		t.Errorf("upload after the reconnect reached Drive as %+v, want token %q and folder %q", last, "Bearer at-for-1//token-new", "folder-new")
	}
	if n := e.drive.tokenRequests(); n != 2 {
		t.Errorf("token endpoint calls after the reconnect = %d, want 2", n)
	}
}

func TestADriveThatNeverAnswersEndsTheRequestWith503WithinTheDeadline(t *testing.T) {
	const deadline = 300 * time.Millisecond
	e := newWiringEnv(t)
	e.connect(t, e.a, "1//token-A", "folder-A")
	doc := e.uploadedDoc(t, "head", e.a, []byte("%PDF-1.7 x"))
	apiURL, uploadURL := e.drive.srv.URL+"/drive/v3", e.drive.srv.URL+"/upload/drive/v3"
	e.useResolver(governance.NewDriveResolver(e.idm, identity.ErrGoogleNotConnected, apiURL, uploadURL, governance.WithDriveTimeout(deadline)))
	e.drive.setHang(true)
	t.Cleanup(e.drive.unhang) // registered after the API server's: runs first, so the server can close

	done := make(chan caseResp, 1)
	start := time.Now()
	go func() { done <- e.download(t, "head", e.a, doc.File.ID) }()
	select {
	case got := <-done:
		if got.status != http.StatusServiceUnavailable || got.code() != "governance.storage_unavailable" {
			t.Errorf("hanging Drive: status = %d code = %q, want 503 governance.storage_unavailable; body %s", got.status, got.code(), got.body)
		}
		if took := time.Since(start); took > deadline+2*time.Second {
			t.Errorf("hanging Drive: answered after %v, want within the %v deadline", took, deadline)
		}
	case <-time.After(deadline + 3*time.Second):
		t.Fatalf("hanging Drive: no answer within %v, want 503 within the %v deadline", deadline+3*time.Second, deadline)
	}
}

func TestAnUnreadableTokenAnswers503OnUploadAndOnDownload(t *testing.T) {
	e := newWiringEnv(t)
	e.connect(t, e.a, "1//token-A", "folder-A")
	doc := e.uploadedDoc(t, "head", e.a, []byte("%PDF-1.7 x"))
	apiURL, uploadURL := e.drive.srv.URL+"/drive/v3", e.drive.srv.URL+"/upload/drive/v3"
	e.useResolver(governance.NewDriveResolver(e.otherKeyModule(t), identity.ErrGoogleNotConnected, apiURL, uploadURL))

	up := e.attach(t, "head", e.a, []byte("%PDF-1.7 y"))
	if up.status != http.StatusServiceUnavailable || up.code() != "governance.storage_unavailable" {
		t.Errorf("upload, unreadable token: status = %d code = %q, want 503 governance.storage_unavailable; body %s", up.status, up.code(), up.body)
	}
	dl := e.download(t, "head", e.a, doc.File.ID)
	if dl.status != http.StatusServiceUnavailable || dl.code() != "governance.storage_unavailable" {
		t.Errorf("download, unreadable token: status = %d code = %q, want 503 governance.storage_unavailable; body %s", dl.status, dl.code(), dl.body)
	}
}

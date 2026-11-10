package governance_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// driveGuardServer serves governance behind the identity guard; the member is on both rosters
// and only the first association has a storage connection.
func driveGuardServer(t *testing.T) (srv *httptest.Server, assocWith, assocWithout string) {
	t.Helper()
	pool, assocWith, assocWithout := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	idm.EnableSignIn(identity.SignInConfig{Provider: guardedFakeProvider{}})
	gov := governance.New(pool)
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardedRoster{gov}))
		for _, err := range []error{idm.Routes(r), gov.Routes(r)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})
	for _, a := range []string{assocWith, assocWithout} {
		id := uuid.MustParse(a)
		err := db.InAssociation(context.Background(), pool, id, "seed@example.test", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			governancetest.Member(t, tx, id, guardedMemberEmail, "member")
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	cookie := guardedSignIn(t, r)
	// The sign-in created the member's account; connect its storage on its behalf.
	a := uuid.MustParse(assocWith)
	err := db.InAssociation(context.Background(), pool, a, "", func(ctx context.Context) error {
		tx, err := db.From(ctx)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `INSERT INTO google_connections
			(id, association_id, google_account_email, refresh_token_ciphertext, refresh_token_nonce,
			 encryption_key_id, scopes, authorized_by_account_id)
			SELECT $1, $2, 'osbb@example.test', '\x00', '\x00', 'k1', 'drive.file', id FROM accounts LIMIT 1`,
			uuid.Must(uuid.NewV7()), a)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.AddCookie(cookie)
		r.ServeHTTP(w, req)
	}))
	t.Cleanup(srv.Close)
	return srv, assocWith, assocWithout
}

func driveGuardMultipart(t *testing.T, fields map[string]string, fileField string) (*bytes.Buffer, string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	if fileField != "" {
		h := textproto.MIMEHeader{}
		h.Set("Content-Disposition", `form-data; name="`+fileField+`"; filename="doc.pdf"`)
		h.Set("Content-Type", "application/pdf")
		fw, err := w.CreatePart(h)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = fw.Write([]byte("%PDF-1.4 test"))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return &buf, w.FormDataContentType()
}

// AC-28: with no storage connection both operations answer 409 governance.drive_not_connected;
// with one they reach the handler's current behaviour (501 until board-protocol T21 and T25).
func TestDriveNotConnectedRefusesDocumentsAndAssistantTurns(t *testing.T) {
	srv, assocWith, assocWithout := driveGuardServer(t)
	caseID, meetingID := uuid.NewString(), uuid.NewString()
	ops := []struct {
		name, path string
		fields     map[string]string
		file       string
	}{
		{"addCaseDocument", "/cases/" + caseID + "/documents", nil, "file"},
		{"createAssistantTurn", "/meetings/" + meetingID + "/assistant-turns", map[string]string{"message": "Підготуйте проєкт"}, ""},
	}
	for _, op := range ops {
		send := func(assoc string) (int, []byte) {
			body, ct := driveGuardMultipart(t, op.fields, op.file)
			req, err := http.NewRequest("POST", srv.URL+"/governance/associations/"+assoc+op.path, body)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", ct)
			req.Header.Set("Idempotency-Key", uuid.NewString())
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			b, _ := io.ReadAll(resp.Body)
			return resp.StatusCode, b
		}
		t.Run(op.name+"/not connected", func(t *testing.T) {
			status, b := send(assocWithout)
			if status != http.StatusConflict {
				t.Fatalf("status = %d, want 409; body %s", status, b)
			}
			var p struct {
				Code   string `json:"code"`
				Detail string `json:"detail"`
			}
			if err := json.Unmarshal(b, &p); err != nil {
				t.Fatal(err)
			}
			if p.Code != "governance.drive_not_connected" {
				t.Errorf("code = %q, want governance.drive_not_connected", p.Code)
			}
			if !strings.Contains(p.Detail, "Спершу підключіть сховище документів ОСББ") {
				t.Errorf("detail = %q, want the connect-the-storage message", p.Detail)
			}
		})
		t.Run(op.name+"/connected", func(t *testing.T) {
			status, b := send(assocWith)
			if status != http.StatusNotImplemented {
				t.Fatalf("status = %d, want 501 (current stub); body %s", status, b)
			}
		})
	}
}

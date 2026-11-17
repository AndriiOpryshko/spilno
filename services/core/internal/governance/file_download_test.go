package governance_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/governancetest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

// Handler tests of downloadFile (association-onboarding T64, carrying the downloadFile part of board-protocol
// T21: AC-25, AC-26), behind the identity guard on the two seeded associations, with a fake Files port that is
// keyed by the storage file id the seeded rows hold.
//
// Association A has a signed protocol (print copy, signed scan and a cited case document pinned by its
// snapshot) and one file no snapshot pins. Association B has its own signed protocol.

const downloadReason = "спір щодо складу правління"

// driveFiles is a ports.Files keyed by storage file id that counts the bytes it streamed out.
type driveFiles struct {
	mu          sync.Mutex
	data        map[string][]byte
	opens       int
	unavailable bool
}

var _ ports.Files = (*driveFiles)(nil)

func (d *driveFiles) put(id string, b []byte) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[id] = b
}

func (d *driveFiles) opened() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.opens
}

func (d *driveFiles) check(id string) ([]byte, error) {
	if d.unavailable {
		return nil, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	b, ok := d.data[id]
	if !ok {
		return nil, ports.ErrFileNotFound
	}
	return b, nil
}

func (d *driveFiles) Open(_ context.Context, id string) (io.ReadCloser, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := d.check(id)
	if err != nil {
		return nil, err
	}
	d.opens++
	return io.NopCloser(bytes.NewReader(b)), nil
}

func (d *driveFiles) Stat(_ context.Context, id string) (ports.FileStatus, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	b, err := d.check(id)
	if err != nil {
		return ports.FileStatus{}, err
	}
	sum := sha256.Sum256(b)
	return ports.FileStatus{StoredFile: ports.StoredFile{ID: id, Revision: "rev-" + id, SHA256: hex.EncodeToString(sum[:]), Size: int64(len(b)), MediaType: "application/pdf"}}, nil
}

func (d *driveFiles) Upload(context.Context, string, string, io.Reader) (ports.StoredFile, error) {
	return ports.StoredFile{}, nil
}
func (d *driveFiles) SetReadOnly(context.Context, string) error { return nil }
func (d *driveFiles) Untrash(context.Context, string) error     { return nil }

type downloadEnv struct {
	srv                   *httptest.Server
	a, b                  string
	cookies               map[string]*http.Cookie
	files                 *driveFiles
	signedA, signedB      governancetest.Signed
	printBytes, scanBytes []byte
	documentBytes         []byte
	unsignedID            uuid.UUID
	unsignedBytes         []byte
	allDrive              []string
}

type download struct {
	status int
	header http.Header
	body   []byte
}

func (r download) code() string {
	var p struct {
		Code string `json:"code"`
	}
	_ = json.Unmarshal(r.body, &p)
	return p.Code
}

func newDownloadEnv(t *testing.T, wire bool) *downloadEnv {
	t.Helper()
	pool, a, b := testdb.NewTwoAssociations(t)
	idm := identity.New(pool)
	people := map[string]identity.ProviderIdentity{}
	for name, email := range map[string]string{"head": rosterHeadEmail, "operator": rosterOperatorEmail} {
		people[name] = identity.ProviderIdentity{Issuer: "https://accounts.example.test", Subject: "sub-" + name, Email: email, EmailVerified: true}
	}
	idm.EnableSignIn(identity.SignInConfig{Provider: rosterProvider{people}, OperatorEmail: rosterOperatorEmail})
	e := &downloadEnv{
		a: a, b: b, cookies: map[string]*http.Cookie{},
		files:      &driveFiles{data: map[string][]byte{}},
		printBytes: []byte("%PDF-1.7 print copy of protocol 2"), scanBytes: []byte("%PDF-1.7 signed scan of protocol 2, with signatures"),
		documentBytes: []byte("%PDF-1.7 cited invoice"), unsignedBytes: []byte("%PDF-1.7 a draft attachment nobody signed"),
		unsignedID: uuid.Must(uuid.NewV7()),
	}
	gov := governance.New(pool, idm.HasStorageConnection)
	if wire {
		gov = gov.WithFiles(e.files)
	}
	r := httpapi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(idm.Guard(guardedRoster{gov}))
		for _, err := range []error{idm.Routes(r), gov.Routes(r)} {
			if err != nil {
				t.Fatal(err)
			}
		}
	})

	aid, bid := uuid.MustParse(a), uuid.MustParse(b)
	seed := func(assoc uuid.UUID, fn func(ctx context.Context, tx db.DBTX)) {
		t.Helper()
		err := db.InAssociation(context.Background(), pool, assoc, "seed@example.test", func(ctx context.Context) error {
			tx, err := db.From(ctx)
			if err != nil {
				return err
			}
			fn(ctx, tx)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	memberOf := func(assoc uuid.UUID, email, position string) (id uuid.UUID) {
		seed(assoc, func(ctx context.Context, tx db.DBTX) {
			governancetest.Member(t, tx, assoc, email, position)
			if err := tx.QueryRow(ctx, `SELECT board_member_id FROM board_member_revisions WHERE association_id = $1 AND sign_in_email = $2`, assoc, email).Scan(&id); err != nil {
				t.Fatal(err)
			}
		})
		return id
	}
	headA, headB := memberOf(aid, rosterHeadEmail, "head"), memberOf(bid, rosterBOnlyEmail, "head")
	seed(aid, func(ctx context.Context, tx db.DBTX) {
		e.signedA = governancetest.SeedSigned(t, tx, aid, governancetest.SignedSpec{
			Number: 2, MeetingDate: date("2026-02-10"), Subject: "Про ремонт покрівлі", LockedBy: headA,
			PrintBytes: e.printBytes, ScanBytes: e.scanBytes,
		})
		// The cited document keeps the seeded synthetic digest; give it the real one of its bytes.
		sum := sha256.Sum256(e.documentBytes)
		if _, err := tx.Exec(ctx, `UPDATE files SET sha256 = $2, size_bytes = $3 WHERE id = $1`, e.signedA.DocumentFileID, sum[:], len(e.documentBytes)); err != nil {
			t.Fatal(err)
		}
		usum := sha256.Sum256(e.unsignedBytes)
		if _, err := tx.Exec(ctx, `INSERT INTO files (id, association_id, drive_file_id, drive_revision_id, sha256, size_bytes, media_type, original_name, uploaded_by_member_id)
			VALUES ($1, $2, 'gdrv-file-unsigned', 'gdrv-rev-unsigned', $3, $4, 'application/pdf', 'draft-attachment.pdf', $5)`,
			e.unsignedID, aid, usum[:], len(e.unsignedBytes), headA); err != nil {
			t.Fatal(err)
		}
	})
	seed(bid, func(ctx context.Context, tx db.DBTX) {
		e.signedB = governancetest.SeedSigned(t, tx, bid, governancetest.SignedSpec{
			Number: 1, MeetingDate: date("2026-05-05"), Subject: "Про фарбування фасаду", LockedBy: headB,
			PrintBytes: []byte("%PDF-1.7 B print"), ScanBytes: []byte("%PDF-1.7 B scan"),
		})
	})
	// Storage ids: DriveIDs is print file, print rev, scan file, scan rev, document file, document rev.
	e.files.put(e.signedA.DriveIDs[0], e.printBytes)
	e.files.put(e.signedA.DriveIDs[2], e.scanBytes)
	e.files.put(e.signedA.DriveIDs[4], e.documentBytes)
	e.files.put("gdrv-file-unsigned", e.unsignedBytes)
	e.files.put(e.signedB.DriveIDs[0], []byte("%PDF-1.7 B print"))
	e.allDrive = append(append([]string{"gdrv-file-unsigned", "gdrv-rev-unsigned"}, e.signedA.DriveIDs...), e.signedB.DriveIDs...)

	e.srv = httptest.NewServer(r)
	t.Cleanup(e.srv.Close)
	for name := range people {
		e.cookies[name] = rosterSignIn(t, r, name)
	}
	return e
}

func (e *downloadEnv) call(t *testing.T, as, method, path, body string) download {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	req.AddCookie(e.cookies[as])
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return download{resp.StatusCode, resp.Header, raw}
}

func (e *downloadEnv) fetch(t *testing.T, assoc string, file uuid.UUID) download {
	t.Helper()
	return e.call(t, "head", "GET", "/governance/associations/"+assoc+"/files/"+file.String(), "")
}

func (e *downloadEnv) disableA(t *testing.T) {
	t.Helper()
	got := e.call(t, "operator", "POST", "/identity/operator/associations/"+e.a+"/disabling", `{"reason":"`+downloadReason+`"}`)
	if got.status != http.StatusOK {
		t.Fatalf("disableAssociation: status = %d, want 200; body %s", got.status, got.body)
	}
}

// requireServed checks a 200 attachment of exactly want, and that nothing storage-side is on the wire.
func (e *downloadEnv) requireServed(t *testing.T, what string, got download, name string, want []byte) {
	t.Helper()
	if got.status != http.StatusOK {
		t.Fatalf("download of %s: status = %d, want 200; body %s", what, got.status, got.body)
	}
	if !bytes.Equal(got.body, want) {
		t.Errorf("download of %s: body = %q, want the stored bytes %q", what, got.body, want)
	}
	sum := sha256.Sum256(got.body)
	wantSum := sha256.Sum256(want)
	if sum != wantSum {
		t.Errorf("download of %s: sha256 of the body = %x, want %x", what, sum, wantSum)
	}
	disp, params, err := mime.ParseMediaType(got.header.Get("Content-Disposition"))
	if err != nil || disp != "attachment" || params["filename"] != name {
		t.Errorf("download of %s: Content-Disposition = %q (%v), want attachment with filename %q", what, got.header.Get("Content-Disposition"), err, name)
	}
	if got.header.Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("download of %s: X-Content-Type-Options = %q, want nosniff", what, got.header.Get("X-Content-Type-Options"))
	}
	if ct := got.header.Get("Content-Type"); !strings.HasPrefix(ct, "application/pdf") {
		t.Errorf("download of %s: Content-Type = %q, want application/pdf", what, ct)
	}
	wire := got.header.Clone()
	wire.Del("Date")
	all := string(got.body) + wire.Get("Content-Disposition") + strings.Join(func() []string {
		var v []string
		for k, vs := range wire {
			v = append(v, k)
			v = append(v, vs...)
		}
		return v
	}(), "\n")
	for _, id := range e.allDrive {
		if strings.Contains(all, id) {
			t.Errorf("download of %s: the response carries the storage id %q", what, id)
		}
	}
}

// AC-26 (and board-protocol AC-18): a member downloads the signed copy of a signed protocol, the print copy,
// the signed scan and a cited document, as attachments, in an enabled association.
func TestDownloadFileServesTheSignedCopyAsAnAttachment(t *testing.T) {
	e := newDownloadEnv(t, true)
	s := e.signedA
	e.requireServed(t, "signed scan", e.fetch(t, e.a, s.ScanFileID), s.ScanFileName, e.scanBytes)
	e.requireServed(t, "print copy", e.fetch(t, e.a, s.PrintFileID), s.PrintFileName, e.printBytes)
	e.requireServed(t, "cited document", e.fetch(t, e.a, s.DocumentFileID), s.DocumentFileName, e.documentBytes)
}

// AC-26: the same signed copy is served unchanged once the association is disabled.
func TestDownloadFileServesASignedCopyInADisabledAssociation(t *testing.T) {
	e := newDownloadEnv(t, true)
	s := e.signedA
	e.disableA(t)
	e.requireServed(t, "signed scan (disabled)", e.fetch(t, e.a, s.ScanFileID), s.ScanFileName, e.scanBytes)
	e.requireServed(t, "print copy (disabled)", e.fetch(t, e.a, s.PrintFileID), s.PrintFileName, e.printBytes)
	e.requireServed(t, "cited document (disabled)", e.fetch(t, e.a, s.DocumentFileID), s.DocumentFileName, e.documentBytes)
}

// AC-25: a file no signed snapshot pins is served while the association is enabled, and refused with the
// reason once it is disabled; nothing is streamed from the storage for the refusal.
func TestDownloadFileRefusesAnUnsignedFileWhileDisabled(t *testing.T) {
	e := newDownloadEnv(t, true)
	e.requireServed(t, "unsigned file (enabled)", e.fetch(t, e.a, e.unsignedID), "draft-attachment.pdf", e.unsignedBytes)

	e.disableA(t)
	before := e.files.opened()
	got := e.fetch(t, e.a, e.unsignedID)
	if got.status != http.StatusForbidden || got.code() != identity.CodeAssociationDisabled {
		t.Fatalf("download of an unsigned file while disabled: status = %d code = %q, want 403 %s; body %s", got.status, got.code(), identity.CodeAssociationDisabled, got.body)
	}
	requireMatchesContractExample(t, got, "AssociationDisabled")
	var p struct {
		Detail string `json:"detail"`
	}
	_ = json.Unmarshal(got.body, &p)
	if !strings.Contains(p.Detail, downloadReason) {
		t.Errorf("detail = %q, want it to carry the reason %q", p.Detail, downloadReason)
	}
	if bytes.Contains(got.body, e.unsignedBytes) {
		t.Errorf("the refusal carries the file bytes")
	}
	if after := e.files.opened(); after != before {
		t.Errorf("storage streams opened = %d, want %d: a refused download must stream nothing", after, before)
	}
}

// Another association's file id answers 404 for a member, whether or not it is signed; so does an unknown id.
func TestDownloadFileHidesAnotherAssociationsFile(t *testing.T) {
	e := newDownloadEnv(t, true)
	for name, id := range map[string]uuid.UUID{"B's print copy": e.signedB.PrintFileID, "B's signed scan": e.signedB.ScanFileID, "unknown id": uuid.New()} {
		before := e.files.opened()
		got := e.fetch(t, e.a, id)
		if got.status != http.StatusNotFound {
			t.Errorf("download of %s through association A: status = %d, want 404; body %s", name, got.status, got.body)
		}
		if after := e.files.opened(); after != before {
			t.Errorf("download of %s: storage streams opened = %d, want %d", name, after, before)
		}
	}
}

// The stored sha256 is checked: bytes the storage returns that differ from the recorded digest are not served.
func TestDownloadFileChecksTheSHA256(t *testing.T) {
	e := newDownloadEnv(t, true)
	tampered := []byte("%PDF-1.7 somebody replaced this scan")
	e.files.put(e.signedA.DriveIDs[2], tampered)
	got := e.fetch(t, e.a, e.signedA.ScanFileID)
	if got.status == http.StatusOK || bytes.Contains(got.body, tampered) {
		t.Errorf("download of a scan whose stored bytes differ from the recorded sha256: status = %d body %q, want an error and no tampered bytes", got.status, got.body)
	}
}

// A storage outage answers 503 governance.storage_unavailable; without a wired Files (no dev storage, no
// Drive adapter yet) the operation answers the same.
func TestDownloadFileAnswersStorageUnavailable(t *testing.T) {
	t.Run("storage outage", func(t *testing.T) {
		e := newDownloadEnv(t, true)
		e.files.mu.Lock()
		e.files.unavailable = true
		e.files.mu.Unlock()
		got := e.fetch(t, e.a, e.signedA.ScanFileID)
		if got.status != http.StatusServiceUnavailable || got.code() != "governance.storage_unavailable" {
			t.Errorf("download during an outage: status = %d code = %q, want 503 governance.storage_unavailable; body %s", got.status, got.code(), got.body)
		}
	})
	t.Run("no Files wired", func(t *testing.T) {
		e := newDownloadEnv(t, false)
		got := e.fetch(t, e.a, e.signedA.ScanFileID)
		if got.status != http.StatusServiceUnavailable || got.code() != "governance.storage_unavailable" {
			t.Errorf("download with no Files wired: status = %d code = %q, want 503 governance.storage_unavailable; body %s", got.status, got.code(), got.body)
		}
	})
}

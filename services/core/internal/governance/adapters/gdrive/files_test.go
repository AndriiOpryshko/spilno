package gdrive

// Expected API (to be implemented in gdrive/files.go, story T13b):
//
//	const ReadOnlyReason = "locked by spilno"
//	type Config struct {
//		HTTP      *http.Client // authorized as the association's Google account
//		APIURL    string       // ".../drive/v3", no trailing slash
//		UploadURL string       // ".../upload/drive/v3", no trailing slash
//		FolderID  string       // every upload goes there
//	}
//	func New(cfg Config) *Files // ports.Files on Google Drive
//
// The fake Drive the tests run on is in drivefake_test.go.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/filestest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

const (
	codeStorageUnavailable = "governance.storage_unavailable"
	mediaPDF               = "application/pdf"
)

// scan returns n deterministic bytes.
func scan(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte((i*7 + i/253) % 256)
	}
	return b
}

func digest(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// newFiles returns a new fake Drive and the adapter of the association whose
// folder is folder-1.
func newFiles(t *testing.T) (*Files, *driveFake) {
	t.Helper()
	d := newDriveFake(t)
	return New(d.config()), d
}

// mustUpload stores b as scan.pdf through the adapter.
func mustUpload(t *testing.T, f *Files, b []byte) ports.StoredFile {
	t.Helper()
	got, err := f.Upload(context.Background(), "scan.pdf", mediaPDF, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	return got
}

// sentBy returns the requests Drive received while call ran.
func sentBy(d *driveFake, call func()) []driveRequest {
	before := len(d.recorded())
	call()
	return d.recorded()[before:]
}

// theRequest fails the test unless op sent exactly one request, method path.
func theRequest(t *testing.T, op string, sent []driveRequest, method, path string) driveRequest {
	t.Helper()
	if len(sent) != 1 {
		t.Fatalf("%s sent %d requests, want 1: %s %s", op, len(sent), method, path)
	}
	if sent[0].Method != method || sent[0].Path != path {
		t.Fatalf("%s sent %s %s, want %s %s", op, sent[0].Method, sent[0].Path, method, path)
	}
	return sent[0]
}

func assertStorageUnavailable(t *testing.T, op string, err error) {
	t.Helper()
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != codeStorageUnavailable {
		t.Errorf("%s: error = %v, want *domain.DomainError %s", op, err, codeStorageUnavailable)
	}
}

// AC-1: Given a new fake Drive per subtest and an adapter on it with the
// folder `folder-1`, Then `filestest.Run` passes, with the fake's `trash` as
// the harness's `Trash`.
func TestFiles_AC1_PassesTheFilesContract(t *testing.T) {
	filestest.Run(t, func(t *testing.T) filestest.Harness {
		f, d := newFiles(t)
		return filestest.Harness{
			Files: f,
			Trash: func(t *testing.T, id string) {
				t.Helper()
				if !d.trash(id) {
					t.Fatalf("trash(%q): Drive holds no such file", id)
				}
			},
		}
	})
}

// AC-2: Given that adapter, When `scan.pdf` of 1000 bytes is uploaded as
// `application/pdf`, Then Drive received exactly two requests:
// `POST /upload/drive/v3/files` with `uploadType=resumable` and a JSON body
// whose `name` is `scan.pdf`, `mimeType` is `application/pdf` and `parents` is
// `["folder-1"]`; then `PUT /upload/drive/v3/files` with the `upload_id` of the
// session, `Content-Type: application/pdf` and exactly those bytes. The
// result's `ID` and `Revision` are the `id` and `headRevisionId` of the file
// the fake holds.
func TestFiles_AC2_UploadIsAResumableSessionThenOnePut(t *testing.T) {
	f, d := newFiles(t)
	b := scan(1000)

	got, err := f.Upload(context.Background(), "scan.pdf", mediaPDF, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}

	reqs := d.recorded()
	if len(reqs) != 2 {
		t.Fatalf("Drive received %d requests, want 2 (the session POST, then the PUT): %+v", len(reqs), methodsAndPaths(reqs))
	}

	post := reqs[0]
	if post.Method != http.MethodPost || post.Path != "/upload/drive/v3/files" {
		t.Errorf("request 1 = %s %s, want POST /upload/drive/v3/files", post.Method, post.Path)
	}
	if ut := post.Query.Get("uploadType"); ut != "resumable" {
		t.Errorf("request 1 uploadType = %q, want resumable", ut)
	}
	var meta struct {
		Name     string   `json:"name"`
		MimeType string   `json:"mimeType"`
		Parents  []string `json:"parents"`
	}
	if err := json.Unmarshal(post.Body, &meta); err != nil {
		t.Fatalf("request 1 body is not JSON: %v: %q", err, post.Body)
	}
	if meta.Name != "scan.pdf" {
		t.Errorf("request 1 name = %q, want scan.pdf", meta.Name)
	}
	if meta.MimeType != mediaPDF {
		t.Errorf("request 1 mimeType = %q, want %s", meta.MimeType, mediaPDF)
	}
	if !slices.Equal(meta.Parents, []string{"folder-1"}) {
		t.Errorf("request 1 parents = %q, want [folder-1]", meta.Parents)
	}

	put := reqs[1]
	if put.Method != http.MethodPut || put.Path != "/upload/drive/v3/files" {
		t.Errorf("request 2 = %s %s, want PUT /upload/drive/v3/files", put.Method, put.Path)
	}
	// A new fake numbers its sessions from 1 and names the file after the session.
	if id := put.Query.Get("upload_id"); id != "1" {
		t.Errorf("request 2 upload_id = %q, want 1, the session the POST opened", id)
	}
	if put.ContentType != mediaPDF {
		t.Errorf("request 2 Content-Type = %q, want %s", put.ContentType, mediaPDF)
	}
	if !bytes.Equal(put.Body, b) {
		t.Errorf("request 2 body = %d bytes that differ from the %d uploaded", len(put.Body), len(b))
	}

	held, ok := d.file("file-1")
	if !ok || d.len() != 1 {
		t.Fatalf("Drive holds %d files and file-1 = %v, want exactly file-1", d.len(), ok)
	}
	if got.ID != held.ID || got.Revision != held.HeadRevisionID {
		t.Errorf("Upload = id %q revision %q, want %q %q as Drive holds them", got.ID, got.Revision, held.ID, held.HeadRevisionID)
	}
}

func methodsAndPaths(reqs []driveRequest) []string {
	out := make([]string, len(reqs))
	for i, r := range reqs {
		out[i] = r.Method + " " + r.Path
	}
	return out
}

// AC-3: Given the fake reports a wrong sha256, When a file is uploaded, Then
// the error is `ports.ErrChecksumMismatch`, the result is the zero
// `StoredFile`, the fake holds no file and its last request was
// `DELETE /drive/v3/files/<the id it had given>`; and with the fault off the
// same upload succeeds and the fake holds one file.
func TestFiles_AC3_WrongReportedSHA256FailsTheUploadAndDeletesTheFile(t *testing.T) {
	ctx := context.Background()
	f, d := newFiles(t)
	b := []byte("case document")

	d.reportWrongSHA256(true)
	got, err := f.Upload(ctx, "a.pdf", mediaPDF, bytes.NewReader(b))
	if !errors.Is(err, ports.ErrChecksumMismatch) {
		t.Fatalf("Upload error = %v, want ports.ErrChecksumMismatch", err)
	}
	if got != (ports.StoredFile{}) {
		t.Errorf("Upload returned %+v with the mismatch, want the zero StoredFile", got)
	}
	if d.len() != 0 {
		t.Errorf("Drive holds %d files after a mismatch, want 0", d.len())
	}
	reqs := d.recorded()
	if len(reqs) == 0 {
		t.Fatal("Drive received no request")
	}
	// The fake gave the first file of a new Drive the id file-1.
	last := reqs[len(reqs)-1]
	if last.Method != http.MethodDelete || last.Path != "/drive/v3/files/file-1" {
		t.Errorf("last request = %s %s, want DELETE /drive/v3/files/file-1", last.Method, last.Path)
	}

	d.reportWrongSHA256(false)
	again, err := f.Upload(ctx, "a.pdf", mediaPDF, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Upload after the fault is off: %v", err)
	}
	if d.len() != 1 {
		t.Fatalf("Drive holds %d files, want 1", d.len())
	}
	if _, ok := d.file(again.ID); !ok {
		t.Errorf("Upload returned id %q, which Drive does not hold", again.ID)
	}
}

// AC-4: Given a stored file that the storage trashed, When Drive answers `503`
// to everything, and likewise `401`, `403`, `429` and `500`, and likewise when
// its server is closed, Then `Upload`, `Open`, `Stat`, `SetReadOnly` and
// `Untrash` each return a `*domain.DomainError` with the code
// `governance.storage_unavailable`; and once Drive answers again it still
// holds one file, trashed, not read-only, with the same sha256.
func TestFiles_AC4_UnavailableDriveAnswersStorageUnavailableAndChangesNothing(t *testing.T) {
	type fault struct {
		name string
		on   func(d *driveFake)
		off  func(d *driveFake)
	}
	var faults []fault
	for _, status := range []int{
		http.StatusServiceUnavailable,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
	} {
		faults = append(faults, fault{
			name: "Drive answers " + strconv.Itoa(status),
			on:   func(d *driveFake) { d.answerEverything(status) },
			off:  func(d *driveFake) { d.answerEverything(0) },
		})
	}
	faults = append(faults, fault{
		name: "the server is closed",
		on:   func(d *driveFake) { d.close() },
		off:  func(*driveFake) {}, // what the fake holds stays readable
	})

	for _, tc := range faults {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			f, d := newFiles(t)
			b := scan(1000)
			stored := mustUpload(t, f, b)
			if !d.trash(stored.ID) {
				t.Fatalf("trash(%q): Drive holds no such file", stored.ID)
			}

			tc.on(d)
			_, errUp := f.Upload(ctx, "b.pdf", mediaPDF, bytes.NewReader([]byte("x")))
			assertStorageUnavailable(t, "Upload", errUp)
			rc, errOpen := f.Open(ctx, stored.ID)
			if rc != nil {
				_ = rc.Close()
			}
			assertStorageUnavailable(t, "Open", errOpen)
			_, errStat := f.Stat(ctx, stored.ID)
			assertStorageUnavailable(t, "Stat", errStat)
			assertStorageUnavailable(t, "SetReadOnly", f.SetReadOnly(ctx, stored.ID))
			assertStorageUnavailable(t, "Untrash", f.Untrash(ctx, stored.ID))
			tc.off(d)

			if d.len() != 1 {
				t.Fatalf("Drive holds %d files after the outage, want 1 (the failed upload stored nothing)", d.len())
			}
			held, ok := d.file(stored.ID)
			if !ok {
				t.Fatalf("Drive lost %q in the outage", stored.ID)
			}
			if !held.Trashed {
				t.Error("the file is not trashed any more after the outage")
			}
			if len(held.ContentRestrictions) != 0 {
				t.Errorf("the file got contentRestrictions %+v in the outage, want none", held.ContentRestrictions)
			}
			if held.SHA256Checksum != stored.SHA256 || held.SHA256Checksum != digest(b) {
				t.Errorf("sha256 after the outage = %q, want %q", held.SHA256Checksum, digest(b))
			}
		})
	}
}

// AC-5: Given a stored file, Then `SetReadOnly` sends
// `PATCH /drive/v3/files/<id>` whose JSON body has one `contentRestrictions`
// item with `readOnly` true and `reason` `locked by spilno`; `Untrash` sends
// `PATCH /drive/v3/files/<id>` whose JSON body has `trashed` false; `Stat`
// sends `GET /drive/v3/files/<id>` whose `fields` names `headRevisionId`,
// `sha256Checksum`, `size`, `mimeType`, `trashed` and `contentRestrictions`;
// and `Open` sends `GET /drive/v3/files/<id>` with `alt=media`.
func TestFiles_AC5_EachCallIsTheDriveRequestTheStoryNames(t *testing.T) {
	ctx := context.Background()
	f, d := newFiles(t)
	stored := mustUpload(t, f, scan(1000))
	path := "/drive/v3/files/" + stored.ID

	t.Run("SetReadOnly", func(t *testing.T) {
		sent := sentBy(d, func() {
			if err := f.SetReadOnly(ctx, stored.ID); err != nil {
				t.Errorf("SetReadOnly: %v", err)
			}
		})
		req := theRequest(t, "SetReadOnly", sent, http.MethodPatch, path)
		var body struct {
			ContentRestrictions []struct {
				ReadOnly bool   `json:"readOnly"`
				Reason   string `json:"reason"`
			} `json:"contentRestrictions"`
		}
		if err := json.Unmarshal(req.Body, &body); err != nil {
			t.Fatalf("body is not JSON: %v: %q", err, req.Body)
		}
		if len(body.ContentRestrictions) != 1 {
			t.Fatalf("body %s has %d contentRestrictions, want 1", req.Body, len(body.ContentRestrictions))
		}
		if r := body.ContentRestrictions[0]; !r.ReadOnly || r.Reason != "locked by spilno" {
			t.Errorf("contentRestrictions[0] = %+v, want readOnly true, reason %q", r, "locked by spilno")
		}
	})

	t.Run("Untrash", func(t *testing.T) {
		sent := sentBy(d, func() {
			if err := f.Untrash(ctx, stored.ID); err != nil {
				t.Errorf("Untrash: %v", err)
			}
		})
		req := theRequest(t, "Untrash", sent, http.MethodPatch, path)
		var body map[string]any
		if err := json.Unmarshal(req.Body, &body); err != nil {
			t.Fatalf("body is not JSON: %v: %q", err, req.Body)
		}
		if trashed, ok := body["trashed"]; !ok || trashed != false {
			t.Errorf("body = %s, want trashed false", req.Body)
		}
	})

	t.Run("Stat", func(t *testing.T) {
		sent := sentBy(d, func() {
			if _, err := f.Stat(ctx, stored.ID); err != nil {
				t.Errorf("Stat: %v", err)
			}
		})
		req := theRequest(t, "Stat", sent, http.MethodGet, path)
		if req.Query.Get("alt") == "media" {
			t.Error("Stat asked for alt=media, want the metadata")
		}
		asked := fieldNames(req.Query.Get("fields"))
		for _, want := range []string{
			"headRevisionId", "sha256Checksum", "size", "mimeType", "trashed", "contentRestrictions",
		} {
			if !slices.Contains(asked, want) {
				t.Errorf("fields = %q, which does not name %s", req.Query.Get("fields"), want)
			}
		}
	})

	t.Run("Open", func(t *testing.T) {
		sent := sentBy(d, func() {
			rc, err := f.Open(ctx, stored.ID)
			if err != nil {
				t.Errorf("Open: %v", err)
				return
			}
			_ = rc.Close()
		})
		req := theRequest(t, "Open", sent, http.MethodGet, path)
		if alt := req.Query.Get("alt"); alt != "media" {
			t.Errorf("alt = %q, want media", alt)
		}
	})
}

// AC-6: Given a stored file and Drive answering `400` to everything, Then each
// of the five methods returns an error that is neither `ports.ErrFileNotFound`
// nor a `*domain.DomainError`, and whose text contains `400`; and afterwards
// the fake holds one file.
func TestFiles_AC6_AnUnexpectedStatusIsAPlainErrorNamingTheStatus(t *testing.T) {
	ctx := context.Background()
	f, d := newFiles(t)
	stored := mustUpload(t, f, scan(1000))

	assertPlain400 := func(op string, err error) {
		t.Helper()
		if err == nil {
			t.Errorf("%s: no error while Drive answers 400", op)
			return
		}
		if errors.Is(err, ports.ErrFileNotFound) {
			t.Errorf("%s: error = %v, which is ports.ErrFileNotFound", op, err)
		}
		var de *domain.DomainError
		if errors.As(err, &de) {
			t.Errorf("%s: error = %v, which is a *domain.DomainError", op, err)
		}
		if !strings.Contains(err.Error(), "400") {
			t.Errorf("%s: error = %q, want the text to contain 400", op, err)
		}
	}

	d.answerEverything(http.StatusBadRequest)
	_, errUp := f.Upload(ctx, "b.pdf", mediaPDF, bytes.NewReader([]byte("x")))
	assertPlain400("Upload", errUp)
	rc, errOpen := f.Open(ctx, stored.ID)
	if rc != nil {
		_ = rc.Close()
	}
	assertPlain400("Open", errOpen)
	_, errStat := f.Stat(ctx, stored.ID)
	assertPlain400("Stat", errStat)
	assertPlain400("SetReadOnly", f.SetReadOnly(ctx, stored.ID))
	assertPlain400("Untrash", f.Untrash(ctx, stored.ID))
	d.answerEverything(0)

	if d.len() != 1 {
		t.Fatalf("Drive holds %d files afterwards, want 1", d.len())
	}
}

// AC-7: Given a stored file and a cancelled context, Then `Open`, `Stat`,
// `SetReadOnly` and `Untrash` each return an error that matches
// `context.Canceled` and is not a `*domain.DomainError`.
func TestFiles_AC7_ACancelledContextIsTheContextsError(t *testing.T) {
	f, d := newFiles(t)
	stored := mustUpload(t, f, scan(1000))
	if d.len() != 1 {
		t.Fatalf("Drive holds %d files, want the stored one", d.len())
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assertCancelled := func(op string, err error) {
		t.Helper()
		if !errors.Is(err, context.Canceled) {
			t.Errorf("%s(cancelled ctx) error = %v, want context.Canceled", op, err)
		}
		var de *domain.DomainError
		if errors.As(err, &de) {
			t.Errorf("%s(cancelled ctx) error = %v, which is a *domain.DomainError", op, err)
		}
	}

	rc, errOpen := f.Open(ctx, stored.ID)
	if rc != nil {
		_ = rc.Close()
	}
	assertCancelled("Open", errOpen)
	_, errStat := f.Stat(ctx, stored.ID)
	assertCancelled("Stat", errStat)
	assertCancelled("SetReadOnly", f.SetReadOnly(ctx, stored.ID))
	assertCancelled("Untrash", f.Untrash(ctx, stored.ID))
}

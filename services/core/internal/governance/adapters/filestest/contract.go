// Package filestest is the contract every ports.Files implementation passes:
// the in-memory fake (T13a) and the Google Drive adapter (T13b) run the same
// suite, so a use case tested on the fake behaves the same on Drive.
package filestest

// Expected API (to be implemented in governance/ports/files.go):
//
//	var ErrFileNotFound = errors.New(...)     // no file has the id
//	var ErrChecksumMismatch = errors.New(...) // the storage reports a sha256 other than the one computed while streaming
//	type StoredFile struct {
//		ID        string // storage file id, never empty
//		Revision  string // storage revision id, never empty
//		SHA256    string // lowercase hex of the uploaded bytes
//		Size      int64
//		MediaType string
//	}
//	type FileStatus struct {
//		StoredFile
//		Trashed  bool
//		ReadOnly bool
//	}
//	type Files interface {
//		Upload(ctx context.Context, name, mediaType string, body io.Reader) (StoredFile, error)
//		Open(ctx context.Context, id string) (io.ReadCloser, error)
//		Stat(ctx context.Context, id string) (FileStatus, error)
//		SetReadOnly(ctx context.Context, id string) error // idempotent
//		Untrash(ctx context.Context, id string) error     // no-op on a file that is not trashed
//	}

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"testing/iotest"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// Harness is one Files implementation under test plus the storage-side action
// the port itself does not offer.
type Harness struct {
	Files ports.Files
	// Trash moves a stored file to the storage's trash, as a person working
	// in the storage's own UI would.
	Trash func(t *testing.T, id string)
}

const mediaPDF = "application/pdf"

// body returns n deterministic, non-repeating-looking bytes.
func body(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte((i*31 + i/251) % 256)
	}
	return b
}

func digest(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func upload(t *testing.T, f ports.Files, b []byte) ports.StoredFile {
	t.Helper()
	got, err := f.Upload(context.Background(), "scan.pdf", mediaPDF, bytes.NewReader(b))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	return got
}

func readAll(t *testing.T, f ports.Files, id string) []byte {
	t.Helper()
	rc, err := f.Open(context.Background(), id)
	if err != nil {
		t.Fatalf("Open(%q): %v", id, err)
	}
	defer rc.Close()
	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read %q: %v", id, err)
	}
	return got
}

func stat(t *testing.T, f ports.Files, id string) ports.FileStatus {
	t.Helper()
	st, err := f.Stat(context.Background(), id)
	if err != nil {
		t.Fatalf("Stat(%q): %v", id, err)
	}
	return st
}

// Run runs the whole contract; newHarness is called once per subtest and must
// return an empty storage.
func Run(t *testing.T, newHarness func(t *testing.T) Harness) {
	t.Helper()

	t.Run("upload reports id, revision, sha256, size and media type", func(t *testing.T) {
		h := newHarness(t)
		b := body(64 << 10)
		got := upload(t, h.Files, b)
		if got.ID == "" || got.Revision == "" {
			t.Fatalf("Upload = %+v, want a non-empty ID and Revision", got)
		}
		if got.SHA256 != digest(b) {
			t.Errorf("SHA256 = %q, want %q", got.SHA256, digest(b))
		}
		if got.Size != int64(len(b)) {
			t.Errorf("Size = %d, want %d", got.Size, len(b))
		}
		if got.MediaType != mediaPDF {
			t.Errorf("MediaType = %q, want %q", got.MediaType, mediaPDF)
		}
	})

	t.Run("upload hashes a body that arrives in small reads", func(t *testing.T) {
		h := newHarness(t)
		b := body(3 << 20)
		got, err := h.Files.Upload(context.Background(), "big.pdf", mediaPDF, iotest.HalfReader(bytes.NewReader(b)))
		if err != nil {
			t.Fatalf("Upload: %v", err)
		}
		if got.SHA256 != digest(b) || got.Size != int64(len(b)) {
			t.Fatalf("Upload = sha256 %q size %d, want %q %d", got.SHA256, got.Size, digest(b), len(b))
		}
	})

	t.Run("open streams back the uploaded bytes", func(t *testing.T) {
		h := newHarness(t)
		b := body(200 << 10)
		got := upload(t, h.Files, b)
		if !bytes.Equal(readAll(t, h.Files, got.ID), b) {
			t.Fatal("Open returned bytes that differ from the upload")
		}
	})

	t.Run("an empty body is a valid file", func(t *testing.T) {
		h := newHarness(t)
		got := upload(t, h.Files, nil)
		if got.Size != 0 || got.SHA256 != digest(nil) {
			t.Fatalf("Upload(empty) = size %d sha256 %q, want 0 %q", got.Size, got.SHA256, digest(nil))
		}
		if len(readAll(t, h.Files, got.ID)) != 0 {
			t.Fatal("Open(empty file) returned bytes")
		}
	})

	t.Run("stat reports what upload reported, not trashed, not read-only", func(t *testing.T) {
		h := newHarness(t)
		got := upload(t, h.Files, body(1000))
		st := stat(t, h.Files, got.ID)
		if st.StoredFile != got {
			t.Errorf("Stat.StoredFile = %+v, want %+v", st.StoredFile, got)
		}
		if st.Trashed || st.ReadOnly {
			t.Errorf("Stat = trashed %v read-only %v, want both false", st.Trashed, st.ReadOnly)
		}
	})

	t.Run("the same bytes uploaded twice are two files", func(t *testing.T) {
		h := newHarness(t)
		b := body(500)
		a, c := upload(t, h.Files, b), upload(t, h.Files, b)
		if a.ID == c.ID {
			t.Fatalf("two uploads share id %q", a.ID)
		}
		if a.SHA256 != c.SHA256 {
			t.Fatalf("same bytes, different sha256: %q vs %q", a.SHA256, c.SHA256)
		}
	})

	t.Run("set read-only is visible in stat and idempotent", func(t *testing.T) {
		h := newHarness(t)
		got := upload(t, h.Files, body(100))
		for i := range 2 {
			if err := h.Files.SetReadOnly(context.Background(), got.ID); err != nil {
				t.Fatalf("SetReadOnly #%d: %v", i+1, err)
			}
		}
		st := stat(t, h.Files, got.ID)
		if !st.ReadOnly {
			t.Fatal("Stat.ReadOnly = false after SetReadOnly")
		}
		if st.SHA256 != got.SHA256 || st.Revision != got.Revision {
			t.Fatalf("SetReadOnly changed the content: %+v, uploaded %+v", st.StoredFile, got)
		}
	})

	t.Run("untrash restores a trashed file with its content", func(t *testing.T) {
		h := newHarness(t)
		b := body(4096)
		got := upload(t, h.Files, b)
		h.Trash(t, got.ID)
		if !stat(t, h.Files, got.ID).Trashed {
			t.Fatal("Stat.Trashed = false after the storage trashed the file")
		}
		if err := h.Files.Untrash(context.Background(), got.ID); err != nil {
			t.Fatalf("Untrash: %v", err)
		}
		st := stat(t, h.Files, got.ID)
		if st.Trashed {
			t.Fatal("Stat.Trashed = true after Untrash")
		}
		if st.SHA256 != digest(b) || !bytes.Equal(readAll(t, h.Files, got.ID), b) {
			t.Fatal("the untrashed file lost its content")
		}
	})

	t.Run("untrash on a file that is not trashed is a no-op", func(t *testing.T) {
		h := newHarness(t)
		got := upload(t, h.Files, body(10))
		if err := h.Files.Untrash(context.Background(), got.ID); err != nil {
			t.Fatalf("Untrash: %v", err)
		}
		if st := stat(t, h.Files, got.ID); st.Trashed || st.StoredFile != got {
			t.Fatalf("Stat after a no-op Untrash = %+v, want %+v not trashed", st, got)
		}
	})

	t.Run("an unknown id is ErrFileNotFound for every lookup", func(t *testing.T) {
		h := newHarness(t)
		ctx := context.Background()
		const id = "no-such-file"
		_, errOpen := h.Files.Open(ctx, id)
		_, errStat := h.Files.Stat(ctx, id)
		for name, err := range map[string]error{
			"Open":        errOpen,
			"Stat":        errStat,
			"SetReadOnly": h.Files.SetReadOnly(ctx, id),
			"Untrash":     h.Files.Untrash(ctx, id),
		} {
			if !errors.Is(err, ports.ErrFileNotFound) {
				t.Errorf("%s(unknown) error = %v, want ports.ErrFileNotFound", name, err)
			}
		}
	})

	t.Run("a body that fails mid-read fails the upload with that error", func(t *testing.T) {
		h := newHarness(t)
		broken := errors.New("client went away")
		r := io.MultiReader(bytes.NewReader(body(1000)), iotest.ErrReader(broken))
		_, err := h.Files.Upload(context.Background(), "cut.pdf", mediaPDF, r)
		if !errors.Is(err, broken) {
			t.Fatalf("Upload(broken body) error = %v, want it to wrap %v", err, broken)
		}
	})

	t.Run("a cancelled context fails the upload", func(t *testing.T) {
		h := newHarness(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := h.Files.Upload(ctx, "late.pdf", mediaPDF, bytes.NewReader(body(10)))
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Upload(cancelled ctx) error = %v, want context.Canceled", err)
		}
	})
}

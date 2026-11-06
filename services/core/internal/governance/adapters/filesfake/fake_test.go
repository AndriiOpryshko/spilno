package filesfake_test

// Expected API (to be implemented in filesfake/fake.go, plus the port in
// governance/ports/files.go, see the filestest package header):
//
//	func New() *Fake                      // empty storage, safe for concurrent use
//	func (f *Fake) Trash(id string) error // storage-side trash; ports.ErrFileNotFound for an unknown id
//	func (f *Fake) SetUnavailable(on bool)    // every port call → *domain.DomainError{Code: domain.CodeStorageUnavailable}
//	func (f *Fake) ReportWrongSHA256(on bool) // the storage reports a sha256 that differs from the bytes received
//	func (f *Fake) Len() int                  // number of stored files
//
// and in domain/errors.go: const CodeStorageUnavailable = "governance.storage_unavailable",
// used by its ErrorRegistry entry.

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/filesfake"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/filestest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

const codeStorageUnavailable = "governance.storage_unavailable"

var _ ports.Files = (*filesfake.Fake)(nil)

func TestFake_PassesTheFilesContract(t *testing.T) {
	filestest.Run(t, func(t *testing.T) filestest.Harness {
		f := filesfake.New()
		return filestest.Harness{
			Files: f,
			Trash: func(t *testing.T, id string) {
				t.Helper()
				if err := f.Trash(id); err != nil {
					t.Fatalf("Trash(%q): %v", id, err)
				}
			},
		}
	})
}

func TestFake_TrashUnknownIDIsNotFound(t *testing.T) {
	if err := filesfake.New().Trash("nope"); !errors.Is(err, ports.ErrFileNotFound) {
		t.Fatalf("Trash(unknown) error = %v, want ports.ErrFileNotFound", err)
	}
}

func assertStorageUnavailable(t *testing.T, op string, err error) {
	t.Helper()
	var de *domain.DomainError
	if !errors.As(err, &de) || de.Code != codeStorageUnavailable {
		t.Errorf("%s while unavailable: error = %v, want *domain.DomainError %s", op, err, codeStorageUnavailable)
	}
}

// Flow 8: the storage is down, every call answers storage_unavailable and
// nothing changes; when it is back, the stored file is as it was.
func TestFake_UnavailableStorageAnswersStorageUnavailableAndChangesNothing(t *testing.T) {
	ctx := context.Background()
	f := filesfake.New()
	stored, err := f.Upload(ctx, "a.pdf", "application/pdf", bytes.NewReader([]byte("signed scan")))
	if err != nil {
		t.Fatalf("Upload: %v", err)
	}
	if err := f.Trash(stored.ID); err != nil {
		t.Fatalf("Trash: %v", err)
	}

	f.SetUnavailable(true)
	_, errUp := f.Upload(ctx, "b.pdf", "application/pdf", bytes.NewReader([]byte("x")))
	assertStorageUnavailable(t, "Upload", errUp)
	_, errOpen := f.Open(ctx, stored.ID)
	assertStorageUnavailable(t, "Open", errOpen)
	_, errStat := f.Stat(ctx, stored.ID)
	assertStorageUnavailable(t, "Stat", errStat)
	assertStorageUnavailable(t, "SetReadOnly", f.SetReadOnly(ctx, stored.ID))
	assertStorageUnavailable(t, "Untrash", f.Untrash(ctx, stored.ID))

	f.SetUnavailable(false)
	if f.Len() != 1 {
		t.Fatalf("Len = %d after the outage, want 1 (the failed upload stored nothing)", f.Len())
	}
	st, err := f.Stat(ctx, stored.ID)
	if err != nil {
		t.Fatalf("Stat after the outage: %v", err)
	}
	if st.ReadOnly || !st.Trashed || st.StoredFile != stored {
		t.Fatalf("Stat after the outage = %+v, want %+v trashed, not read-only", st, stored)
	}
}

// ADR-0005: sha256 is computed while streaming and compared with the
// storage's; a mismatch fails the upload and leaves nothing behind.
func TestFake_WrongReportedSHA256FailsTheUploadAndStoresNothing(t *testing.T) {
	ctx := context.Background()
	f := filesfake.New()
	f.ReportWrongSHA256(true)
	got, err := f.Upload(ctx, "a.pdf", "application/pdf", bytes.NewReader([]byte("case document")))
	if !errors.Is(err, ports.ErrChecksumMismatch) {
		t.Fatalf("Upload error = %v, want ports.ErrChecksumMismatch", err)
	}
	if got != (ports.StoredFile{}) {
		t.Errorf("Upload returned %+v with the mismatch, want the zero StoredFile", got)
	}
	if f.Len() != 0 {
		t.Fatalf("Len = %d after a mismatch, want 0", f.Len())
	}

	f.ReportWrongSHA256(false)
	if _, err := f.Upload(ctx, "a.pdf", "application/pdf", bytes.NewReader([]byte("case document"))); err != nil {
		t.Fatalf("Upload after the fault is off: %v", err)
	}
	if f.Len() != 1 {
		t.Fatalf("Len = %d, want 1", f.Len())
	}
}

func TestStorageUnavailableCode_IsNamedAndRegisteredAs503(t *testing.T) {
	if domain.CodeStorageUnavailable != codeStorageUnavailable {
		t.Fatalf("CodeStorageUnavailable = %q, want %q", domain.CodeStorageUnavailable, codeStorageUnavailable)
	}
	for _, def := range domain.ErrorRegistry {
		if def.Code == codeStorageUnavailable {
			if def.Status != http.StatusServiceUnavailable {
				t.Fatalf("ErrorRegistry %s status = %d, want 503", def.Code, def.Status)
			}
			return
		}
	}
	t.Fatalf("%s is missing from ErrorRegistry", codeStorageUnavailable)
}

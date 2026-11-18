package governance_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// T92 (AC-26, AC-28, review N09): the file storage is resolved per association (board-protocol ADR-0005),
// so a file kept for one association is not readable through the storage of another.

// The development resolver keeps one in-memory storage per association.
func TestDevFilesKeepOneStoragePerAssociation(t *testing.T) {
	ctx := context.Background()
	a, b := uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())
	resolve := governance.NewDevFiles()

	storeA, err := resolve(ctx, a)
	if err != nil {
		t.Fatalf("resolve(A): %v", err)
	}
	stored, err := storeA.Upload(ctx, "contract.pdf", "application/pdf", bytes.NewReader([]byte("%PDF-1.7 A only")))
	if err != nil {
		t.Fatalf("upload into A: %v", err)
	}

	again, err := resolve(ctx, a)
	if err != nil {
		t.Fatalf("resolve(A) again: %v", err)
	}
	rc, err := again.Open(ctx, stored.ID)
	if err != nil {
		t.Fatalf("open A's file through A's storage resolved a second time: %v, want the same storage", err)
	}
	got, _ := io.ReadAll(rc)
	_ = rc.Close()
	if string(got) != "%PDF-1.7 A only" {
		t.Errorf("A's file through A = %q", got)
	}

	storeB, err := resolve(ctx, b)
	if err != nil {
		t.Fatalf("resolve(B): %v", err)
	}
	if rc, err := storeB.Open(ctx, stored.ID); !errors.Is(err, ports.ErrFileNotFound) {
		if rc != nil {
			_ = rc.Close()
		}
		t.Errorf("open A's storage id through B's storage: err = %v, want ports.ErrFileNotFound", err)
	}
}

// Download resolves the storage of the file's own association: B's storage does not hold A's bytes, and an
// association whose resolver yields no storage answers 503 governance.storage_unavailable.
func TestDownloadFileResolvesTheStoragePerAssociation(t *testing.T) {
	t.Run("A's bytes are not served through B's storage", func(t *testing.T) {
		var bStorage *driveFiles
		e := newDownloadEnvWith(t, func(e *downloadEnv) filesResolver {
			bStorage = &driveFiles{data: map[string][]byte{}}
			bid := uuid.MustParse(e.b)
			return func(_ context.Context, assoc uuid.UUID) (governance.Files, error) {
				if assoc == bid {
					return bStorage, nil
				}
				return e.files, nil
			}
		})
		e.requireServed(t, "A's scan through A's storage", e.fetch(t, e.a, e.signedA.ScanFileID), e.signedA.ScanFileName, e.scanBytes)
		// B's голова asks for B's file: the bytes sit only in the shared storage, so B's own (empty)
		// storage answers "not found". 200 would mean the shared storage was read, 403 that the request
		// never reached the resolver.
		got := e.fetchAs(t, "bhead", e.b, e.signedB.PrintFileID)
		if got.status != http.StatusNotFound {
			t.Errorf("download of B's file through B's empty storage: status = %d code = %q, want 404; body %q", got.status, got.code(), got.body)
		}
	})
	t.Run("no storage for the association", func(t *testing.T) {
		e := newDownloadEnvWith(t, func(e *downloadEnv) filesResolver {
			bid := uuid.MustParse(e.b)
			return func(_ context.Context, assoc uuid.UUID) (governance.Files, error) {
				if assoc == bid {
					return nil, nil
				}
				return e.files, nil
			}
		})
		got := e.fetchAs(t, "bhead", e.b, e.signedB.PrintFileID)
		if got.status != http.StatusServiceUnavailable || got.code() != "governance.storage_unavailable" {
			t.Errorf("download for an association without a storage: status = %d code = %q, want 503 governance.storage_unavailable; body %s", got.status, got.code(), got.body)
		}
		e.requireServed(t, "A's scan", e.fetch(t, e.a, e.signedA.ScanFileID), e.signedA.ScanFileName, e.scanBytes)
	})
}

// T127 (AC-26, review S07): files.drive_file_id is unique across associations, so the first upload of two
// associations under the development storage must not get the same storage id.
func TestDevFilesGiveFirstUploadsOfTwoAssociationsDistinctIDs(t *testing.T) {
	ctx := context.Background()
	resolve := governance.NewDevFiles()
	ids := map[string]uuid.UUID{}
	for _, assoc := range []uuid.UUID{uuid.Must(uuid.NewV7()), uuid.Must(uuid.NewV7())} {
		store, err := resolve(ctx, assoc)
		if err != nil {
			t.Fatalf("resolve(%s): %v", assoc, err)
		}
		stored, err := store.Upload(ctx, "contract.pdf", "application/pdf", bytes.NewReader([]byte("%PDF-1.7 first")))
		if err != nil {
			t.Fatalf("first upload of %s: %v", assoc, err)
		}
		if other, dup := ids[stored.ID]; dup {
			t.Fatalf("first upload of %s got storage id %q, already issued to %s: files.drive_file_id would collide", assoc, stored.ID, other)
		}
		ids[stored.ID] = assoc
	}
}

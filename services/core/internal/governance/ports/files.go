package ports

import (
	"context"
	"errors"
	"io"
)

// ErrFileNotFound means no stored file has the id.
var ErrFileNotFound = errors.New("governance: file not found")

// ErrChecksumMismatch means the storage reports a sha256 other than the one
// computed while streaming the upload (ADR-0005); nothing is kept.
var ErrChecksumMismatch = errors.New("governance: storage sha256 differs from the uploaded bytes")

// StoredFile is a file as the storage holds it after an upload.
type StoredFile struct {
	ID        string // storage file id, never empty
	Revision  string // storage revision id, never empty
	SHA256    string // lowercase hex of the uploaded bytes
	Size      int64
	MediaType string
}

// FileStatus is a stored file plus the state the storage reports for it.
type FileStatus struct {
	StoredFile
	Trashed  bool
	ReadOnly bool
}

// Files is the association-owned file storage (ADR-0005). An unavailable
// storage answers *domain.DomainError{Code: domain.CodeStorageUnavailable}
// from every method and changes nothing.
type Files interface {
	// Upload streams body to the storage, computing its sha256 on the way,
	// and fails with ErrChecksumMismatch when the storage reports another one.
	Upload(ctx context.Context, name, mediaType string, body io.Reader) (StoredFile, error)
	// Open streams back the stored bytes; the caller closes the reader.
	Open(ctx context.Context, id string) (io.ReadCloser, error)
	// Stat reports the file as the storage sees it now.
	Stat(ctx context.Context, id string) (FileStatus, error)
	// SetReadOnly forbids changing the file's content; idempotent.
	SetReadOnly(ctx context.Context, id string) error
	// Untrash restores a trashed file; a no-op on a file that is not trashed.
	Untrash(ctx context.Context, id string) error
}

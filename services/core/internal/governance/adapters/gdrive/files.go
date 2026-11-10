// Package gdrive is ports.Files on Google Drive: an association's files live
// in its own Drive folder (ADR-0005).
package gdrive

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// ReadOnlyReason is the reason Drive shows on a file spilno made read-only.
const ReadOnlyReason = "locked by spilno"

// Config is one association's Drive storage.
type Config struct {
	HTTP      *http.Client // authorized as the association's Google account
	APIURL    string       // "https://www.googleapis.com/drive/v3" in production; no trailing slash
	UploadURL string       // "https://www.googleapis.com/upload/drive/v3" in production; no trailing slash
	FolderID  string       // the association's spilno folder; every upload goes there
}

// Files is ports.Files on Google Drive.
type Files struct {
	cfg Config
}

var _ ports.Files = (*Files)(nil)

var errNotImplemented = errors.New("not implemented")

// New returns the storage of one association.
func New(cfg Config) *Files {
	return &Files{cfg: cfg}
}

// Upload streams body into the association's folder.
func (f *Files) Upload(ctx context.Context, name, mediaType string, body io.Reader) (ports.StoredFile, error) {
	return ports.StoredFile{}, errNotImplemented
}

// Open streams back the stored bytes; the caller closes the reader.
func (f *Files) Open(ctx context.Context, id string) (io.ReadCloser, error) {
	return nil, errNotImplemented
}

// Stat reports the file as Drive sees it now.
func (f *Files) Stat(ctx context.Context, id string) (ports.FileStatus, error) {
	return ports.FileStatus{}, errNotImplemented
}

// SetReadOnly forbids changing the file's content.
func (f *Files) SetReadOnly(ctx context.Context, id string) error {
	return errNotImplemented
}

// Untrash restores a trashed file.
func (f *Files) Untrash(ctx context.Context, id string) error {
	return errNotImplemented
}

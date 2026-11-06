// Package filesfake is an in-memory ports.Files for use case and worker
// tests, with switchable storage faults. It passes the filestest contract.
package filesfake

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"sync"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

type file struct {
	status ports.FileStatus
	data   []byte
}

// Fake is an in-memory file storage, safe for concurrent use.
type Fake struct {
	mu          sync.Mutex
	files       map[string]*file
	seq         int
	unavailable bool
	wrongSHA256 bool
}

var _ ports.Files = (*Fake)(nil)

// New returns an empty storage.
func New() *Fake {
	return &Fake{files: map[string]*file{}}
}

// SetUnavailable switches the outage fault: while on, every port call answers
// governance.storage_unavailable and changes nothing.
func (f *Fake) SetUnavailable(on bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.unavailable = on
}

// ReportWrongSHA256 switches the checksum fault: while on, the storage
// reports a sha256 that differs from the bytes it received.
func (f *Fake) ReportWrongSHA256(on bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.wrongSHA256 = on
}

// Len returns the number of stored files.
func (f *Fake) Len() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.files)
}

// Trash moves a file to the trash, as a person in the storage's own UI would.
func (f *Fake) Trash(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fl, ok := f.files[id]
	if !ok {
		return notFound(id)
	}
	fl.status.Trashed = true
	return nil
}

// Upload implements ports.Files.
func (f *Fake) Upload(ctx context.Context, _, mediaType string, body io.Reader) (ports.StoredFile, error) {
	if err := f.checkAvailable(); err != nil {
		return ports.StoredFile{}, err
	}
	if err := ctx.Err(); err != nil {
		return ports.StoredFile{}, err
	}
	h := sha256.New()
	var buf bytes.Buffer
	if _, err := io.Copy(io.MultiWriter(&buf, h), body); err != nil {
		return ports.StoredFile{}, fmt.Errorf("filesfake: read upload body: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return ports.StoredFile{}, err
	}
	computed := hex.EncodeToString(h.Sum(nil))

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unavailable {
		return ports.StoredFile{}, unavailable()
	}
	reported := computed
	if f.wrongSHA256 {
		reported = flipFirstHexDigit(computed)
	}
	if reported != computed {
		return ports.StoredFile{}, ports.ErrChecksumMismatch
	}
	f.seq++
	stored := ports.StoredFile{
		ID:        "fake-file-" + strconv.Itoa(f.seq),
		Revision:  "fake-rev-" + strconv.Itoa(f.seq),
		SHA256:    reported,
		Size:      int64(buf.Len()),
		MediaType: mediaType,
	}
	f.files[stored.ID] = &file{status: ports.FileStatus{StoredFile: stored}, data: buf.Bytes()}
	return stored, nil
}

// Open implements ports.Files.
func (f *Fake) Open(_ context.Context, id string) (io.ReadCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fl, err := f.lookup(id)
	if err != nil {
		return nil, err
	}
	return io.NopCloser(bytes.NewReader(fl.data)), nil
}

// Stat implements ports.Files.
func (f *Fake) Stat(_ context.Context, id string) (ports.FileStatus, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	fl, err := f.lookup(id)
	if err != nil {
		return ports.FileStatus{}, err
	}
	return fl.status, nil
}

// SetReadOnly implements ports.Files.
func (f *Fake) SetReadOnly(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fl, err := f.lookup(id)
	if err != nil {
		return err
	}
	fl.status.ReadOnly = true
	return nil
}

// Untrash implements ports.Files.
func (f *Fake) Untrash(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fl, err := f.lookup(id)
	if err != nil {
		return err
	}
	fl.status.Trashed = false
	return nil
}

func (f *Fake) checkAvailable() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.unavailable {
		return unavailable()
	}
	return nil
}

// lookup must be called with f.mu held.
func (f *Fake) lookup(id string) (*file, error) {
	if f.unavailable {
		return nil, unavailable()
	}
	fl, ok := f.files[id]
	if !ok {
		return nil, notFound(id)
	}
	return fl, nil
}

func unavailable() error {
	return &domain.DomainError{Code: domain.CodeStorageUnavailable}
}

func notFound(id string) error {
	return fmt.Errorf("filesfake: %q: %w", id, ports.ErrFileNotFound)
}

func flipFirstHexDigit(s string) string {
	if s[0] == '0' {
		return "1" + s[1:]
	}
	return "0" + s[1:]
}

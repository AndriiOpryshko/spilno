package app

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// ErrFileNotFound is returned when the association has no file with the id.
var ErrFileNotFound = errors.New("file not found")

// ErrFileNotSigned is returned when a read-only association asks for a file no signed snapshot pins.
var ErrFileNotSigned = errors.New("file is not pinned by a signed protocol")

// FileService serves stored files; every call runs on the transaction in ctx.
type FileService struct {
	catalog ports.FileCatalog
	files   ports.FilesResolver // nil until the storage is wired
}

func NewFileService(catalog ports.FileCatalog) *FileService { return &FileService{catalog: catalog} }

// SetFiles wires the storage; call it before serving.
func (s *FileService) SetFiles(files ports.FilesResolver) { s.files = files }

// HasFiles reports whether the association has at least one stored file, on the transaction in ctx.
func (s *FileService) HasFiles(ctx context.Context, association uuid.UUID) (bool, error) {
	return s.catalog.Any(ctx, association)
}

// Download is a file whose bytes were read back and matched its recorded sha256.
type Download struct {
	Name, MediaType string
	Body            io.Reader
	Size            int64
}

// Download returns the file's bytes after checking their sha256. With readOnly set only a file a signed
// snapshot pins is served (ErrFileNotSigned otherwise, nothing streamed). An unwired or unreachable
// storage is *domain.DomainError{Code: CodeStorageUnavailable}.
func (s *FileService) Download(ctx context.Context, association, id uuid.UUID, readOnly bool) (Download, error) {
	// The record and the pin are read in a short transaction; the bytes are read after it closes.
	var rec ports.FileRecord
	err := db.InScope(ctx, func(ctx context.Context) error {
		var e error
		rec, e = s.catalog.Find(ctx, association, id)
		if errors.Is(e, ports.ErrFileNotFound) {
			return ErrFileNotFound
		}
		if e != nil {
			return e
		}
		if readOnly && !rec.Pinned {
			return ErrFileNotSigned
		}
		return nil
	})
	if err != nil {
		return Download{}, err
	}
	store, err := resolveFiles(ctx, s.files, association)
	if err != nil {
		return Download{}, err
	}
	rc, err := store.Open(ctx, rec.DriveFileID)
	if errors.Is(err, ports.ErrFileNotFound) {
		return Download{}, ErrFileNotFound
	}
	if err != nil {
		return Download{}, err
	}
	defer rc.Close()
	// The bytes are held back until their digest matches, so a replaced file is never sent.
	data, err := io.ReadAll(io.LimitReader(rc, rec.Size+1))
	if err != nil {
		if ctx.Err() != nil {
			return Download{}, err
		}
		return Download{}, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != rec.SHA256 {
		return Download{}, fmt.Errorf("file %s: stored bytes differ from the recorded sha256", id)
	}
	return Download{Name: rec.Name, MediaType: rec.MediaType, Body: bytes.NewReader(data), Size: int64(len(data))}, nil
}

// resolveFiles is the association's storage, or storage_unavailable when none is wired or connected.
func resolveFiles(ctx context.Context, resolve ports.FilesResolver, association uuid.UUID) (ports.Files, error) {
	if resolve == nil {
		return nil, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	store, err := resolve(ctx, association)
	if err != nil || store == nil {
		// A resolver failure (an unreadable token, a database error) is a storage outage to the caller.
		return nil, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	return store, nil
}

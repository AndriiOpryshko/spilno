package ports

import (
	"context"

	"github.com/google/uuid"
)

// FileRecord is the stored metadata of one file as the download needs it.
type FileRecord struct {
	DriveFileID string
	SHA256      string // lowercase hex
	Size        int64
	MediaType   string
	Name        string
	// Pinned reports that a signed protocol snapshot references the file.
	Pinned bool
}

// FileCatalog reads file metadata on the transaction in ctx.
type FileCatalog interface {
	// Find returns ErrFileNotFound when the association has no file with the id.
	Find(ctx context.Context, association, id uuid.UUID) (FileRecord, error)
}

package postgres

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// Files reads file metadata on the transaction in ctx.
type Files struct{}

var _ ports.FileCatalog = Files{}

func NewFiles() Files { return Files{} }

func (Files) Find(ctx context.Context, association, id uuid.UUID) (ports.FileRecord, error) {
	q, err := queries(ctx)
	if err != nil {
		return ports.FileRecord{}, err
	}
	r, err := q.GetFileForDownload(ctx, sqlcgen.GetFileForDownloadParams{AssociationID: uuidOf(association), ID: uuidOf(id)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.FileRecord{}, ports.ErrFileNotFound
	}
	if err != nil {
		return ports.FileRecord{}, err
	}
	return ports.FileRecord{
		DriveFileID: r.DriveFileID, SHA256: hex.EncodeToString(r.Sha256), Size: r.SizeBytes,
		MediaType: r.MediaType, Name: r.OriginalName, Pinned: r.Pinned,
	}, nil
}

package postgres

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
)

// Cases stores справи and their documents on the transaction in ctx.
type Cases struct{}

var _ ports.CaseRepository = Cases{}

func NewCases() Cases { return Cases{} }

func caseOf(id, by pgtype.UUID, title string, at pgtype.Timestamptz) ports.CaseRecord {
	return ports.CaseRecord{ID: uuid.UUID(id.Bytes), Title: title, CreatedByMemberID: uuid.UUID(by.Bytes), CreatedAt: at.Time}
}

func (Cases) Create(ctx context.Context, association, id uuid.UUID, title string, by uuid.UUID) (ports.CaseRecord, error) {
	q, err := queries(ctx)
	if err != nil {
		return ports.CaseRecord{}, err
	}
	r, err := q.InsertCase(ctx, sqlcgen.InsertCaseParams{
		ID: uuidOf(id), AssociationID: uuidOf(association), Title: title, CreatedByMemberID: uuidOf(by),
	})
	if err != nil {
		return ports.CaseRecord{}, err
	}
	return caseOf(r.ID, r.CreatedByMemberID, r.Title, r.CreatedAt), nil
}

func (Cases) List(ctx context.Context, association uuid.UUID, after *ports.MeetingKey, limit int) ([]ports.CaseRecord, error) {
	q, err := queries(ctx)
	if err != nil {
		return nil, err
	}
	// limit is bounded by app.MaxPageSize, so it fits int32.
	arg := sqlcgen.ListCasesParams{AssociationID: uuidOf(association), RowLimit: int32(limit)}
	if after != nil {
		arg.AfterCreatedAt = pgtype.Timestamptz{Time: after.CreatedAt, Valid: true}
		arg.AfterID = uuidOf(after.ID)
	}
	rows, err := q.ListCases(ctx, arg)
	if err != nil {
		return nil, err
	}
	out := make([]ports.CaseRecord, len(rows))
	for i, r := range rows {
		out[i] = caseOf(r.ID, r.CreatedByMemberID, r.Title, r.CreatedAt)
	}
	return out, nil
}

func (Cases) Get(ctx context.Context, association, id uuid.UUID) (ports.CaseRecord, bool, error) {
	q, err := queries(ctx)
	if err != nil {
		return ports.CaseRecord{}, false, err
	}
	r, err := q.GetCase(ctx, sqlcgen.GetCaseParams{AssociationID: uuidOf(association), ID: uuidOf(id)})
	if errors.Is(err, pgx.ErrNoRows) {
		return ports.CaseRecord{}, false, nil
	}
	if err != nil {
		return ports.CaseRecord{}, false, err
	}
	return caseOf(r.ID, r.CreatedByMemberID, r.Title, r.CreatedAt), true, nil
}

func (Cases) Documents(ctx context.Context, association, caseID uuid.UUID) ([]ports.CaseDocumentRecord, error) {
	q, err := queries(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := q.ListCaseDocuments(ctx, sqlcgen.ListCaseDocumentsParams{AssociationID: uuidOf(association), CaseID: uuidOf(caseID)})
	if err != nil {
		return nil, err
	}
	out := make([]ports.CaseDocumentRecord, len(rows))
	for i, r := range rows {
		out[i] = documentOf(r)
	}
	return out, nil
}

func (Cases) HasDocument(ctx context.Context, association, caseID, document uuid.UUID) (bool, error) {
	q, err := queries(ctx)
	if err != nil {
		return false, err
	}
	return q.CaseHasDocument(ctx, sqlcgen.CaseHasDocumentParams{
		AssociationID: uuidOf(association), CaseID: uuidOf(caseID), ID: uuidOf(document),
	})
}

func (Cases) AddDocument(ctx context.Context, association uuid.UUID, d ports.NewCaseDocument) (ports.CaseDocumentRecord, error) {
	q, err := queries(ctx)
	if err != nil {
		return ports.CaseDocumentRecord{}, err
	}
	sum, err := hex.DecodeString(d.SHA256)
	if err != nil {
		return ports.CaseDocumentRecord{}, fmt.Errorf("sha256 %q: %w", d.SHA256, err)
	}
	if err := q.InsertFile(ctx, sqlcgen.InsertFileParams{
		ID: uuidOf(d.FileID), AssociationID: uuidOf(association),
		DriveFileID: d.DriveFileID, DriveRevisionID: d.DriveRevisionID,
		Sha256: sum, SizeBytes: d.Size, MediaType: d.MediaType, OriginalName: d.Name,
		UploadedByMemberID: uuidOf(d.ByMemberID),
	}); err != nil {
		return ports.CaseDocumentRecord{}, err
	}
	doc := sqlcgen.InsertCaseDocumentParams{
		ID: uuidOf(d.ID), AssociationID: uuidOf(association), CaseID: uuidOf(d.CaseID), FileID: uuidOf(d.FileID),
		DocumentDate: optDateOf(d.DocumentDate), CreatedByMemberID: uuidOf(d.ByMemberID),
	}
	if d.PreviousVersionID != nil {
		doc.PreviousVersionID = uuidOf(*d.PreviousVersionID)
	}
	if err := q.InsertCaseDocument(ctx, doc); err != nil {
		return ports.CaseDocumentRecord{}, err
	}
	r, err := q.GetCaseDocument(ctx, sqlcgen.GetCaseDocumentParams{AssociationID: uuidOf(association), ID: uuidOf(d.ID)})
	if err != nil {
		return ports.CaseDocumentRecord{}, err
	}
	return documentOf(sqlcgen.ListCaseDocumentsRow(r)), nil
}

func documentOf(r sqlcgen.ListCaseDocumentsRow) ports.CaseDocumentRecord {
	out := ports.CaseDocumentRecord{
		ID: uuid.UUID(r.ID.Bytes), CaseID: uuid.UUID(r.CaseID.Bytes),
		File: ports.StoredDocumentFile{
			ID: uuid.UUID(r.FileID.Bytes), Name: r.OriginalName, MediaType: r.MediaType, Size: r.SizeBytes,
			SHA256: hex.EncodeToString(r.Sha256), CreatedAt: r.FileCreatedAt.Time,
		},
		CreatedByMemberID: uuid.UUID(r.CreatedByMemberID.Bytes), CreatedAt: r.CreatedAt.Time,
	}
	if r.UploadedByMemberID.Valid {
		id := uuid.UUID(r.UploadedByMemberID.Bytes)
		out.File.UploadedByMemberID = &id
	}
	if r.PreviousVersionID.Valid {
		id := uuid.UUID(r.PreviousVersionID.Bytes)
		out.PreviousVersionID = &id
	}
	if r.DocumentDate.Valid {
		d := r.DocumentDate.Time
		out.DocumentDate = &d
	}
	return out
}

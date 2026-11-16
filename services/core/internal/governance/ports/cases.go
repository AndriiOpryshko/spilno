package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// CaseRecord is one справа.
type CaseRecord struct {
	ID                uuid.UUID
	Title             string
	CreatedByMemberID uuid.UUID
	CreatedAt         time.Time
}

// StoredDocumentFile is the files row of a document as the API shows it; it holds no storage id.
type StoredDocumentFile struct {
	ID                 uuid.UUID
	Name, MediaType    string
	Size               int64
	SHA256             string // lowercase hex
	UploadedByMemberID *uuid.UUID
	CreatedAt          time.Time
}

// CaseDocumentRecord is one version of a document of a справа.
type CaseDocumentRecord struct {
	ID, CaseID        uuid.UUID
	File              StoredDocumentFile
	PreviousVersionID *uuid.UUID
	DocumentDate      *time.Time // a calendar date
	CreatedByMemberID uuid.UUID
	CreatedAt         time.Time
}

// NewCaseDocument is a document version to store: its files row (with the storage ids) and its case_documents row.
type NewCaseDocument struct {
	ID, FileID, CaseID uuid.UUID
	PreviousVersionID  *uuid.UUID
	DocumentDate       *time.Time
	ByMemberID         uuid.UUID
	DriveFileID        string
	DriveRevisionID    string
	SHA256             string // lowercase hex
	Size               int64
	MediaType, Name    string
}

// CaseRepository stores справи and their documents on the transaction in ctx (db.InAssociation).
type CaseRepository interface {
	Create(ctx context.Context, association, id uuid.UUID, title string, by uuid.UUID) (CaseRecord, error)
	// List returns up to limit справи, newest first, strictly after the key (from the top when nil).
	List(ctx context.Context, association uuid.UUID, after *MeetingKey, limit int) ([]CaseRecord, error)
	// Get returns false when the association has no справа with the id.
	Get(ctx context.Context, association, id uuid.UUID) (CaseRecord, bool, error)
	// Documents returns every version of the справа's documents, oldest first.
	Documents(ctx context.Context, association, caseID uuid.UUID) ([]CaseDocumentRecord, error)
	// HasDocument reports whether the document belongs to the справа.
	HasDocument(ctx context.Context, association, caseID, document uuid.UUID) (bool, error)
	// AddDocument stores the files row and the document row and returns the document.
	AddDocument(ctx context.Context, association uuid.UUID, d NewCaseDocument) (CaseDocumentRecord, error)
}

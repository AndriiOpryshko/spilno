package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/kyivday"
)

// maxFileNameLen is the width of files.original_name.
const maxFileNameLen = 255

// CaseService serves справи and the document versions in them; every call runs on the transaction in ctx.
type CaseService struct {
	repo   ports.CaseRepository
	roster ports.RosterWriter
	files  ports.FilesResolver // nil until the storage is wired
	now    func() time.Time
}

func NewCaseService(repo ports.CaseRepository, roster ports.RosterWriter) *CaseService {
	return &CaseService{repo: repo, roster: roster, now: time.Now}
}

// SetFiles wires the storage; call it before serving.
func (s *CaseService) SetFiles(files ports.FilesResolver) { s.files = files }

// CasePage is one slice of the newest-first справи listing.
type CasePage struct {
	Items []ports.CaseRecord
	// NextCursor resumes the listing after Items; empty on the last page.
	NextCursor string
}

// CaseWithDocuments is a справа and all its document versions.
type CaseWithDocuments struct {
	ports.CaseRecord
	Documents []ports.CaseDocumentRecord
}

// Create opens a справа for the signed-in член правління. A bad title is *domain.ValidationError.
func (s *CaseService) Create(ctx context.Context, association uuid.UUID, title string) (ports.CaseRecord, error) {
	if err := domain.ValidateCaseTitle(title); err != nil {
		return ports.CaseRecord{}, err
	}
	by, err := s.member(ctx, association)
	if err != nil {
		return ports.CaseRecord{}, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return ports.CaseRecord{}, fmt.Errorf("case id: %w", err)
	}
	return s.repo.Create(ctx, association, id, title, by)
}

// List returns up to limit справи after cursor, newest first, like the meeting listing.
func (s *CaseService) List(ctx context.Context, association uuid.UUID, limit int, cursor string) (CasePage, error) {
	switch {
	case limit <= 0:
		limit = DefaultPageSize
	case limit > MaxPageSize:
		limit = MaxPageSize
	}
	var after *ports.MeetingKey
	if cursor != "" {
		key, err := decodeCursor(cursor)
		if err != nil {
			return CasePage{}, err
		}
		after = &key
	}
	// One extra row tells whether another page follows.
	items, err := s.repo.List(ctx, association, after, limit+1)
	if err != nil {
		return CasePage{}, err
	}
	page := CasePage{Items: items}
	if len(items) > limit {
		page.Items = items[:limit]
		last := page.Items[limit-1]
		page.NextCursor = encodeCursor(ports.MeetingKey{CreatedAt: last.CreatedAt, ID: last.ID})
	}
	return page, nil
}

// Get returns the справа with its documents, or domain.ErrCaseNotFound.
func (s *CaseService) Get(ctx context.Context, association, id uuid.UUID) (CaseWithDocuments, error) {
	c, ok, err := s.repo.Get(ctx, association, id)
	if err != nil {
		return CaseWithDocuments{}, err
	}
	if !ok {
		return CaseWithDocuments{}, domain.ErrCaseNotFound
	}
	docs, err := s.repo.Documents(ctx, association, id)
	if err != nil {
		return CaseWithDocuments{}, err
	}
	return CaseWithDocuments{CaseRecord: c, Documents: docs}, nil
}

// Upload is a document as the caller sent it.
type Upload struct {
	Name              string
	Data              []byte
	PreviousVersionID *uuid.UUID
	DocumentDate      *time.Time
}

// AddDocument stores the upload as a new version in the справа, after the storage took the bytes.
// An unknown справа is domain.ErrCaseNotFound, a previous version outside the справа
// domain.ErrPreviousVersionInvalid, a file over the limit or not PDF/JPEG/PNG by content
// *domain.CodedFieldsError{Code: file_invalid}, an unwired or unreachable storage
// *domain.DomainError{Code: storage_unavailable}. Everything is checked before the upload, so a refusal
// leaves nothing in the storage.
func (s *CaseService) AddDocument(ctx context.Context, association, caseID uuid.UUID, up Upload) (ports.CaseDocumentRecord, error) {
	var (
		by        uuid.UUID
		mediaType string
		store     ports.Files
	)
	err := db.InScope(ctx, func(ctx context.Context) error {
		var err error
		if by, err = s.member(ctx, association); err != nil {
			return err
		}
		if _, ok, err := s.repo.Get(ctx, association, caseID); err != nil {
			return err
		} else if !ok {
			return domain.ErrCaseNotFound
		}
		if up.PreviousVersionID != nil {
			ok, err := s.repo.HasDocument(ctx, association, caseID, *up.PreviousVersionID)
			if err != nil {
				return err
			}
			if !ok {
				return domain.ErrPreviousVersionInvalid
			}
		}
		if mediaType, err = checkUpload(up); err != nil {
			return err
		}
		store, err = resolveFiles(ctx, s.files, association)
		return err
	})
	if err != nil {
		return ports.CaseDocumentRecord{}, err
	}
	stored, err := store.Upload(ctx, up.Name, mediaType, bytes.NewReader(up.Data))
	if errors.Is(err, ports.ErrChecksumMismatch) {
		return ports.CaseDocumentRecord{}, &domain.DomainError{Code: domain.CodeStorageUnavailable}
	}
	if err != nil {
		return ports.CaseDocumentRecord{}, err
	}
	doc, err := s.recordDocument(ctx, association, caseID, by, mediaType, stored, up)
	if err != nil {
		s.removeOrphan(ctx, store, stored.ID)
		return ports.CaseDocumentRecord{}, err
	}
	return doc, nil
}

// recordDocument writes the files row and the document row in one transaction.
func (s *CaseService) recordDocument(ctx context.Context, association, caseID, by uuid.UUID, mediaType string, stored ports.StoredFile, up Upload) (doc ports.CaseDocumentRecord, err error) {
	docID, err := uuid.NewV7()
	if err != nil {
		return doc, fmt.Errorf("document id: %w", err)
	}
	fileID, err := uuid.NewV7()
	if err != nil {
		return doc, fmt.Errorf("file id: %w", err)
	}
	err = db.InScope(ctx, func(ctx context.Context) (e error) {
		doc, e = s.repo.AddDocument(ctx, association, ports.NewCaseDocument{
			ID: docID, FileID: fileID, CaseID: caseID,
			PreviousVersionID: up.PreviousVersionID, DocumentDate: up.DocumentDate, ByMemberID: by,
			DriveFileID: stored.ID, DriveRevisionID: stored.Revision,
			SHA256: stored.SHA256, Size: stored.Size, MediaType: mediaType, Name: up.Name,
		})
		return e
	})
	return doc, err
}

// orphanRemoveTimeout bounds the cleanup of an uploaded file after a failed write.
const orphanRemoveTimeout = 30 * time.Second

// removeOrphan deletes an uploaded file whose record was not written, best effort; a failure is logged
// with the storage id only.
func (s *CaseService) removeOrphan(ctx context.Context, store ports.Files, id string) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), orphanRemoveTimeout)
	defer cancel()
	if err := store.Remove(ctx, id); err != nil {
		slog.ErrorContext(ctx, "remove the orphan uploaded file failed", "storage_file_id", id, "error", err)
	}
}

// checkUpload returns the media type of the content, or file_invalid naming /file.
func checkUpload(up Upload) (string, error) {
	invalid := func(detail string) error {
		return &domain.CodedFieldsError{Code: domain.CodeFileInvalid, Fields: []domain.FieldIssue{{Pointer: "/file", Detail: detail}}}
	}
	switch {
	case len(up.Data) == 0:
		return "", invalid("is empty")
	case len(up.Data) > domain.MaxDocumentSize:
		return "", invalid("is over 20 MB")
	case up.Name == "" || utf8.RuneCountInString(up.Name) > maxFileNameLen:
		return "", invalid("needs a file name of 1 to 255 characters")
	}
	mediaType, ok := domain.SniffDocumentMediaType(up.Data)
	if !ok {
		return "", invalid("must be a PDF, JPEG or PNG")
	}
	return mediaType, nil
}

// member is the roster entry of the signed-in член правління.
func (s *CaseService) member(ctx context.Context, association uuid.UUID) (uuid.UUID, error) {
	id, ok, err := s.roster.ActorEntryID(ctx, association, kyivday.Today(s.now()))
	if err != nil {
		return uuid.Nil, err
	}
	if !ok {
		return uuid.Nil, errors.New("the signed-in person has no current roster entry")
	}
	return id, nil
}

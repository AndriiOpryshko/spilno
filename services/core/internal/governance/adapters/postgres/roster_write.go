package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/adapters/postgres/sqlcgen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/governance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
)

// WriteRevision appends a revision of a roster entry on the transaction in ctx. The entry
// is validated in Go first: a missing or unknown position never reaches the database.
func (Roster) WriteRevision(ctx context.Context, association, boardMember uuid.UUID, entry domain.BoardMemberEntry) error {
	if err := validateForWrite(entry); err != nil {
		return err
	}
	var decision pgtype.UUID
	if entry.DecisionID != "" {
		id, err := uuid.Parse(entry.DecisionID)
		if err != nil {
			return &domain.CodedFieldsError{Code: domain.CodeRosterEntryInvalid, Fields: []domain.FieldIssue{{Pointer: "/decision_id", Detail: "must be a uuid"}}}
		}
		decision = uuidOf(id)
	}
	tx, err := db.From(ctx)
	if err != nil {
		return err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("revision id: %w", err)
	}
	params := sqlcgen.InsertBoardMemberRevisionParams{
		ID: uuidOf(id), AssociationID: uuidOf(association), BoardMemberID: uuidOf(boardMember),
		FullName:    entry.FullName,
		SignInEmail: pgtype.Text{String: entry.SignInEmail, Valid: entry.SignInEmail != ""},
		TermStart:   dateOf(entry.Term.Start),
		Position:    string(entry.Position),
		DecisionID:  decision,
	}
	if !entry.Term.End.IsZero() {
		params.TermEnd = dateOf(entry.Term.End)
	}
	return sqlcgen.New(tx).InsertBoardMemberRevision(ctx, params)
}

// validateForWrite is Validate plus the rule that a stored revision always has a position.
func validateForWrite(entry domain.BoardMemberEntry) error {
	err := entry.Validate()
	if entry.Position != "" {
		return err
	}
	var issues []domain.FieldIssue
	var coded *domain.CodedFieldsError
	if errors.As(err, &coded) {
		issues = append(issues, coded.Fields...)
	}
	issues = append(issues, domain.FieldIssue{Pointer: "/position", Detail: "is required"})
	return &domain.CodedFieldsError{Code: domain.CodeRosterEntryInvalid, Fields: issues}
}

var _ ports.RosterWriter = Roster{}

// CurrentMemberID finds the entry whose latest revision carries email and covers date.
func (Roster) CurrentMemberID(ctx context.Context, association uuid.UUID, email string, date time.Time) (uuid.UUID, bool, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return uuid.Nil, false, err
	}
	id, err := sqlcgen.New(tx).CurrentMemberIDByEmail(ctx, sqlcgen.CurrentMemberIDByEmailParams{
		AssociationID: uuidOf(association), Email: email, OnDate: dateOf(date),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	return uuid.UUID(id.Bytes), true, nil
}

// LatestEntry returns the newest revision of an entry.
func (Roster) LatestEntry(ctx context.Context, association, boardMember uuid.UUID) (domain.BoardMemberEntry, bool, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return domain.BoardMemberEntry{}, false, err
	}
	row, err := sqlcgen.New(tx).LatestRevisionOfMember(ctx, sqlcgen.LatestRevisionOfMemberParams{
		AssociationID: uuidOf(association), BoardMemberID: uuidOf(boardMember),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.BoardMemberEntry{}, false, nil
	}
	if err != nil {
		return domain.BoardMemberEntry{}, false, err
	}
	entry := domain.BoardMemberEntry{
		FullName: row.FullName, SignInEmail: row.SignInEmail,
		Term:     domain.Term{Start: row.TermStart.Time},
		Position: domain.Position(row.Position),
	}
	if row.TermEnd.Valid {
		entry.Term.End = row.TermEnd.Time
	}
	return entry, true, nil
}

// CreateMember stores a new, empty roster entry.
func (Roster) CreateMember(ctx context.Context, association uuid.UUID) (uuid.UUID, error) {
	tx, err := db.From(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("member id: %w", err)
	}
	if err := sqlcgen.New(tx).InsertBoardMember(ctx, sqlcgen.InsertBoardMemberParams{ID: uuidOf(id), AssociationID: uuidOf(association)}); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

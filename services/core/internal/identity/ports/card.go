package ports

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrCardNotFound is the one answer for a missing association or extract of it.
var ErrCardNotFound = errors.New("identity: association or extract not found")

// AssociationBrief is the name and state of an association, as listMemberships shows them.
type AssociationBrief struct {
	Name  string
	State string
}

// CardEvent is one step of the confirmation chain: an approval or a state change.
type CardEvent struct {
	Kind           string // connected | transfer_approved | disabled | enabled
	OccurredAt     time.Time
	DecisionID     *uuid.UUID
	FiledAt        *time.Time
	Reason         *string
	HeadFullName   *string
	SignerFullName *string
	ExtractID      *uuid.UUID
	ExtractDatedOn *time.Time
}

// Card is the association card, history newest first.
type Card struct {
	ID              uuid.UUID
	Name            string
	State           string
	Edrpou          *string
	Address         *string
	HeadFullName    *string
	ConnectedByHand bool
	History         []CardEvent
}

// CardRepository reads the association card; Card and Extract run on the transaction in ctx.
type CardRepository interface {
	Briefs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]AssociationBrief, error)
	Card(ctx context.Context, id uuid.UUID, today time.Time) (Card, error)
	// Extract returns ErrCardNotFound unless an approval of the association relied on the extract.
	Extract(ctx context.Context, association, extract uuid.UUID) (StoredExtract, error)
}

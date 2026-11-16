package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RegistryFilter narrows a registry search; every nil field is off and the set ones combine with AND.
type RegistryFilter struct {
	Counterparty, Subject        *string // case-insensitive "contains"
	Amount, AmountFrom, AmountTo *int64  // kopiyky; the range bounds are inclusive
	DateFrom, DateTo             *time.Time
}

// RegistryKey is a decision's position in the newest-meeting-first order.
type RegistryKey struct {
	MeetingDate time.Time // a calendar date
	ID          uuid.UUID // the registry entry
}

// RegistryDecision is one registry_entries row: a decision of a signed protocol.
type RegistryDecision struct {
	Key           RegistryKey
	ProtocolID    uuid.UUID
	AgendaItemID  uuid.UUID
	Number        int
	Subject       string
	Counterparty  *string
	AmountKopiyky *int64
	Currency      *string
}

// SignedSnapshot is the signed protocol_snapshots row of one protocol.
type SignedSnapshot struct {
	ProtocolID  uuid.UUID
	Number      int
	MeetingDate time.Time // a calendar date
	// Payload is the snapshot document (ADR-0007); it still holds storage ids that never leave the API.
	Payload []byte
}

// RegistryReader reads the реєстр протоколів on the transaction in ctx (db.InAssociation).
type RegistryReader interface {
	// Search returns up to limit decisions matching the filter, newest meeting first, strictly after
	// the key (from the top when nil).
	Search(ctx context.Context, association uuid.UUID, f RegistryFilter, after *RegistryKey, limit int) ([]RegistryDecision, error)
	// WithdrawnNumbers returns the numbers held by withdrawn protocols from lo through hi, ascending.
	WithdrawnNumbers(ctx context.Context, association uuid.UUID, lo, hi int) ([]int, error)
	// SignedSnapshot returns the signed snapshot of the protocol; false when it has none (unknown or withdrawn).
	SignedSnapshot(ctx context.Context, association, protocol uuid.UUID) (SignedSnapshot, bool, error)
}

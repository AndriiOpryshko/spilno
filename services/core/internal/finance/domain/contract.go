// Package domain holds the finance context's model. It imports only the standard library and uuid.
package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ErrContractNotFound is returned when no contract has the requested ID.
var ErrContractNotFound = errors.New("contract not found")

type Currency string

const (
	CurrencyUAH Currency = "UAH"
	CurrencyEUR Currency = "EUR"
	CurrencyUSD Currency = "USD"
)

func (c Currency) Valid() bool {
	switch c {
	case CurrencyUAH, CurrencyEUR, CurrencyUSD:
		return true
	}
	return false
}

const MaxCounterpartyLen = 200

// Field names reported in a ValidationError.
const (
	FieldCounterparty = "counterparty"
	FieldAmountCents  = "amount_cents"
	FieldCurrency     = "currency"
	FieldEndsAt       = "ends_at"
)

// ContractTerms is what the association agrees with a counterparty. SignedAt
// and EndsAt are calendar dates; only their year, month and day matter.
type ContractTerms struct {
	Counterparty string
	AmountCents  int64
	Currency     Currency
	SignedAt     time.Time
	// EndsAt is the last day of the contract; nil means open-ended.
	EndsAt *time.Time
}

// Validate checks every invariant and reports all broken ones at once as a
// *ValidationError.
func (t ContractTerms) Validate() error {
	var v ValidationError
	if n := utf8.RuneCountInString(t.Counterparty); n < 1 || n > MaxCounterpartyLen {
		v.add(FieldCounterparty, fmt.Sprintf("must be 1 to %d characters", MaxCounterpartyLen))
	}
	if t.AmountCents <= 0 {
		v.add(FieldAmountCents, "must be greater than 0")
	}
	if !t.Currency.Valid() {
		v.add(FieldCurrency, "must be one of UAH, EUR, USD")
	}
	if t.EndsAt != nil && dateOf(*t.EndsAt).Before(dateOf(t.SignedAt)) {
		v.add(FieldEndsAt, "must be on or after signed_at")
	}
	if len(v.Fields) > 0 {
		return &v
	}
	return nil
}

// Contract is a stored contract: its terms plus the identity and timestamps
// the store assigns.
type Contract struct {
	ID uuid.UUID
	ContractTerms
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FieldError struct {
	Field  string
	Detail string
}

// ValidationError lists the invariants a ContractTerms breaks.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Field + ": " + f.Detail
	}
	return "invalid contract: " + strings.Join(parts, "; ")
}

func (e *ValidationError) add(field, detail string) {
	e.Fields = append(e.Fields, FieldError{Field: field, Detail: detail})
}

func dateOf(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

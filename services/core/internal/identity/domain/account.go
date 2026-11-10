package domain

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

// ErrSignInRefused is returned when the provider does not vouch for the address
// or the sign-in state is unknown, expired or already used.
var ErrSignInRefused = errors.New("identity: sign-in refused")

// Principal is the signed-in person.
type Principal struct {
	AccountID  uuid.UUID
	Email      string
	IsOperator bool
}

// IsOperator reports whether email is the configured оператор address, ignoring letter case.
// An empty configured address matches nobody.
func IsOperator(email, operatorEmail string) bool {
	operatorEmail = strings.TrimSpace(operatorEmail)
	return operatorEmail != "" && strings.EqualFold(strings.TrimSpace(email), operatorEmail)
}

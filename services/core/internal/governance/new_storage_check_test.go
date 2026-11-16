package governance_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/AndriiOpryshko/spilno/services/core/internal/governance"
)

// governance.New takes the storage check from cmd/api instead of building an identity module.
func TestNewTakesTheStorageCheckAsAnArgument(t *testing.T) {
	check := func(context.Context, uuid.UUID) (bool, error) { return true, nil }
	if m := governance.New(nil, check); m == nil {
		t.Fatal("governance.New returned nil")
	}
}

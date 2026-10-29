// Package http adapts the identity context to the generated strict server.
package http

import (
	"context"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

// Handler stubs every operation until the identity tasks implement them.
type Handler struct{}

var _ gen.StrictServerInterface = (*Handler)(nil)

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) GetSession(_ context.Context, _ gen.GetSessionRequestObject) (gen.GetSessionResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) StartSignIn(_ context.Context, _ gen.StartSignInRequestObject) (gen.StartSignInResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) CompleteSignIn(_ context.Context, _ gen.CompleteSignInRequestObject) (gen.CompleteSignInResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

func (h *Handler) SignOut(_ context.Context, _ gen.SignOutRequestObject) (gen.SignOutResponseObject, error) {
	return nil, httpapi.ErrNotImplemented
}

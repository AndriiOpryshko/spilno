// Package identity is the identity bounded context: sign-in, sessions and the current principal.
//
// module.go is the only package other contexts and cmd/ may import from this
// context; everything below it is private to the context (enforced by depguard).
package identity

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	identityhttp "github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

type Module struct {
	handler *identityhttp.Handler
}

func New(_ *pgxpool.Pool) *Module {
	return &Module{handler: identityhttp.NewHandler()}
}

// Routes attaches the context's routes to r inside their own group, validated
// against the context's tag-pruned spec.
func (m *Module) Routes(r chi.Router) error {
	spec, err := gen.GetSpec()
	if err != nil {
		return fmt.Errorf("identity: load spec: %w", err)
	}
	r.Group(func(r chi.Router) {
		r.Use(httpapi.RequestValidator(spec))
		strict := gen.NewStrictHandlerWithOptions(m.handler, nil, gen.StrictHTTPServerOptions{
			RequestErrorHandlerFunc:  httpapi.StrictRequestError,
			ResponseErrorHandlerFunc: httpapi.StrictResponseError,
		})
		gen.HandlerWithOptions(strict, gen.ChiServerOptions{BaseRouter: r, ErrorHandlerFunc: httpapi.ParamBindError})
	})
	return nil
}

// Package system is the system bounded context: service health and runtime info.
//
// module.go is the only package other contexts and cmd/ may import from this
// context; everything below it is private to the context (enforced by depguard).
package system

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
	systemhttp "github.com/AndriiOpryshko/spilno/services/core/internal/system/adapters/http"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system/app"
)

type Module struct {
	handler *systemhttp.Handler
}

func New(pool *pgxpool.Pool) *Module {
	health := app.NewHealthService(postgres.NewProbe(pool))
	return &Module{handler: systemhttp.NewHandler(health)}
}

// Routes attaches the context's routes to r inside their own group, validated
// against the context's tag-pruned spec.
func (m *Module) Routes(r chi.Router) error {
	spec, err := gen.GetSpec()
	if err != nil {
		return fmt.Errorf("system: load spec: %w", err)
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

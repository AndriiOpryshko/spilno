// Package finance is the finance bounded context: the association's contracts and money.
//
// module.go is the only package other contexts and cmd/ may import from this
// context; everything below it is private to the context (enforced by depguard).
package finance

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	financehttp "github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/http"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/mcprates"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/adapters/postgres"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/ports"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/httpapi"
)

type Module struct {
	handler *financehttp.Handler
}

// Config carries what the context needs from outside. The zero value disables the rate source.
type Config struct {
	// Rates reads the official rate; nil means the source is not configured.
	Rates ports.RateProvider
}

func New(_ *pgxpool.Pool, cfg Config) *Module {
	var rates ports.RateProvider = mcprates.Disabled{}
	if cfg.Rates != nil {
		rates = cfg.Rates
	}
	contracts := app.NewContractService(postgres.NewContracts(), rates)
	return &Module{handler: financehttp.NewHandler(contracts)}
}

// Routes attaches the context's routes to r inside their own group, validated
// against the context's tag-pruned spec.
func (m *Module) Routes(r chi.Router) error {
	spec, err := gen.GetSpec()
	if err != nil {
		return fmt.Errorf("finance: load spec: %w", err)
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

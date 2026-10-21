// Package http adapts the system context's use cases to the generated strict server.
package http

import (
	"context"

	"github.com/AndriiOpryshko/spilno/services/core/internal/system/adapters/http/gen"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system/app"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system/domain"
)

type Handler struct {
	health *app.HealthService
}

var _ gen.StrictServerInterface = (*Handler)(nil)

func NewHandler(health *app.HealthService) *Handler {
	return &Handler{health: health}
}

func (h *Handler) GetHealth(ctx context.Context, _ gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	health := h.health.Check(ctx)
	body := toAPI(health)
	if !health.OK() {
		return gen.GetHealth503JSONResponse(body), nil
	}
	return gen.GetHealth200JSONResponse(body), nil
}

func toAPI(h domain.Health) gen.Health {
	out := gen.Health{
		Status:   gen.HealthStatusOk,
		Database: gen.DependencyHealth{Status: gen.DependencyHealthStatusUp},
	}
	if !h.OK() {
		out.Status = gen.HealthStatusDegraded
	}
	if h.Database.Up {
		out.Database.Version = &h.Database.Version
	} else {
		out.Database.Status = gen.DependencyHealthStatusDown
	}
	return out
}

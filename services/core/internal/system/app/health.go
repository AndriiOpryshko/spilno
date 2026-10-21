// Package app holds the system context's use cases.
package app

import (
	"context"

	"github.com/AndriiOpryshko/spilno/services/core/internal/system/domain"
	"github.com/AndriiOpryshko/spilno/services/core/internal/system/ports"
)

type HealthService struct {
	db ports.DatabaseProbe
}

func NewHealthService(db ports.DatabaseProbe) *HealthService {
	return &HealthService{db: db}
}

// Check probes every dependency. A failing probe is reported, not returned as an error.
func (s *HealthService) Check(ctx context.Context) domain.Health {
	version, err := s.db.ServerVersion(ctx)
	if err != nil {
		return domain.Health{Database: domain.DependencyHealth{Up: false}}
	}
	return domain.Health{Database: domain.DependencyHealth{Up: true, Version: version}}
}

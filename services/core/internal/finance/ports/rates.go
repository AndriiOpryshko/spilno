package ports

import (
	"context"
	"time"

	"github.com/AndriiOpryshko/spilno/services/core/internal/finance/domain"
)

// RateProvider returns the official rate of a currency for a calendar day.
// It answers domain.ErrRateNotPublished or domain.ErrRateUnavailable.
type RateProvider interface {
	Rate(ctx context.Context, currency domain.Currency, day time.Time) (domain.Rate, error)
}

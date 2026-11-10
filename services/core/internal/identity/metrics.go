package identity

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// RegisterMetrics registers the gauge identity.applications.pending_age_days:
// whole days since the oldest pending application was filed, 0 when none.
func (m *Module) RegisterMetrics() error {
	_, err := otel.Meter("identity").Int64ObservableGauge("identity.applications.pending_age_days",
		metric.WithDescription("Age in days of the oldest pending application"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			var days int64
			err := m.pool.QueryRow(ctx, `SELECT COALESCE(now()::date - min(created_at)::date, 0)
				FROM applications WHERE state = 'pending'`).Scan(&days)
			if err != nil {
				return nil
			}
			o.Observe(days)
			return nil
		}))
	return err
}

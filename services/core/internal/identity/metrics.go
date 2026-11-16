package identity

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// RegisterMetrics registers the gauge identity.applications.pending_age_days:
// whole days since the oldest pending application was last filed, 0 when none.
func (m *Module) RegisterMetrics() error {
	_, err := otel.Meter("identity").Int64ObservableGauge("identity.applications.pending_age_days",
		metric.WithDescription("Age in days of the oldest pending application"),
		metric.WithInt64Callback(func(ctx context.Context, o metric.Int64Observer) error {
			var days int64
			err := m.pool.QueryRow(ctx, `SELECT COALESCE(now()::date - min(COALESCE(
					(SELECT max(f.created_at) FROM application_filings f WHERE f.application_id = a.id),
					a.created_at))::date, 0)
				FROM applications a WHERE a.state = 'pending'`).Scan(&days)
			if err != nil {
				slog.ErrorContext(ctx, "identity: pending application age query failed", "error", err)
				return nil
			}
			o.Observe(days)
			return nil
		}))
	return err
}

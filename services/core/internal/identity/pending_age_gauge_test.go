package identity_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/AndriiOpryshko/spilno/services/core/internal/identity"
	"github.com/AndriiOpryshko/spilno/services/core/internal/identity/identitytest"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const pendingAgeGauge = "identity.applications.pending_age_days"

var (
	pendingAgeOnce   sync.Once
	pendingAgeReader *sdkmetric.ManualReader
)

func pendingAgeValue(t *testing.T) (int64, bool) {
	t.Helper()
	pendingAgeOnce.Do(func() {
		pendingAgeReader = sdkmetric.NewManualReader()
		otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(pendingAgeReader)))
	})
	var rm metricdata.ResourceMetrics
	if err := pendingAgeReader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != pendingAgeGauge {
				continue
			}
			g, ok := m.Data.(metricdata.Gauge[int64])
			if !ok {
				t.Fatalf("%s is %T, want an int64 gauge", pendingAgeGauge, m.Data)
			}
			if len(g.DataPoints) != 1 {
				t.Fatalf("%s has %d data points, want 1", pendingAgeGauge, len(g.DataPoints))
			}
			return g.DataPoints[0].Value, true
		}
	}
	return 0, false
}

func TestPendingAgeGaugeReportsTheOldestPendingApplicationInDays(t *testing.T) {
	pendingAgeValue(t) // install the reader before the gauge is registered
	pool := testdb.NewApp(t)
	if err := identity.New(pool).RegisterMetrics(); err != nil {
		t.Fatalf("RegisterMetrics: %v", err)
	}

	if got, ok := pendingAgeValue(t); !ok || got != 0 {
		t.Fatalf("%s with no applications = %d (reported %v), want 0", pendingAgeGauge, got, ok)
	}

	now := time.Now().UTC()
	identitytest.PendingApplication(t, pool, "11111111", now.AddDate(0, 0, -1))
	identitytest.PendingApplication(t, pool, "22222222", now.AddDate(0, 0, -4))
	if got, _ := pendingAgeValue(t); got != 4 {
		t.Errorf("%s = %d, want 4 (the oldest pending)", pendingAgeGauge, got)
	}
}

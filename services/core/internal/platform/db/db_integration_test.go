package db_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/db"
	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/testdb"
)

const counterName = "platform.db.association_unset"

var (
	readerOnce sync.Once
	reader     *sdkmetric.ManualReader
)

func metricReader() *sdkmetric.ManualReader {
	readerOnce.Do(func() {
		reader = sdkmetric.NewManualReader()
		otel.SetMeterProvider(sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)))
	})
	return reader
}

func counterValue(t *testing.T) int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	if err := metricReader().Collect(context.Background(), &rm); err != nil {
		t.Fatalf("collect: %v", err)
	}
	var total int64
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != counterName {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			if !ok {
				t.Fatalf("%s is %T, want an int64 sum", counterName, m.Data)
			}
			for _, dp := range sum.DataPoints {
				total += dp.Value
			}
		}
	}
	return total
}

func setting(ctx context.Context, t *testing.T, q db.DBTX, key string) string {
	t.Helper()
	var v string
	if err := q.QueryRow(ctx, "SELECT coalesce(current_setting($1, true), '')", key).Scan(&v); err != nil {
		t.Fatalf("read %s: %v", key, err)
	}
	return v
}

func TestInAssociation_SetsAndClearsSettings(t *testing.T) {
	metricReader()
	pool := testdb.New(t)
	ctx := context.Background()
	assoc := uuid.New()
	const actor = "member@example.test"

	// One connection, so "after commit" is observed on the same session.
	cfg := pool.Config().Copy()
	cfg.MaxConns = 1
	one, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(one.Close)

	err = db.InAssociation(ctx, one, assoc, actor, func(ctx context.Context) error {
		q, err := db.From(ctx)
		if err != nil {
			return err
		}
		if got := setting(ctx, t, q, "app.association_id"); got != assoc.String() {
			t.Errorf("app.association_id in fn = %q, want %q", got, assoc)
		}
		if got := setting(ctx, t, q, "app.actor_email"); got != actor {
			t.Errorf("app.actor_email in fn = %q, want %q", got, actor)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("InAssociation: %v", err)
	}
	for _, key := range []string{"app.association_id", "app.actor_email"} {
		if got := setting(ctx, t, one, key); got != "" {
			t.Errorf("%s after commit = %q, want empty", key, got)
		}
	}
}

func TestInAssociation_RollsBackOnError(t *testing.T) {
	pool := testdb.New(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, "CREATE TABLE t04_probe (v int)"); err != nil {
		t.Fatal(err)
	}
	boom := errors.New("boom")
	err := db.InAssociation(ctx, pool, uuid.New(), "a@example.test", func(ctx context.Context) error {
		q, err := db.From(ctx)
		if err != nil {
			return err
		}
		if _, err := q.Exec(ctx, "INSERT INTO t04_probe VALUES (1)"); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("InAssociation error = %v, want boom", err)
	}
	var n int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM t04_probe").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("rows after rollback = %d, want 0", n)
	}
}

func TestInAssociation_RollsBackOnPanic(t *testing.T) {
	pool := testdb.New(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, "CREATE TABLE t04_probe (v int)"); err != nil {
		t.Fatal(err)
	}
	func() {
		defer func() { _ = recover() }()
		_ = db.InAssociation(ctx, pool, uuid.New(), "a@example.test", func(ctx context.Context) error {
			q, _ := db.From(ctx)
			_, _ = q.Exec(ctx, "INSERT INTO t04_probe VALUES (1)")
			panic("boom")
		})
	}()
	var n int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM t04_probe").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("rows after panic = %d, want 0", n)
	}
}

func TestInAssociation_Nested(t *testing.T) {
	pool := testdb.New(t)
	ctx := context.Background()
	assoc := uuid.New()

	err := db.InAssociation(ctx, pool, assoc, "a@example.test", func(outer context.Context) error {
		a, err := db.From(outer)
		if err != nil {
			return err
		}
		return db.InAssociation(outer, pool, assoc, "a@example.test", func(inner context.Context) error {
			b, err := db.From(inner)
			if err != nil {
				return err
			}
			if a != b {
				t.Error("nested call for the same association did not reuse the transaction")
			}
			return nil
		})
	})
	if err != nil {
		t.Fatalf("same-association nesting: %v", err)
	}

	err = db.InAssociation(ctx, pool, assoc, "a@example.test", func(outer context.Context) error {
		return db.InAssociation(outer, pool, uuid.New(), "a@example.test", func(context.Context) error {
			t.Error("fn ran for another association inside an open transaction")
			return nil
		})
	})
	if err == nil {
		t.Error("nesting with another association returned nil, want an error")
	}
}

func TestFrom_WithoutTransactionIsCounted(t *testing.T) {
	before := counterValue(t)
	q, err := db.From(context.Background())
	if err == nil || q != nil {
		t.Errorf("From without a transaction = (%v, %v), want (nil, error)", q, err)
	}
	if got := counterValue(t); got != before+1 {
		t.Errorf("%s = %d, want %d", counterName, got, before+1)
	}
}

package nbu_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/AndriiOpryshko/spilno/mcp/nbu-rates/internal/nbu"
)

var day = time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC)

func upstream(t *testing.T, h http.HandlerFunc) *nbu.Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return nbu.NewClient(srv.URL)
}

func TestRateKeepsThePublishedDecimal(t *testing.T) {
	var query string
	c := upstream(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.RawQuery
		_, _ = w.Write([]byte(`[{"r030":978,"txt":"Євро","rate":50.5311,"cc":"EUR","exchangedate":"20.10.2026"}]`))
	})
	got, err := c.Rate(context.Background(), "EUR", day)
	if err != nil {
		t.Fatal(err)
	}
	if got.Value != "50.5311" || got.Currency != "EUR" || !got.Date.Equal(day) {
		t.Errorf("rate = %+v, want EUR 50.5311 on %s", got, day)
	}
	if query != "date=20261020&valcode=EUR&json" {
		t.Errorf("query = %q", query)
	}
}

func TestRateIsFetchedOncePerCurrencyAndDay(t *testing.T) {
	var calls atomic.Int32
	c := upstream(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		_, _ = w.Write([]byte(`[{"rate":45.051,"cc":"USD"}]`))
	})
	for range 3 {
		if _, err := c.Rate(context.Background(), "USD", day); err != nil {
			t.Fatal(err)
		}
	}
	if n := calls.Load(); n != 1 {
		t.Errorf("upstream calls = %d, want 1", n)
	}
}

func TestRateErrors(t *testing.T) {
	tests := []struct {
		name     string
		currency string
		status   int
		body     string
		want     error
	}{
		{"empty list is not published", "EUR", 200, `[]`, nbu.ErrNotPublished},
		{"server error is unavailable", "EUR", 502, `oops`, nbu.ErrUnavailable},
		{"broken body is unavailable", "EUR", 200, `<html>`, nbu.ErrUnavailable},
		{"another currency's row is unavailable", "EUR", 200, `[{"rate":45.05,"cc":"USD"}]`, nbu.ErrUnavailable},
		{"non-decimal rate is unavailable", "EUR", 200, `[{"rate":1e2,"cc":"EUR"}]`, nbu.ErrUnavailable},
		{"hryvnia has no rate", "UAH", 200, `[]`, nbu.ErrInvalidCurrency},
		{"lower case is rejected", "eur", 200, `[]`, nbu.ErrInvalidCurrency},
		{"query injection is rejected", "EUR&date=1", 200, `[]`, nbu.ErrInvalidCurrency},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := upstream(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			})
			_, err := c.Rate(context.Background(), tt.currency, day)
			if !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}

// A body over the 1 MiB cap is refused even when it is otherwise a valid answer.
func TestRateRefusesAnOversizedBody(t *testing.T) {
	pad := strings.Repeat("x", 1<<20)
	c := upstream(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`[{"rate":45.05,"cc":"EUR","txt":"` + pad + `"}]`))
	})
	if _, err := c.Rate(context.Background(), "EUR", day); !errors.Is(err, nbu.ErrUnavailable) {
		t.Errorf("err = %v, want %v", err, nbu.ErrUnavailable)
	}
}

// An upstream that never answers is given up on at the caller's deadline.
func TestRateGivesUpAtTheDeadline(t *testing.T) {
	c := upstream(t, func(_ http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := c.Rate(ctx, "EUR", day)
	if !errors.Is(err, nbu.ErrUnavailable) {
		t.Errorf("err = %v, want %v", err, nbu.ErrUnavailable)
	}
	if d := time.Since(start); d > time.Second {
		t.Errorf("Rate returned after %v, want it at the 50ms deadline", d)
	}
}

// Without a caller deadline the client still gives up after 10 s.
func TestNewClientBoundsEveryRequest(t *testing.T) {
	if got := nbu.NewClient("").HTTP.Timeout; got != 10*time.Second {
		t.Errorf("client timeout = %v, want 10s", got)
	}
}

func TestAFailureIsNotCached(t *testing.T) {
	var calls atomic.Int32
	c := upstream(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`[{"rate":50.5311,"cc":"EUR"}]`))
	})
	if _, err := c.Rate(context.Background(), "EUR", day); !errors.Is(err, nbu.ErrUnavailable) {
		t.Fatalf("first call err = %v, want unavailable", err)
	}
	if _, err := c.Rate(context.Background(), "EUR", day); err != nil {
		t.Fatalf("second call err = %v, want a rate", err)
	}
}

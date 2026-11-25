package domain

import (
	"errors"
	"math"
	"testing"
)

func TestRateConvertRoundsHalfUp(t *testing.T) {
	tests := []struct {
		name   string
		amount int64
		rate   string
		want   int64
	}{
		{"AC-02 EUR contract", 125000, "50.5311", 6316388},
		{"below half", 1, "45.051", 45},
		{"45.055 rounds to 45", 1, "45.055", 45},
		{"450.55 rounds up", 10, "45.055", 451},
		{"no overflow", 9_000_000_000_000_000, "50.5311", 454_779_900_000_000_000},
		{"largest amount at rate 1", math.MaxInt64, "1", math.MaxInt64},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := ParseRate(CurrencyEUR, date("2026-10-20"), tc.rate)
			if err != nil {
				t.Fatalf("ParseRate(%q) error: %v", tc.rate, err)
			}
			got, err := r.Convert(tc.amount)
			if err != nil {
				t.Fatalf("Convert(%d) error: %v", tc.amount, err)
			}
			if got.AmountKopiyky != tc.want {
				t.Errorf("%d x %s = %d kopiyky, want %d", tc.amount, tc.rate, got.AmountKopiyky, tc.want)
			}
			if got.Rate != r {
				t.Errorf("equivalent carries rate %+v, want %+v", got.Rate, r)
			}
		})
	}
}

// An equivalent that does not fit int64 kopiyky is refused, never truncated.
func TestRateConvertRefusesOutOfRange(t *testing.T) {
	for _, tc := range []struct {
		amount int64
		rate   string
	}{
		{200_000_000_000_000_000, "50.5311"},
		{math.MaxInt64, "1.000001"},
	} {
		r, err := ParseRate(CurrencyEUR, date("2026-10-20"), tc.rate)
		if err != nil {
			t.Fatal(err)
		}
		got, err := r.Convert(tc.amount)
		if !errors.Is(err, ErrEquivalentOutOfRange) {
			t.Errorf("%d x %s = %+v, %v; want ErrEquivalentOutOfRange", tc.amount, tc.rate, got, err)
		}
	}
}

func TestParseRateRefusesMalformed(t *testing.T) {
	for _, in := range []string{"", "-1", "0", "1e2", "1.2345678", "1,5"} {
		t.Run(in, func(t *testing.T) {
			if _, err := ParseRate(CurrencyUSD, date("2026-10-20"), in); err == nil {
				t.Errorf("ParseRate(%q) accepted a malformed rate", in)
			}
		})
	}
}

func TestRateDecimalRoundTrips(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"50.5311", "50.5311"},
		{"45.051", "45.051"},
		{"45", "45"},
		{"45.0510", "45.051"},
	} {
		t.Run(c.in, func(t *testing.T) {
			r, err := ParseRate(CurrencyEUR, date("2026-10-20"), c.in)
			if err != nil {
				t.Fatalf("ParseRate(%q) error: %v", c.in, err)
			}
			if got := r.Decimal(); got != c.want {
				t.Errorf("Decimal() = %q, want %q", got, c.want)
			}
		})
	}
}

func TestCurrencyForeign(t *testing.T) {
	for c, want := range map[Currency]bool{CurrencyUAH: false, CurrencyEUR: true, CurrencyUSD: true} {
		if got := c.Foreign(); got != want {
			t.Errorf("%s.Foreign() = %v, want %v", c, got, want)
		}
	}
}

package domain

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// ErrRateNotPublished: the source has no rate for that currency and day.
var ErrRateNotPublished = errors.New("rate not published")

// ErrRateUnavailable: the source could not be asked or its answer is untrusted.
var ErrRateUnavailable = errors.New("rate unavailable")

// ErrEquivalentOutOfRange: the hryvnia equivalent does not fit int64 kopiyky.
var ErrEquivalentOutOfRange = errors.New("equivalent out of range")

// Rate is an official rate: hryvnia times 10^6 per one unit of Currency.
type Rate struct {
	Currency Currency
	Day      time.Time
	Micros   int64
}

// UAHEquivalent is the hryvnia value of a foreign-currency amount.
type UAHEquivalent struct {
	AmountKopiyky int64
	Rate          Rate
}

// ParseRate reads a positive decimal with at most six fraction digits.
func ParseRate(currency Currency, day time.Time, decimal string) (Rate, error) {
	whole, frac, _ := strings.Cut(decimal, ".")
	if whole == "" || len(frac) > 6 || !digitsOnly(whole) || !digitsOnly(frac) {
		return Rate{}, ErrRateUnavailable
	}
	if strings.Contains(decimal, ".") && frac == "" {
		return Rate{}, ErrRateUnavailable
	}
	micros, err := strconv.ParseInt(whole+frac+strings.Repeat("0", 6-len(frac)), 10, 64)
	if err != nil || micros <= 0 {
		return Rate{}, ErrRateUnavailable
	}
	return Rate{Currency: currency, Day: day, Micros: micros}, nil
}

func digitsOnly(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Decimal formats the rate as published, trailing zeros dropped.
func (r Rate) Decimal() string {
	whole := r.Micros / 1_000_000
	frac := strings.TrimRight(fmt.Sprintf("%06d", r.Micros%1_000_000), "0")
	if frac == "" {
		return strconv.FormatInt(whole, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + frac
}

// Convert returns amountCents times the rate in kopiyky, rounded half up, or
// ErrEquivalentOutOfRange when the result does not fit int64.
func (r Rate) Convert(amountCents int64) (UAHEquivalent, error) {
	n := new(big.Int).Mul(big.NewInt(amountCents), big.NewInt(r.Micros))
	n.Add(n, big.NewInt(500_000))
	n.Quo(n, big.NewInt(1_000_000))
	if !n.IsInt64() {
		return UAHEquivalent{}, ErrEquivalentOutOfRange
	}
	return UAHEquivalent{AmountKopiyky: n.Int64(), Rate: r}, nil
}

// Foreign reports whether the currency is not the hryvnia.
func (c Currency) Foreign() bool { return c.Valid() && c != CurrencyUAH }

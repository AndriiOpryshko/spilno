// Package nbu reads official hryvnia exchange rates from the National Bank of
// Ukraine's public statistics API.
package nbu

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

// DefaultBaseURL is the NBU endpoint that lists official rates for one day.
const DefaultBaseURL = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchange"

var (
	// ErrNotPublished means the NBU has no official rate of that currency for
	// that day: the day is too far ahead, predates the records, or the code is
	// not a currency the NBU quotes.
	ErrNotPublished = errors.New("rate not published")
	// ErrUnavailable means the NBU could not be reached or answered with
	// something other than a rate list.
	ErrUnavailable = errors.New("nbu unavailable")
	// ErrInvalidCurrency means the code is not three Latin letters, or is UAH.
	ErrInvalidCurrency = errors.New("invalid currency")
)

// Rate is the official price of one unit of Currency in hryvnias on Date.
type Rate struct {
	Currency string
	Date     time.Time
	// Value is the decimal exactly as the NBU published it, e.g. "50.5311".
	Value string
}

var (
	currencyRE = regexp.MustCompile(`^[A-Z]{3}$`)
	decimalRE  = regexp.MustCompile(`^\d+(\.\d+)?$`)
)

// Client fetches rates and remembers the ones it has seen: a published
// official rate never changes.
type Client struct {
	BaseURL string
	HTTP    *http.Client

	mu    sync.Mutex
	cache map[string]Rate
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{BaseURL: baseURL, HTTP: &http.Client{Timeout: 10 * time.Second}, cache: map[string]Rate{}}
}

const maxCached = 4096

// Rate returns the official rate of currency (ISO 4217, upper case) on the
// calendar day of date.
func (c *Client) Rate(ctx context.Context, currency string, date time.Time) (Rate, error) {
	if !currencyRE.MatchString(currency) || currency == "UAH" {
		return Rate{}, fmt.Errorf("%w: %q", ErrInvalidCurrency, currency)
	}
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	key := currency + day.Format("20060102")

	c.mu.Lock()
	hit, ok := c.cache[key]
	c.mu.Unlock()
	if ok {
		return hit, nil
	}

	q := url.Values{"valcode": {currency}, "date": {day.Format("20060102")}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"?"+q.Encode()+"&json", nil)
	if err != nil {
		return Rate{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return Rate{}, fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Rate{}, fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode)
	}

	var rows []struct {
		Rate json.Number `json:"rate"`
		CC   string      `json:"cc"`
	}
	dec := json.NewDecoder(http.MaxBytesReader(nil, resp.Body, 1<<20))
	dec.UseNumber()
	if err := dec.Decode(&rows); err != nil {
		return Rate{}, fmt.Errorf("%w: decode: %v", ErrUnavailable, err)
	}
	if len(rows) == 0 {
		return Rate{}, fmt.Errorf("%w: %s on %s", ErrNotPublished, currency, day.Format(time.DateOnly))
	}
	v := rows[0].Rate.String()
	if rows[0].CC != currency || !decimalRE.MatchString(v) {
		return Rate{}, fmt.Errorf("%w: unexpected row %q %q", ErrUnavailable, rows[0].CC, v)
	}

	r := Rate{Currency: currency, Date: day, Value: v}
	c.mu.Lock()
	if len(c.cache) >= maxCached {
		clear(c.cache)
	}
	c.cache[key] = r
	c.mu.Unlock()
	return r, nil
}

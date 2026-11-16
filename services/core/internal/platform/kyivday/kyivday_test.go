package kyivday_test

import (
	"testing"
	"time"

	"github.com/AndriiOpryshko/spilno/services/core/internal/platform/kyivday"
)

// Contract this test pins: kyivday.Today(now) is the Europe/Kyiv calendar day of the instant now,
// as midnight UTC of that date (the shape roster dates are compared in).
func TestToday_TakesTheKyivCalendarDay(t *testing.T) {
	cases := []struct {
		name string
		now  string // RFC3339 instant
		want string
	}{
		{"winter, 23:59 Kyiv is still the day", "2026-11-04T21:59:00Z", "2026-11-04"},
		{"winter, 00:00 Kyiv starts the next day", "2026-11-04T22:00:00Z", "2026-11-05"},
		{"winter, 00:40 Kyiv, UTC still the day before", "2026-11-04T22:40:00Z", "2026-11-05"},
		{"summer, 23:59 Kyiv is still the day", "2026-07-14T20:59:00Z", "2026-07-14"},
		{"summer, 00:00 Kyiv starts the next day", "2026-07-14T21:00:00Z", "2026-07-15"},
		{"summer, 00:40 Kyiv, UTC still the day before", "2026-07-14T21:40:00Z", "2026-07-15"},
		{"an instant given in another zone", "2026-11-04T17:40:00-05:00", "2026-11-05"},
	}
	for _, c := range cases {
		now, err := time.Parse(time.RFC3339, c.now)
		if err != nil {
			t.Fatal(err)
		}
		got := kyivday.Today(now)
		want, _ := time.Parse("2006-01-02", c.want)
		if !got.Equal(want) || got.Location() != time.UTC {
			t.Errorf("%s: Today(%s) = %s, want midnight UTC %s", c.name, c.now, got.Format(time.RFC3339), c.want)
		}
	}
}

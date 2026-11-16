// Package kyivday takes «today» for roster dates from the Europe/Kyiv calendar.
package kyivday

import (
	"time"
	_ "time/tzdata" // the zone must resolve on images without a system tz database
)

var zone = func() *time.Location {
	if l, err := time.LoadLocation("Europe/Kyiv"); err == nil {
		return l
	}
	return time.UTC
}()

// Today returns midnight UTC of the Europe/Kyiv calendar date of now, the shape roster dates are compared in.
func Today(now time.Time) time.Time {
	y, m, d := now.In(zone).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

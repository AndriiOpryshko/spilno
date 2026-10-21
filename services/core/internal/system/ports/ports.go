// Package ports declares what the system context needs from the outside world.
package ports

import "context"

// DatabaseProbe reports whether the database answers and which version it runs.
type DatabaseProbe interface {
	ServerVersion(ctx context.Context) (string, error)
}

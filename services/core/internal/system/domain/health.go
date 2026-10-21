// Package domain holds the system context's model. It imports nothing outside the standard library.
package domain

// DependencyHealth is the observed state of one dependency.
type DependencyHealth struct {
	Up      bool
	Version string
}

// Health is the service's overall state.
type Health struct {
	Database DependencyHealth
}

// OK reports whether every dependency is up.
func (h Health) OK() bool { return h.Database.Up }

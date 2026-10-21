package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/AndriiOpryshko/spilno/services/core/internal/system/app"
)

type fakeProbe struct {
	version string
	err     error
}

func (f fakeProbe) ServerVersion(context.Context) (string, error) { return f.version, f.err }

func TestCheck(t *testing.T) {
	tests := []struct {
		name        string
		probe       fakeProbe
		wantOK      bool
		wantVersion string
	}{
		{name: "database up", probe: fakeProbe{version: "18.6"}, wantOK: true, wantVersion: "18.6"},
		{name: "database down", probe: fakeProbe{err: errors.New("connection refused")}, wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := app.NewHealthService(tt.probe).Check(context.Background())
			if h.OK() != tt.wantOK {
				t.Errorf("OK() = %v, want %v", h.OK(), tt.wantOK)
			}
			if h.Database.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", h.Database.Version, tt.wantVersion)
			}
		})
	}
}

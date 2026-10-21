// Package config reads service configuration from the environment.
package config

import "os"

type Config struct {
	ServiceName string
	HTTPAddr    string
	DatabaseURL string
	// OTLPEndpoint enables OpenTelemetry export when set. The exporters read the
	// standard OTEL_EXPORTER_OTLP_* variables themselves; this only gates them.
	OTLPEndpoint string
}

func Load() Config {
	return Config{
		ServiceName:  env("OTEL_SERVICE_NAME", "spilno-core"),
		HTTPAddr:     env("HTTP_ADDR", ":8080"),
		DatabaseURL:  env("DATABASE_URL", "postgres://spilno:spilno@localhost:5432/spilno?sslmode=disable"),
		OTLPEndpoint: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Package config reads service configuration from the environment.
package config

import "os"

type Config struct {
	ServiceName string
	HTTPAddr    string
	// DatabaseURL is the login role the API serves with (a member of spilno_app).
	DatabaseURL string
	// MigrationsDatabaseURL is the owner that runs goose on start; it falls back to DatabaseURL.
	MigrationsDatabaseURL string
	// OTLPEndpoint enables OpenTelemetry export when set. The exporters read the
	// standard OTEL_EXPORTER_OTLP_* variables themselves; this only gates them.
	OTLPEndpoint string
	// DevSignIn turns on the development sign-in shim; off unless SPILNO_DEV_SIGN_IN=true.
	DevSignIn bool
	// OperatorEmail is the address whose session is the оператор (SPILNO_OPERATOR_EMAIL).
	OperatorEmail string
}

func Load() Config {
	dbURL := env("DATABASE_URL", "postgres://spilno:spilno@localhost:5432/spilno?sslmode=disable")
	return Config{
		ServiceName:           env("OTEL_SERVICE_NAME", "spilno-core"),
		HTTPAddr:              env("HTTP_ADDR", ":8080"),
		DatabaseURL:           dbURL,
		MigrationsDatabaseURL: env("MIGRATIONS_DATABASE_URL", dbURL),
		OTLPEndpoint:          os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		DevSignIn:             os.Getenv("SPILNO_DEV_SIGN_IN") == "true",
		OperatorEmail:         os.Getenv("SPILNO_OPERATOR_EMAIL"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

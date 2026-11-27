// Package config reads service configuration from the environment.
package config

import "os"

type Config struct {
	ServiceName string
	HTTPAddr    string
	// DatabaseURL is the login role the API serves with (a member of spilno_app); it has no default.
	DatabaseURL string
	// MigrationsDatabaseURL is the owner that runs goose on start; it has no default.
	MigrationsDatabaseURL string
	// OTLPEndpoint enables OpenTelemetry export when set. The exporters read the
	// standard OTEL_EXPORTER_OTLP_* variables themselves; this only gates them.
	OTLPEndpoint string
	// DevSignIn turns on the development sign-in shim; off unless SPILNO_DEV_SIGN_IN=true.
	DevSignIn bool
	// DevStorage swaps the Google storage consent for a fake; off unless SPILNO_DEV_STORAGE=true.
	DevStorage bool
	// OperatorEmail is the address whose session is the оператор (SPILNO_OPERATOR_EMAIL).
	OperatorEmail string
	// WebOrigin is where the web app is served (SPILNO_WEB_ORIGIN); the storage consent returns to it.
	WebOrigin string
	// NBURatesMCPURL is the NBU rates MCP endpoint (NBU_RATES_MCP_URL); empty disables the rate source.
	NBURatesMCPURL string
}

func Load() Config {
	return Config{
		ServiceName:           env("OTEL_SERVICE_NAME", "spilno-core"),
		HTTPAddr:              env("HTTP_ADDR", ":8080"),
		DatabaseURL:           os.Getenv("DATABASE_URL"),
		MigrationsDatabaseURL: os.Getenv("MIGRATIONS_DATABASE_URL"),
		OTLPEndpoint:          os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		DevSignIn:             os.Getenv("SPILNO_DEV_SIGN_IN") == "true",
		DevStorage:            os.Getenv("SPILNO_DEV_STORAGE") == "true",
		OperatorEmail:         os.Getenv("SPILNO_OPERATOR_EMAIL"),
		WebOrigin:             env("SPILNO_WEB_ORIGIN", "http://localhost:5180"),
		NBURatesMCPURL:        os.Getenv("NBU_RATES_MCP_URL"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

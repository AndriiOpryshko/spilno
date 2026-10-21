// Package migrations embeds the goose SQL migrations so the binary and tests apply the same files.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

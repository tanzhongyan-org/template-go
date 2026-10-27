// Package sql embeds the canonical schema so the server applies exactly the
// DDL that sqlc generates code from.
package sql

import _ "embed"

// Schema is the contents of schema.sql.
//
//go:embed schema.sql
var Schema string

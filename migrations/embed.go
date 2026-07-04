package migrations

import "embed"

// Files contains versioned SQL migrations for golang-migrate.
//
//go:embed *.sql
var Files embed.FS

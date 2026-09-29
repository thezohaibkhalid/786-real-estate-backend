package migrations

import "embed"

// Files contains ordered, forward-only SQL migrations.
//
//go:embed *.sql
var Files embed.FS

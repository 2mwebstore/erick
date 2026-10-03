// Package migrations holds the schema as numbered SQL files and embeds them in
// the binary, so the API can apply them itself when it starts.
package migrations

import "embed"

// FS holds every forward migration. The down files stay out: rolling back is a
// deliberate act done with scripts/migrate.sh, never something a restart does.
//
//go:embed *.up.sql
var FS embed.FS

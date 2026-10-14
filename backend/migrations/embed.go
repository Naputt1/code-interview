// Package migrations embeds the numbered SQL migrations so the API can bring
// the schema up to date on boot without an external migration binary.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS

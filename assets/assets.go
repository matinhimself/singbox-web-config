package assets

import "embed"

//go:embed web/templates/*.html
var TemplatesFS embed.FS

//go:embed web/static
var StaticFS embed.FS

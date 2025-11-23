package webassets

import "embed"

// TemplatesFS embeds all HTML templates from web/templates/
//
//go:embed web/templates/*.html web/templates/components/*.html
var TemplatesFS embed.FS

// StaticFS embeds all static files from web/static/
//
//go:embed web/static
var StaticFS embed.FS

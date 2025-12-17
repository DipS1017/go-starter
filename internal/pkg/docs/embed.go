package docs

import "embed"

//go:embed index.html
var IndexDoc embed.FS

//go:embed swagger.json
var SwaggerDoc embed.FS

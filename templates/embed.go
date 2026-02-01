package templates

import "embed"

//go:embed all:pom all:java all:docs all:idea all:docker
var FS embed.FS

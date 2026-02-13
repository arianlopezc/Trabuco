package templates

import "embed"

//go:embed all:pom all:java all:docs all:idea all:docker all:ai all:claude all:cursor
var FS embed.FS

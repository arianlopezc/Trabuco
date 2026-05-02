package templates

import "embed"

//go:embed all:pom all:java all:docs all:idea all:docker all:ai all:claude all:cursor all:copilot all:codex all:github all:trabuco all:skills all:maven-wrapper
var FS embed.FS

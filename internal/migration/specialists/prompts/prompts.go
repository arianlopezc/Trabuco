// Package prompts holds the embedded system prompts for every LLM-driven
// migration specialist (M4-M10). Centralizing prompts here keeps the
// specialist Go layer thin: registry.go invokes llm.New() with the
// right prompt and phase, no per-specialist Go boilerplate needed.
package prompts

import _ "embed"

//go:embed model.md
var Model string

//go:embed datastore.md
var Datastore string

//go:embed shared.md
var Shared string

//go:embed api.md
var API string

//go:embed worker.md
var Worker string

//go:embed eventconsumer.md
var EventConsumer string

//go:embed aiagent.md
var AIAgent string

//go:embed config.md
var Config string

//go:embed deployment.md
var Deployment string

//go:embed tests.md
var Tests string

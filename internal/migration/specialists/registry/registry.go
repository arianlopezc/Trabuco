// Package registry collects all specialist registrations into a single
// place that the CLI and MCP layers import. Each milestone (M2-M10) adds
// its specialist's blank-import here, which triggers the specialist's
// init() to call specialists.Default().Register(...).
//
// In M1 (foundations) no specialists are registered yet — calls to
// orchestrator.RunPhase() return "no specialist registered for phase X"
// until that phase's milestone lands. This is intentional and lets the
// foundation ship cleanly without requiring all specialists to exist.
package registry

import (
	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
)

// Wired specialists go below as blank imports. Each specialist package
// contains an init() that registers itself with the default registry.
//
// Example (uncomment as M2-M10 land):
//	_ "github.com/arianlopezc/Trabuco/internal/migration/specialists/assessor"
//	_ "github.com/arianlopezc/Trabuco/internal/migration/specialists/skeleton"
//	_ "github.com/arianlopezc/Trabuco/internal/migration/specialists/model"
//	... etc.

// keep specialists import alive for blank-import hygiene.
var _ = specialists.NewRegistry

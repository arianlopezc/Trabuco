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
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/assessor"
)

// init() registers every specialist with specialists.Default(). Each
// milestone (M2-M10) appends its specialist registration here.
func init() {
	r := specialists.Default()
	r.Register(assessor.New())
	// M3+ specialists register themselves below as they land.
}

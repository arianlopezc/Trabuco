package skeleton

import (
	"context"
	"fmt"

	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// Specialist is the Phase 1 skeleton-builder.
//
// It does NOT use the LLM runner — skeleton bootstrap is structural file
// generation driven by state.TargetConfig (which the assessor + user-gate
// already finalized in Phase 0). The LLM-driven planning happens
// implicitly via the assessor's RecommendedTarget; if the user wanted a
// different shape, they edited it before this phase ran.
type Specialist struct{}

// New constructs the skeleton-builder specialist.
func New() *Specialist { return &Specialist{} }

// Phase implements specialists.Specialist.
func (s *Specialist) Phase() types.Phase { return types.PhaseSkeleton }

// Name implements specialists.Specialist.
func (s *Specialist) Name() string { return "skeleton-builder" }

// Run implements specialists.Specialist. Generates the multi-module
// skeleton in migration mode and wraps legacy code into a legacy/ module.
func (s *Specialist) Run(ctx context.Context, in *specialists.Input) (*specialists.Output, error) {
	target := in.State.TargetConfig
	if len(target.Modules) == 0 {
		return nil, fmt.Errorf("phase 1 requires state.TargetConfig.Modules to be set (run phase 0 first)")
	}

	groupID, projectName := LoadGroupAndProjectFromState(in.RepoRoot, &target)
	javaVersion := target.JavaVersion
	if javaVersion == "" {
		javaVersion = "21"
	}

	gen := &Generator{
		RepoRoot:    in.RepoRoot,
		GroupID:     groupID,
		ProjectName: projectName,
		JavaVersion: javaVersion,
		Modules:     target.Modules,
	}

	if err := gen.Generate(); err != nil {
		return nil, fmt.Errorf("generate skeleton: %w", err)
	}
	if err := gen.WrapLegacy(); err != nil {
		return nil, fmt.Errorf("wrap legacy: %w", err)
	}

	items := []types.OutputItem{
		{
			ID:          "skeleton-parent-pom",
			State:       types.ItemApplied,
			Description: "wrote migration-mode parent pom.xml (no enforcer/spotless — added by activator in Phase 12)",
		},
		{
			ID:          "skeleton-modules",
			State:       types.ItemApplied,
			Description: fmt.Sprintf("created %d module skeletons: %v", len(target.Modules), target.Modules),
		},
		{
			ID:          "skeleton-legacy-wrap",
			State:       types.ItemApplied,
			Description: "wrapped existing source in legacy/ Maven module; original root pom.xml preserved at legacy/legacy-original-pom.xml",
		},
		{
			ID:          "skeleton-root-files",
			State:       types.ItemApplied,
			Description: ".trabuco.json + .editorconfig written; .gitignore augmented with .trabuco-migration/ exclusion",
		},
	}

	return &specialists.Output{
		Phase:   types.PhaseSkeleton,
		Items:   items,
		Summary: "Multi-module skeleton bootstrapped in migration mode. Enforcement deferred to Phase 12. Legacy code now lives in legacy/ — should still build with mvn compile -pl :legacy.",
	}, nil
}

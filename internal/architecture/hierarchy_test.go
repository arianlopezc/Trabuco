package architecture

import (
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

func TestFilterFor_FullProject(t *testing.T) {
	full := []string{
		config.ModuleModel,
		config.ModuleJobs,
		config.ModuleSQLDatastore,
		config.ModuleShared,
		config.ModuleAPI,
		config.ModuleWorker,
		config.ModuleEvents,
		config.ModuleEventConsumer,
		config.ModuleAIAgent,
	}
	h := FilterFor(full)
	if len(h.Layers) != 5 {
		t.Fatalf("expected 5 layers (Foundation/Contracts/Persistence/BusinessLogic/Edge), got %d", len(h.Layers))
	}
	wantLayers := []string{"Foundation", "Contracts", "Persistence", "BusinessLogic", "Edge"}
	for i, name := range wantLayers {
		if h.Layers[i].Name != name {
			t.Errorf("layer[%d] = %q, want %q", i, h.Layers[i].Name, name)
		}
	}
	// Edge layer should hold all four edge modules in canonical order.
	if got := len(h.Layers[4].Modules); got != 4 {
		t.Errorf("Edge layer has %d modules, want 4", got)
	}
}

func TestFilterFor_ApiOnly(t *testing.T) {
	// Pure API + datastore + shared — no Jobs/Events should appear.
	h := FilterFor([]string{
		config.ModuleModel,
		config.ModuleSQLDatastore,
		config.ModuleShared,
		config.ModuleAPI,
	})
	if len(h.Layers) != 4 {
		t.Fatalf("expected 4 layers (no Contracts), got %d", len(h.Layers))
	}
	for _, layer := range h.Layers {
		if layer.Name == "Contracts" {
			t.Errorf("Contracts layer should not appear in API-only project")
		}
	}
}

func TestFilterFor_ModelOnly(t *testing.T) {
	h := FilterFor([]string{config.ModuleModel})
	if len(h.Layers) != 1 || h.Layers[0].Name != "Foundation" {
		t.Errorf("Model-only should produce just Foundation, got %v", h.Layers)
	}
}

func TestFilterFor_JobsAppearsWithWorker(t *testing.T) {
	h := FilterFor([]string{config.ModuleModel, config.ModuleJobs, config.ModuleWorker})
	if len(h.Layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(h.Layers))
	}
	contracts := h.Layers[1]
	if contracts.Name != "Contracts" {
		t.Errorf("layer[1] = %q, want Contracts", contracts.Name)
	}
	if len(contracts.Modules) != 1 || contracts.Modules[0].Name != config.ModuleJobs {
		t.Errorf("Contracts should hold only Jobs when EventConsumer absent, got %v", contracts.Modules)
	}
}

func TestFilterFor_EventsAppearsWithEventConsumer(t *testing.T) {
	h := FilterFor([]string{config.ModuleModel, config.ModuleEvents, config.ModuleEventConsumer})
	if len(h.Layers) != 3 {
		t.Fatalf("expected 3 layers, got %d", len(h.Layers))
	}
	contracts := h.Layers[1]
	if len(contracts.Modules) != 1 || contracts.Modules[0].Name != config.ModuleEvents {
		t.Errorf("Contracts should hold only Events when Worker absent, got %v", contracts.Modules)
	}
}

func TestFilterFor_BothContractModules(t *testing.T) {
	h := FilterFor([]string{
		config.ModuleModel, config.ModuleJobs, config.ModuleEvents,
		config.ModuleWorker, config.ModuleEventConsumer,
	})
	contracts := h.Layers[1]
	if len(contracts.Modules) != 2 {
		t.Errorf("Contracts should hold both Jobs and Events when Worker + EventConsumer selected, got %v", contracts.Modules)
	}
	// Order matters: Jobs before Events as registered.
	if contracts.Modules[0].Name != config.ModuleJobs || contracts.Modules[1].Name != config.ModuleEvents {
		t.Errorf("Contracts module order = %v, want [Jobs, Events]", contracts.Modules)
	}
}

func TestStagesFor_FullProject(t *testing.T) {
	stages := StagesFor([]string{
		config.ModuleModel, config.ModuleJobs, config.ModuleSQLDatastore,
		config.ModuleShared, config.ModuleAPI, config.ModuleWorker,
	})
	want := []string{
		"Stage 1 — Model",
		"Stage 2 — Jobs",
		"Stage 3 — SQLDatastore",
		"Stage 4 — Shared",
		"Stage 5 — API | Worker",
	}
	if len(stages) != len(want) {
		t.Fatalf("got %d stages, want %d (%v)", len(stages), len(want), stages)
	}
	for i := range stages {
		if stages[i] != want[i] {
			t.Errorf("stage[%d] = %q, want %q", i, stages[i], want[i])
		}
	}
}

func TestStagesFor_ApiOnly(t *testing.T) {
	stages := StagesFor([]string{
		config.ModuleModel, config.ModuleSQLDatastore, config.ModuleShared, config.ModuleAPI,
	})
	want := []string{
		"Stage 1 — Model",
		"Stage 2 — SQLDatastore",
		"Stage 3 — Shared",
		"Stage 4 — API",
	}
	if len(stages) != 4 {
		t.Fatalf("got %d stages, want 4", len(stages))
	}
	for i, s := range stages {
		if s != want[i] {
			t.Errorf("stage[%d] = %q, want %q", i, s, want[i])
		}
	}
}

func TestStagesFor_PipeJoinSorted(t *testing.T) {
	// Edge with multiple siblings should pipe-join in canonical order.
	stages := StagesFor([]string{
		config.ModuleModel, config.ModuleShared,
		config.ModuleAPI, config.ModuleAIAgent,
	})
	last := stages[len(stages)-1]
	if !strings.Contains(last, "API") || !strings.Contains(last, "AIAgent") {
		t.Errorf("last stage missing both API and AIAgent: %q", last)
	}
	if !strings.Contains(last, "|") {
		t.Errorf("last stage should pipe-join siblings: %q", last)
	}
}

func TestAll_ContainsEveryModule(t *testing.T) {
	h := All()
	got := map[string]bool{}
	for _, layer := range h.Layers {
		for _, m := range layer.Modules {
			got[m.Name] = true
		}
	}
	want := []string{
		config.ModuleModel,
		config.ModuleJobs,
		config.ModuleSQLDatastore,
		config.ModuleNoSQLDatastore,
		config.ModuleShared,
		config.ModuleAPI,
		config.ModuleWorker,
		config.ModuleEvents,
		config.ModuleEventConsumer,
		config.ModuleAIAgent,
	}
	for _, m := range want {
		if !got[m] {
			t.Errorf("All() missing module %q", m)
		}
	}
}

func TestModuleStageHints(t *testing.T) {
	// Every module must have a non-empty StageHint so rendered docs
	// always have prose for each stage.
	for _, layer := range hierarchy {
		for _, m := range layer.Modules {
			if m.StageHint == "" {
				t.Errorf("module %q has empty StageHint", m.Name)
			}
			if m.Purpose == "" {
				t.Errorf("module %q has empty Purpose", m.Name)
			}
		}
	}
}

package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

func TestNew_AllPhasesPending(t *testing.T) {
	s := New("1.10.0-test")
	for _, p := range types.AllPhases() {
		rec, ok := s.Phases[p]
		if !ok {
			t.Errorf("phase %s missing from new state", p)
			continue
		}
		if rec.State != types.PhasePending {
			t.Errorf("phase %s state = %s, want pending", p, rec.State)
		}
	}
}

func TestSaveLoad_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	s := New("1.10.0-test")
	s.SourceConfig.Framework = "spring-boot-2.7"
	s.TargetConfig.Modules = []string{"Model", "API"}
	s.Phases[types.PhaseAssessment].State = types.PhaseCompleted
	s.Blockers = append(s.Blockers, BlockerRecord{
		Phase: types.PhaseDatastore,
		Code:  types.BlockerFKRequired,
		File:  "src/main/java/com/x/User.java",
	})

	if err := Save(dir, s); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.SourceConfig.Framework != "spring-boot-2.7" {
		t.Errorf("Framework = %q, want spring-boot-2.7", loaded.SourceConfig.Framework)
	}
	if len(loaded.TargetConfig.Modules) != 2 {
		t.Errorf("Modules = %v, want [Model API]", loaded.TargetConfig.Modules)
	}
	if loaded.Phases[types.PhaseAssessment].State != types.PhaseCompleted {
		t.Errorf("PhaseAssessment state = %s, want completed", loaded.Phases[types.PhaseAssessment].State)
	}
	if len(loaded.Blockers) != 1 || loaded.Blockers[0].Code != types.BlockerFKRequired {
		t.Errorf("blocker not roundtripped: %+v", loaded.Blockers)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	dir := t.TempDir()
	if _, err := Load(dir); err == nil {
		t.Error("Load on empty dir should error")
	}
}

func TestLoad_WrongSchemaVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(MigrationDirPath(dir), 0o755); err != nil {
		t.Fatal(err)
	}
	bogus := []byte(`{"schemaVersion": 999}`)
	if err := os.WriteFile(StatePath(dir), bogus, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil {
		t.Error("Load with bogus schema version should error")
	}
}

func TestExists(t *testing.T) {
	dir := t.TempDir()
	if Exists(dir) {
		t.Error("empty dir should not have state")
	}
	s := New("1.10")
	if err := Save(dir, s); err != nil {
		t.Fatal(err)
	}
	if !Exists(dir) {
		t.Error("after Save, Exists should return true")
	}
}

func TestPathHelpers(t *testing.T) {
	repo := "/tmp/repo"
	cases := map[string]string{
		MigrationDirPath(repo):                          filepath.Join(repo, ".trabuco-migration"),
		StatePath(repo):                                 filepath.Join(repo, ".trabuco-migration", "state.json"),
		LockPath(repo):                                  filepath.Join(repo, ".trabuco-migration", "lock.json"),
		AssessmentPath(repo):                            filepath.Join(repo, ".trabuco-migration", "assessment.json"),
		CompletionReportPath(repo):                      filepath.Join(repo, ".trabuco-migration", "completion-report.md"),
		PhaseInputPath(repo, types.PhaseModel):          filepath.Join(repo, ".trabuco-migration", "phase-2-input.json"),
		PhaseOutputPath(repo, types.PhaseDeployment):    filepath.Join(repo, ".trabuco-migration", "phase-10-output.json"),
		PhaseDiffPath(repo, types.PhaseActivation):      filepath.Join(repo, ".trabuco-migration", "phase-12-diff.patch"),
		PhaseReportPath(repo, types.PhaseFinalization):  filepath.Join(repo, ".trabuco-migration", "phase-13-report.md"),
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("path = %q, want %q", got, want)
		}
	}
}

func TestLock_AcquireRelease(t *testing.T) {
	dir := t.TempDir()
	if err := AcquireLock(dir, "cli"); err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}
	// Re-acquiring while held by self should fail.
	if err := AcquireLock(dir, "cli"); err == nil {
		t.Error("second AcquireLock should fail while lock is held by live process")
	}
	if err := ReleaseLock(dir); err != nil {
		t.Errorf("ReleaseLock: %v", err)
	}
	// Releasing again should be idempotent.
	if err := ReleaseLock(dir); err != nil {
		t.Errorf("second ReleaseLock: %v", err)
	}
}

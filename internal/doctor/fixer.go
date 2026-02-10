package doctor

import (
	"fmt"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/fatih/color"
)

// Fixer handles auto-fixing of issues found by the doctor
type Fixer struct {
	projectPath string
	metadata    *config.ProjectMetadata
}

// NewFixer creates a new Fixer
func NewFixer(projectPath string, metadata *config.ProjectMetadata) *Fixer {
	return &Fixer{
		projectPath: projectPath,
		metadata:    metadata,
	}
}

// FixAll attempts to fix all auto-fixable issues
func (f *Fixer) FixAll(result *DoctorResult) []FixResult {
	var results []FixResult

	for _, check := range result.GetFixableChecks() {
		fixResult := f.Fix(check)
		results = append(results, fixResult)
	}

	return results
}

// Fix attempts to fix a single check
func (f *Fixer) Fix(check CheckResult) FixResult {
	// Find the checker implementation
	for _, checker := range GetAllChecks() {
		if checker.ID() == check.ID {
			err := checker.Fix(f.projectPath, f.metadata)
			if err != nil {
				return FixResult{
					CheckID: check.ID,
					Name:    check.Name,
					Success: false,
					Error:   err.Error(),
				}
			}
			return FixResult{
				CheckID: check.ID,
				Name:    check.Name,
				Success: true,
			}
		}
	}

	return FixResult{
		CheckID: check.ID,
		Name:    check.Name,
		Success: false,
		Error:   "checker not found",
	}
}

// FixResult represents the result of a fix attempt
type FixResult struct {
	CheckID string `json:"checkId"`
	Name    string `json:"name"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// PrintFixResults prints the results of fix attempts
func PrintFixResults(results []FixResult) {
	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	if len(results) == 0 {
		fmt.Println("No issues to fix.")
		return
	}

	fmt.Println()
	fmt.Println("Fix Results:")
	fmt.Println()

	successCount := 0
	for _, r := range results {
		if r.Success {
			green.Printf("  \u2713 ")
			fmt.Printf("Fixed: %s\n", r.Name)
			successCount++
		} else {
			red.Printf("  \u2717 ")
			fmt.Printf("Failed: %s\n", r.Name)
			if r.Error != "" {
				fmt.Printf("      Error: %s\n", r.Error)
			}
		}
	}

	fmt.Println()
	if successCount == len(results) {
		green.Printf("All %d issues fixed successfully.\n", successCount)
	} else {
		fmt.Printf("%d of %d issues fixed.\n", successCount, len(results))
	}
}

// RegenerateMetadata regenerates .trabuco.json from POM
func RegenerateMetadata(projectPath string) (*config.ProjectMetadata, error) {
	metadata, err := InferFromPOM(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to infer metadata from POM: %w", err)
	}

	if err := config.SaveMetadata(projectPath, metadata); err != nil {
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return metadata, nil
}

// SyncMetadataWithPOM updates .trabuco.json to match POM modules
func SyncMetadataWithPOM(projectPath string, metadata *config.ProjectMetadata) error {
	pomModules, err := GetModulesFromPOM(projectPath)
	if err != nil {
		return fmt.Errorf("failed to read POM modules: %w", err)
	}

	metadata.Modules = pomModules
	metadata.UpdateGeneratedAt()

	if err := config.SaveMetadata(projectPath, metadata); err != nil {
		return fmt.Errorf("failed to save metadata: %w", err)
	}

	return nil
}

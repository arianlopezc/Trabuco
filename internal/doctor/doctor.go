package doctor

import (
	"os"
	"path/filepath"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// Doctor orchestrates health checks for a Trabuco project
type Doctor struct {
	checks      []Checker
	projectPath string
	version     string
}

// New creates a new Doctor with all checks
func New(projectPath string, version string) *Doctor {
	return &Doctor{
		checks:      GetAllChecks(),
		projectPath: projectPath,
		version:     version,
	}
}

// NewWithChecks creates a Doctor with specific checks
func NewWithChecks(projectPath string, version string, checks []Checker) *Doctor {
	return &Doctor{
		checks:      checks,
		projectPath: projectPath,
		version:     version,
	}
}

// Run executes all health checks and returns the result
func (d *Doctor) Run() (*DoctorResult, error) {
	absPath, err := filepath.Abs(d.projectPath)
	if err != nil {
		absPath = d.projectPath
	}

	result := &DoctorResult{
		Location:       absPath,
		TrabucoVersion: d.version,
		Checks:         make([]CheckResult, 0),
	}

	// Try to detect project and load metadata
	metadata, err := DetectProject(d.projectPath)
	if err == nil {
		result.Project = metadata.ProjectName
		result.Metadata = metadata
		if metadata.Version != "" {
			result.TrabucoVersion = metadata.Version
		}
	} else {
		// Try to get project name from POM
		pom, pomErr := ParseParentPOM(filepath.Join(d.projectPath, "pom.xml"))
		if pomErr == nil {
			result.Project = extractProjectName(pom.ArtifactID)
		} else {
			// Use directory name as fallback
			result.Project = filepath.Base(absPath)
		}
	}

	// Run all checks
	for _, check := range d.checks {
		checkResult := check.Check(d.projectPath, metadata)
		result.Checks = append(result.Checks, checkResult)
	}

	// Compute summary
	result.ComputeSummary()

	return result, nil
}

// RunAndFix executes checks and attempts to fix any issues
func (d *Doctor) RunAndFix() (*DoctorResult, []FixResult, error) {
	// First run to identify issues
	result, err := d.Run()
	if err != nil {
		return nil, nil, err
	}

	// If no fixable issues, return early
	fixable := result.GetFixableChecks()
	if len(fixable) == 0 {
		return result, nil, nil
	}

	// Create fixer and fix issues
	fixer := NewFixer(d.projectPath, result.Metadata)
	fixResults := fixer.FixAll(result)

	// Re-run checks to get updated status
	result, err = d.Run()
	if err != nil {
		return nil, fixResults, err
	}

	return result, fixResults, nil
}

// RunCategory executes checks for a specific category
func (d *Doctor) RunCategory(category string) (*DoctorResult, error) {
	categoryChecks := GetChecksByCategory(category)
	if len(categoryChecks) == 0 {
		// If invalid category, run all checks
		return d.Run()
	}

	// Create a doctor with only the category checks
	categoryDoctor := NewWithChecks(d.projectPath, d.version, categoryChecks)
	return categoryDoctor.Run()
}

// Validate runs checks and returns an error if the project has errors
// This is useful for the add command to validate before proceeding
func (d *Doctor) Validate() error {
	result, err := d.Run()
	if err != nil {
		return err
	}

	if result.HasErrors() {
		return &ValidationError{Result: result}
	}

	return nil
}

// ValidationError is returned when project validation fails
type ValidationError struct {
	Result *DoctorResult
}

func (e *ValidationError) Error() string {
	return "project validation failed"
}

// QuickCheck performs a fast check to see if this is a valid Trabuco project
// Returns metadata if valid, error otherwise
func QuickCheck(projectPath string) (*config.ProjectMetadata, error) {
	// Check pom.xml exists
	pomPath := filepath.Join(projectPath, "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return nil, &NotMavenProjectError{Path: projectPath}
	}

	// Try to detect project
	metadata, err := DetectProject(projectPath)
	if err != nil {
		return nil, &NotTrabucoProjectError{Path: projectPath, Reason: err.Error()}
	}

	return metadata, nil
}

// NotMavenProjectError is returned when the directory is not a Maven project
type NotMavenProjectError struct {
	Path string
}

func (e *NotMavenProjectError) Error() string {
	return "not a Maven project (pom.xml not found)"
}

// NotTrabucoProjectError is returned when the directory is not a Trabuco project
type NotTrabucoProjectError struct {
	Path   string
	Reason string
}

func (e *NotTrabucoProjectError) Error() string {
	return "not a Trabuco project: " + e.Reason
}

// GetProjectMetadata returns the project metadata, loading or inferring as needed
func GetProjectMetadata(projectPath string) (*config.ProjectMetadata, error) {
	// Try to load from .trabuco.json
	if config.MetadataExists(projectPath) {
		return config.LoadMetadata(projectPath)
	}

	// Infer from POM
	return InferFromPOM(projectPath)
}

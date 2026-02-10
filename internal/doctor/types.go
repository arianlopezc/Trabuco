package doctor

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/fatih/color"
)

// Severity represents the severity level of a check result
type Severity int

const (
	// SeverityPass indicates the check passed successfully
	SeverityPass Severity = iota
	// SeverityWarn indicates a warning that doesn't block operations
	SeverityWarn
	// SeverityError indicates an error that blocks operations like 'add'
	SeverityError
)

// String returns the string representation of a Severity
func (s Severity) String() string {
	switch s {
	case SeverityPass:
		return "PASS"
	case SeverityWarn:
		return "WARN"
	case SeverityError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// MarshalJSON implements json.Marshaler for Severity
func (s Severity) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// CheckResult represents the result of a single health check
type CheckResult struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Status     Severity `json:"status"`
	Message    string   `json:"message,omitempty"`
	Details    []string `json:"details,omitempty"`
	FixAction  string   `json:"fixAction,omitempty"`
	CanAutoFix bool     `json:"canAutoFix"`
}

// DoctorResult represents the complete result of running all health checks
type DoctorResult struct {
	Project        string               `json:"project"`
	Location       string               `json:"location"`
	TrabucoVersion string               `json:"trabucoVersion"`
	Status         string               `json:"status"`
	Summary        DoctorSummary        `json:"summary"`
	Checks         []CheckResult        `json:"checks"`
	Metadata       *config.ProjectMetadata `json:"-"` // Internal use, not serialized
}

// DoctorSummary contains summary statistics for the doctor run
type DoctorSummary struct {
	Passed   int `json:"passed"`
	Warnings int `json:"warnings"`
	Errors   int `json:"errors"`
}

// HasErrors returns true if any check resulted in an error
func (r *DoctorResult) HasErrors() bool {
	return r.Summary.Errors > 0
}

// HasWarnings returns true if any check resulted in a warning
func (r *DoctorResult) HasWarnings() bool {
	return r.Summary.Warnings > 0
}

// IsHealthy returns true if all checks passed (no errors or warnings)
func (r *DoctorResult) IsHealthy() bool {
	return !r.HasErrors() && !r.HasWarnings()
}

// ComputeSummary calculates the summary statistics from checks
func (r *DoctorResult) ComputeSummary() {
	r.Summary = DoctorSummary{}
	for _, check := range r.Checks {
		switch check.Status {
		case SeverityPass:
			r.Summary.Passed++
		case SeverityWarn:
			r.Summary.Warnings++
		case SeverityError:
			r.Summary.Errors++
		}
	}

	// Set overall status
	if r.HasErrors() {
		r.Status = "UNHEALTHY"
	} else if r.HasWarnings() {
		r.Status = "WARNINGS"
	} else {
		r.Status = "HEALTHY"
	}
}

// ToJSON serializes the result to JSON
func (r *DoctorResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// PrintSummary prints a formatted summary to stdout
func (r *DoctorResult) PrintSummary(verbose bool) {
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	bold := color.New(color.Bold)
	cyan := color.New(color.FgCyan)

	// Header
	fmt.Println()
	bold.Println("Trabuco Project Health Check")
	fmt.Println(strings.Repeat("\u2501", 28))
	fmt.Println()

	// Project info
	cyan.Printf("Project: ")
	fmt.Println(r.Project)
	cyan.Printf("Location: ")
	fmt.Println(r.Location)
	if r.TrabucoVersion != "" {
		cyan.Printf("Trabuco Version: ")
		fmt.Println(r.TrabucoVersion)
	}
	fmt.Println()

	fmt.Println("Running checks...")
	fmt.Println()

	// Check results
	for _, check := range r.Checks {
		// Skip passed checks in non-verbose mode unless there are no issues
		if !verbose && check.Status == SeverityPass && (r.HasErrors() || r.HasWarnings()) {
			continue
		}

		switch check.Status {
		case SeverityPass:
			green.Printf("  \u2713 ")
			fmt.Println(check.Name)
		case SeverityWarn:
			yellow.Printf("  \u26a0 ")
			fmt.Println(check.Name)
			if check.Message != "" {
				fmt.Printf("      %s\n", check.Message)
			}
			for _, detail := range check.Details {
				fmt.Printf("      %s\n", detail)
			}
			if check.FixAction != "" {
				yellow.Printf("      Run 'trabuco doctor --fix' to %s\n", check.FixAction)
			}
		case SeverityError:
			red.Printf("  \u2717 ")
			fmt.Println(check.Name)
			if check.Message != "" {
				fmt.Printf("      %s\n", check.Message)
			}
			for _, detail := range check.Details {
				fmt.Printf("      %s\n", detail)
			}
		}
	}

	// Footer
	fmt.Println()
	fmt.Println(strings.Repeat("\u2501", 28))

	// Status
	switch r.Status {
	case "HEALTHY":
		green.Printf("Status: ")
		green.Println("HEALTHY")
		fmt.Printf("All %d checks passed\n", r.Summary.Passed)
	case "WARNINGS":
		yellow.Printf("Status: ")
		yellow.Println("WARNINGS")
		fmt.Printf("%d passed, %d warnings\n", r.Summary.Passed, r.Summary.Warnings)
		fmt.Println()
		yellow.Println("Run 'trabuco doctor --fix' to auto-fix warnings.")
	case "UNHEALTHY":
		red.Printf("Status: ")
		red.Println("UNHEALTHY")
		fmt.Printf("%d passed, %d warnings, %d errors\n", r.Summary.Passed, r.Summary.Warnings, r.Summary.Errors)
		fmt.Println()
		red.Println("Errors must be fixed before running 'trabuco add'.")
		if r.HasWarnings() {
			yellow.Println("Run 'trabuco doctor --fix' to auto-fix warnings.")
		}
	}
}

// PrintWarnings prints just the warnings (used by add command)
func (r *DoctorResult) PrintWarnings() {
	yellow := color.New(color.FgYellow)
	for _, check := range r.Checks {
		if check.Status == SeverityWarn {
			yellow.Printf("  \u26a0 ")
			fmt.Println(check.Name)
			if check.Message != "" {
				fmt.Printf("      %s\n", check.Message)
			}
		}
	}
}

// GetFixableChecks returns checks that can be auto-fixed
func (r *DoctorResult) GetFixableChecks() []CheckResult {
	var fixable []CheckResult
	for _, check := range r.Checks {
		if check.CanAutoFix && check.Status != SeverityPass {
			fixable = append(fixable, check)
		}
	}
	return fixable
}

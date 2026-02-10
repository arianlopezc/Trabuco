package cli

import (
	"fmt"
	"os"

	"github.com/arianlopezc/Trabuco/internal/doctor"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	doctorVerbose bool
	doctorFix     bool
	doctorJSON    bool
	doctorCheck   string
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate Trabuco project structure and health",
	Long: `The doctor command validates that a project is a healthy Trabuco project.

It checks for:
  - Valid project structure (pom.xml exists)
  - Trabuco project detection (.trabuco.json or structure match)
  - Metadata file validity
  - Parent POM configuration
  - Module directories and POMs
  - Configuration consistency
  - Docker Compose synchronization

Examples:
  trabuco doctor              Run all checks
  trabuco doctor --verbose    Show all checks (not just failures)
  trabuco doctor --fix        Auto-fix issues that can be fixed
  trabuco doctor --json       Output as JSON (for scripting)
  trabuco doctor --check=metadata  Check specific category`,
	Run: runDoctor,
}

func init() {
	doctorCmd.Flags().BoolVarP(&doctorVerbose, "verbose", "v", false, "Show all checks, not just failures")
	doctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Auto-fix issues that can be fixed")
	doctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output as JSON")
	doctorCmd.Flags().StringVar(&doctorCheck, "check", "", "Run specific check category (structure, metadata, consistency)")
}

func runDoctor(cmd *cobra.Command, args []string) {
	// Get current working directory
	projectPath, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not get current directory: %v\n", err)
		os.Exit(1)
	}

	// Create doctor
	doc := doctor.New(projectPath, Version)

	var result *doctor.DoctorResult
	var fixResults []doctor.FixResult

	if doctorFix {
		// Run with fix
		result, fixResults, err = doc.RunAndFix()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running doctor: %v\n", err)
			os.Exit(1)
		}
	} else if doctorCheck != "" {
		// Run specific category
		result, err = doc.RunCategory(doctorCheck)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running doctor: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Run all checks
		result, err = doc.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error running doctor: %v\n", err)
			os.Exit(1)
		}
	}

	// Output results
	if doctorJSON {
		jsonOutput, err := result.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(jsonOutput))
	} else {
		result.PrintSummary(doctorVerbose)

		// Print fix results if we did fixes
		if len(fixResults) > 0 {
			doctor.PrintFixResults(fixResults)
		}
	}

	// Exit with appropriate code
	if result.HasErrors() {
		os.Exit(1)
	}

	// Show hint if there are warnings and we didn't fix
	if result.HasWarnings() && !doctorFix && !doctorJSON {
		fmt.Println()
		yellow := color.New(color.FgYellow)
		yellow.Println("Tip: Run 'trabuco doctor --fix' to auto-fix warnings.")
	}
}

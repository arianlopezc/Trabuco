package cli

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Java project",
	Long: `Initialize a new Java multi-module Maven project.

This command will interactively prompt you for:
  - Project name
  - Group ID (e.g., com.company.project)
  - Modules to include
  - Database type (PostgreSQL, MySQL, or Generic)
  - Additional options (Docker, GitHub Actions, etc.)`,
	Run: runInit,
}

func runInit(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)

	cyan.Println("\nWelcome to Trabuco - Java Project Generator\n")

	// TODO: Implement prompts (Stage 4)
	// TODO: Implement generation (Stage 9)

	green.Println("Project generation not yet implemented.")
	fmt.Println("This will be completed in upcoming stages.")
}

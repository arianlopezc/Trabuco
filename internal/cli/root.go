package cli

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "trabuco",
	Short: "Java project generator CLI",
	Long: `Trabuco is a CLI tool that generates production-ready Java
multi-module Maven projects based on proven patterns.

It creates a complete project structure with:
  - Model module (DTOs, Entities, Enums)
  - SQLDatastore module (Repositories, Flyway migrations)
  - Shared module (Services, Circuit breaker)
  - API module (REST endpoints, Validation)

Plus Docker configs, GitHub Actions, and IntelliJ run configurations.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(migrateCmd)
	rootCmd.AddCommand(authCmd)
}

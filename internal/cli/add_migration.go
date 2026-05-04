package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addMigrationDescription string
	addMigrationModule      string
	addMigrationDryRun      bool
	addMigrationJSON        bool
)

var addMigrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Generate a new Flyway migration file",
	Long: `Generate a new empty Flyway migration under
SQLDatastore/src/main/resources/db/migration/V{N}__{description}.sql

The next available V{N} is picked automatically. Description is
snake_cased into the filename suffix. The body is a TODO header that
you (or the coding agent) edit to add the actual DDL.

This command is addition-only: it never edits or deletes files. If a
migration at the target path already exists, the command refuses to
clobber it; delete the file first to regenerate.

Examples:
  trabuco add migration --description="add orders table"
  trabuco add migration --description="Add Customer Profiles" --dry-run
  trabuco add migration --description=add_users --json`,
	Run: runAddMigration,
}

func init() {
	addMigrationCmd.Flags().StringVar(&addMigrationDescription, "description", "", "Migration description (required; snake_cased into the filename)")
	addMigrationCmd.Flags().StringVar(&addMigrationModule, "module", "SQLDatastore", "Module the migration belongs to (only SQLDatastore today)")
	addMigrationCmd.Flags().BoolVar(&addMigrationDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addMigrationCmd.Flags().BoolVar(&addMigrationJSON, "json", false, "Emit machine-readable JSON output")
	_ = addMigrationCmd.MarkFlagRequired("description")
	addCmd.AddCommand(addMigrationCmd)
}

func runAddMigration(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addMigrationJSON)
		os.Exit(1)
	}

	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addMigrationJSON)
		os.Exit(1)
	}
	ctx.DryRun = addMigrationDryRun

	result, err := addgen.GenerateMigration(ctx, addgen.MigrationOpts{
		Description: addMigrationDescription,
		Module:      addMigrationModule,
	})
	if err != nil {
		printAddError(err, addMigrationJSON)
		os.Exit(1)
	}

	printAddResult(result, addMigrationDryRun, addMigrationJSON)
}

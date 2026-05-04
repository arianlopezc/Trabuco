package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addServiceEntity string
	addServiceDryRun bool
	addServiceJSON   bool
)

var addServiceCmd = &cobra.Command{
	Use:   "service <Name>",
	Short: "Generate a Spring @Service class with constructor injection",
	Long: `Generate Shared/.../service/{Name}.java — a @Service class with
constructor injection. When --entity is provided, the constructor
wires in the matching repository (Spring Data JDBC for SQL,
MongoRepository for Mongo).

The class body is intentionally a stub; replace doSomething() with
your real business operations.

Examples:
  trabuco add service OrderService --entity=Order
  trabuco add service NotificationService                # no repository injection`,
	Args: cobra.ExactArgs(1),
	Run:  runAddService,
}

func init() {
	addServiceCmd.Flags().StringVar(&addServiceEntity, "entity", "", "Inject {Entity}Repository (SQL) or {Entity}DocumentRepository (Mongo)")
	addServiceCmd.Flags().BoolVar(&addServiceDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addServiceCmd.Flags().BoolVar(&addServiceJSON, "json", false, "Emit machine-readable JSON output")
	addCmd.AddCommand(addServiceCmd)
}

func runAddService(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addServiceJSON)
		os.Exit(1)
	}
	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addServiceJSON)
		os.Exit(1)
	}
	ctx.DryRun = addServiceDryRun

	result, err := addgen.GenerateService(ctx, addgen.ServiceOpts{
		Name:   args[0],
		Entity: addServiceEntity,
	})
	if err != nil {
		printAddError(err, addServiceJSON)
		os.Exit(1)
	}
	printAddResult(result, addServiceDryRun, addServiceJSON)
}

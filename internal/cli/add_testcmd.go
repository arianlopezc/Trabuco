package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addTestType    string
	addTestModule  string
	addTestSubPkg  string
	addTestDryRun  bool
	addTestJSON    bool
)

var addTestCmd = &cobra.Command{
	Use:   "test <target>",
	Short: "Generate a test class skeleton (unit, integration, or repository)",
	Long: `Generate an empty test class for an existing class. The CLI does not
analyze the target — it produces a deterministic skeleton with the
correct imports, annotations, and Testcontainers wiring (for
repository tests). The agent fills in the @Test methods.

Type:
  unit         (default) JUnit 5 + Mockito @ExtendWith(MockitoExtension.class)
  integration  @SpringBootTest, file ends in *IT.java for Failsafe
  repository   @DataJdbcTest (SQL) or @DataMongoTest (Mongo) with the
               right Testcontainers @ServiceConnection wiring

Subpackage is inferred from the Target's suffix (Service → service,
Controller → controller, etc.) unless --subpackage overrides it.

Examples:
  trabuco add test OrderService --module=Shared
  trabuco add test OrderController --module=API --type=integration
  trabuco add test OrderRepository --module=SQLDatastore --type=repository
  trabuco add test ProcessShipmentJobHandler --module=Worker
  trabuco add test MyUtil --module=Shared --subpackage=util`,
	Args: cobra.ExactArgs(1),
	Run:  runAddTest,
}

func init() {
	addTestCmd.Flags().StringVar(&addTestType, "type", "unit", "Test shape: unit | integration | repository")
	addTestCmd.Flags().StringVar(&addTestModule, "module", "", "Module the test belongs to (required)")
	addTestCmd.Flags().StringVar(&addTestSubPkg, "subpackage", "", "Override the inferred subpackage")
	addTestCmd.Flags().BoolVar(&addTestDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addTestCmd.Flags().BoolVar(&addTestJSON, "json", false, "Emit machine-readable JSON output")
	_ = addTestCmd.MarkFlagRequired("module")
	addCmd.AddCommand(addTestCmd)
}

func runAddTest(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addTestJSON)
		os.Exit(1)
	}
	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addTestJSON)
		os.Exit(1)
	}
	ctx.DryRun = addTestDryRun

	result, err := addgen.GenerateTest(ctx, addgen.TestOpts{
		Target:     args[0],
		Type:       addTestType,
		Module:     addTestModule,
		Subpackage: addTestSubPkg,
	})
	if err != nil {
		printAddError(err, addTestJSON)
		os.Exit(1)
	}
	printAddResult(result, addTestDryRun, addTestJSON)
}

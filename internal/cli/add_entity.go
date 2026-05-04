package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addEntityFields    string
	addEntityModule    string
	addEntityTableName string
	addEntityDryRun    bool
	addEntityJSON      bool
)

var addEntityCmd = &cobra.Command{
	Use:   "entity <Name>",
	Short: "Generate an entity (interface + record + repository + migration)",
	Long: `Generate the full entity bundle for an existing project.

For SQLDatastore projects, four files are emitted: the Immutables
interface (Model module), the Spring Data JDBC record, the
CrudRepository, and a Flyway migration with the CREATE TABLE DDL.

For NoSQLDatastore (Mongo) projects, three files: the Immutables
interface, the Spring Data Mongo document, and the MongoRepository.

Field syntax: --fields="name:type[?],..." where type is one of:
  string text integer long decimal boolean instant localdate
  uuid json bytes  enum:EnumName

Trailing "?" marks the field nullable. Enum fields auto-emit a
placeholder enum class in Model/.../entities/ if it doesn't exist.

Examples:
  trabuco add entity Order \
      --fields="customerId:string,total:decimal,placedAt:instant"

  trabuco add entity ShippingAddress \
      --fields="street:string,city:string,zip:string,country:string,verified:boolean"

  trabuco add entity Invoice \
      --fields="orderId:string,amount:decimal,status:enum:InvoiceStatus,notes:text?"

  trabuco add entity Person --table-name=people --fields="firstName:string,lastName:string"`,
	Args: cobra.ExactArgs(1),
	Run:  runAddEntity,
}

func init() {
	addEntityCmd.Flags().StringVar(&addEntityFields, "fields", "", `Comma-separated field spec (required), e.g. "customerId:string,total:decimal,placedAt:instant,notes:text?"`)
	addEntityCmd.Flags().StringVar(&addEntityModule, "module", "", "Force SQLDatastore or NoSQLDatastore (default: auto-detect from project)")
	addEntityCmd.Flags().StringVar(&addEntityTableName, "table-name", "", "Override the auto-derived plural snake_case table/collection name")
	addEntityCmd.Flags().BoolVar(&addEntityDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addEntityCmd.Flags().BoolVar(&addEntityJSON, "json", false, "Emit machine-readable JSON output")
	_ = addEntityCmd.MarkFlagRequired("fields")
	addCmd.AddCommand(addEntityCmd)
}

func runAddEntity(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addEntityJSON)
		os.Exit(1)
	}
	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addEntityJSON)
		os.Exit(1)
	}
	ctx.DryRun = addEntityDryRun

	result, err := addgen.GenerateEntity(ctx, addgen.EntityOpts{
		Name:      args[0],
		Fields:    addEntityFields,
		Module:    addEntityModule,
		TableName: addEntityTableName,
	})
	if err != nil {
		printAddError(err, addEntityJSON)
		os.Exit(1)
	}
	printAddResult(result, addEntityDryRun, addEntityJSON)
}

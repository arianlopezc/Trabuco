package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addEndpointType   string
	addEndpointPath   string
	addEndpointDryRun bool
	addEndpointJSON   bool
)

var addEndpointCmd = &cobra.Command{
	Use:   "endpoint <Name>",
	Short: "Generate a REST controller skeleton (plain or CRUD)",
	Long: `Generate API/.../controller/{Name}Controller.java.

  --type=plain  (default) empty controller with @RestController + @RequestMapping
  --type=crud   five method stubs: POST /, GET /{id}, GET /, PUT /{id}, DELETE /{id}

The URL path is auto-derived as /api/{plural-lower-snake} unless --path overrides.

Examples:
  trabuco add endpoint Order --type=crud
  trabuco add endpoint Health --path=/healthz`,
	Args: cobra.ExactArgs(1),
	Run:  runAddEndpoint,
}

func init() {
	addEndpointCmd.Flags().StringVar(&addEndpointType, "type", "plain", "Controller shape: plain | crud")
	addEndpointCmd.Flags().StringVar(&addEndpointPath, "path", "", "Override the URL path (default: /api/{plural})")
	addEndpointCmd.Flags().BoolVar(&addEndpointDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addEndpointCmd.Flags().BoolVar(&addEndpointJSON, "json", false, "Emit machine-readable JSON output")
	addCmd.AddCommand(addEndpointCmd)
}

func runAddEndpoint(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addEndpointJSON)
		os.Exit(1)
	}
	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addEndpointJSON)
		os.Exit(1)
	}
	ctx.DryRun = addEndpointDryRun

	result, err := addgen.GenerateEndpoint(ctx, addgen.EndpointOpts{
		Name: args[0],
		Type: addEndpointType,
		Path: addEndpointPath,
	})
	if err != nil {
		printAddError(err, addEndpointJSON)
		os.Exit(1)
	}
	printAddResult(result, addEndpointDryRun, addEndpointJSON)
}

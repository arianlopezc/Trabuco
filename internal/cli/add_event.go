package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addEventFields string
	addEventDryRun bool
	addEventJSON   bool
)

var addEventCmd = &cobra.Command{
	Use:   "event <Name>",
	Short: "Generate a domain event record",
	Long: `Generate Model/.../events/{Name}.java — an immutable Java record
representing a domain event.

The CLI does NOT add the new event to any sealed permits clause or
listener switch — those edits stay with the coding agent.

Example:
  trabuco add event OrderShipped \
      --fields="orderId:string,shippedAt:instant,carrierRef:string?"`,
	Args: cobra.ExactArgs(1),
	Run:  runAddEvent,
}

func init() {
	addEventCmd.Flags().StringVar(&addEventFields, "fields", "", `Comma-separated field spec (required), e.g. "orderId:string,shippedAt:instant"`)
	addEventCmd.Flags().BoolVar(&addEventDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addEventCmd.Flags().BoolVar(&addEventJSON, "json", false, "Emit machine-readable JSON output")
	_ = addEventCmd.MarkFlagRequired("fields")
	addCmd.AddCommand(addEventCmd)
}

func runAddEvent(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addEventJSON)
		os.Exit(1)
	}
	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addEventJSON)
		os.Exit(1)
	}
	ctx.DryRun = addEventDryRun

	result, err := addgen.GenerateEvent(ctx, addgen.EventOpts{
		Name:   args[0],
		Fields: addEventFields,
	})
	if err != nil {
		printAddError(err, addEventJSON)
		os.Exit(1)
	}
	printAddResult(result, addEventDryRun, addEventJSON)
}

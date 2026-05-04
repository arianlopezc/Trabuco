package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addJobPayload string
	addJobDryRun  bool
	addJobJSON    bool
)

var addJobCmd = &cobra.Command{
	Use:   "job <Name>",
	Short: "Generate a JobRunr background job (request + handler base + concrete)",
	Long: `Generate the three-file JobRunr job bundle:

  - Model/.../jobs/{Name}JobRequest.java         (record implementing JobRequest)
  - Model/.../jobs/{Name}JobRequestHandler.java  (no-op base class)
  - Worker/.../handler/{Name}JobRequestHandler.java  (@Component subclass with TODO body)

Recurring schedule registration is left to the agent — edit
Worker/.../config/RecurringJobsConfig.java to add a cron entry if needed.

Example:
  trabuco add job ProcessShipment \
      --payload="orderId:string,carrierRef:string?,priority:integer"`,
	Args: cobra.ExactArgs(1),
	Run:  runAddJob,
}

func init() {
	addJobCmd.Flags().StringVar(&addJobPayload, "payload", "", `JobRequest payload fields (required), e.g. "orderId:string,amount:decimal"`)
	addJobCmd.Flags().BoolVar(&addJobDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addJobCmd.Flags().BoolVar(&addJobJSON, "json", false, "Emit machine-readable JSON output")
	_ = addJobCmd.MarkFlagRequired("payload")
	addCmd.AddCommand(addJobCmd)
}

func runAddJob(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addJobJSON)
		os.Exit(1)
	}
	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addJobJSON)
		os.Exit(1)
	}
	ctx.DryRun = addJobDryRun

	result, err := addgen.GenerateJob(ctx, addgen.JobOpts{
		Name:    args[0],
		Payload: addJobPayload,
	})
	if err != nil {
		printAddError(err, addJobJSON)
		os.Exit(1)
	}
	printAddResult(result, addJobDryRun, addJobJSON)
}

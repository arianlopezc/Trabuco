package cli

import (
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/spf13/cobra"
)

var (
	addStreamingPath   string
	addStreamingDryRun bool
	addStreamingJSON   bool
)

var addStreamingEndpointCmd = &cobra.Command{
	Use:   "streaming-endpoint <Name>",
	Short: "Generate an SSE streaming controller (AIAgent module)",
	Long: `Generate AIAgent/.../protocol/{Name}StreamController.java — a
Server-Sent Events controller skeleton suitable for LLM token
streaming or any long-lived request.

The skeleton uses a virtual thread (Project Loom) inside the emitter
loop; replace the placeholder send with real token streaming.

Example:
  trabuco add streaming-endpoint Conversation
  trabuco add streaming-endpoint Suggestion --path=/api/agent/suggest`,
	Args: cobra.ExactArgs(1),
	Run:  runAddStreamingEndpoint,
}

func init() {
	addStreamingEndpointCmd.Flags().StringVar(&addStreamingPath, "path", "", "Override the URL path (default: /api/agent/stream/{name})")
	addStreamingEndpointCmd.Flags().BoolVar(&addStreamingDryRun, "dry-run", false, "Print what would be created without writing to disk")
	addStreamingEndpointCmd.Flags().BoolVar(&addStreamingJSON, "json", false, "Emit machine-readable JSON output")
	addCmd.AddCommand(addStreamingEndpointCmd)
}

func runAddStreamingEndpoint(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		printAddError(err, addStreamingJSON)
		os.Exit(1)
	}
	ctx, err := addgen.LoadContext(cwd)
	if err != nil {
		printAddError(err, addStreamingJSON)
		os.Exit(1)
	}
	ctx.DryRun = addStreamingDryRun

	result, err := addgen.GenerateStreamingEndpoint(ctx, addgen.StreamingEndpointOpts{
		Name: args[0],
		Path: addStreamingPath,
	})
	if err != nil {
		printAddError(err, addStreamingJSON)
		os.Exit(1)
	}
	printAddResult(result, addStreamingDryRun, addStreamingJSON)
}

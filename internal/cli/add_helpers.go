package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/fatih/color"
)

// printAddResult renders an addgen.Result for the user. In JSON mode
// the structured form goes to stdout (so a calling agent or script
// can parse it). In human mode the colored summary mirrors the rest
// of the Trabuco CLI's output style.
func printAddResult(result *addgen.Result, dryRun bool, jsonMode bool) {
	if jsonMode {
		out := map[string]any{
			"status":  "success",
			"dry_run": dryRun,
			"created": result.Created,
		}
		if len(result.NextSteps) > 0 {
			out["next_steps"] = result.NextSteps
		}
		if len(result.Notes) > 0 {
			out["notes"] = result.Notes
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	if dryRun {
		yellow.Println("DRY RUN — no files were written.")
		fmt.Println()
	}
	green.Println("Created:")
	for _, p := range result.Created {
		fmt.Printf("  + %s\n", p)
	}
	if len(result.NextSteps) > 0 {
		fmt.Println()
		cyan.Println("Next steps:")
		for _, s := range result.NextSteps {
			fmt.Printf("  - %s\n", s)
		}
	}
	for _, n := range result.Notes {
		fmt.Printf("Note: %s\n", n)
	}
}

// printAddError formats an error from an add-command so JSON output
// stays parseable. CLI error text always goes to stderr; success
// output (above) goes to stdout. Callers handle os.Exit themselves.
func printAddError(err error, jsonMode bool) {
	if jsonMode {
		out := map[string]any{
			"status": "error",
			"error":  err.Error(),
		}
		enc := json.NewEncoder(os.Stderr)
		enc.SetIndent("", "  ")
		_ = enc.Encode(out)
		return
	}
	color.New(color.FgRed).Fprintf(os.Stderr, "Error: %v\n", err)
}

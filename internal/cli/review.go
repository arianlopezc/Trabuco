package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/generator"
)

// review.go — post-creation lifecycle for the code-review automation.
//
// Subcommands:
//   trabuco review status    — print current state
//   trabuco review enable    — flip config.enabled → true
//   trabuco review disable   — flip config.enabled → false (keeps artifacts)
//   trabuco review remove    — delete all review artifacts from the project
//   trabuco review install   — (re)scaffold review artifacts into an existing project
//
// All subcommands operate on the current working directory's Trabuco project
// (detected by .trabuco.json). They target Claude-specific artifacts since
// Codex/Cursor get the directive via their instruction files (which are part
// of the regular generation, not the review layer).

const reviewConfigRelPath = ".trabuco/review.config.json"

// reviewConfigFile is the on-disk shape of .trabuco/review.config.json. This
// mirrors what the template emits, plus the list of agents for human reading.
type reviewConfigFile struct {
	Enabled         bool     `json:"enabled"`
	Mode            string   `json:"mode"`
	GeneratedAt     string   `json:"generatedAt"`
	Agents          []string `json:"agents"`
	MaxReviewCycles int      `json:"maxReviewCycles"`
}

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Manage post-turn code review automation",
	Long: `Manage the code review automation in a generated Trabuco project.

The review system runs post-turn checks via subagents (code-reviewer,
performance-reviewer, prompt-reviewer) with an optional Stop-hook guard that
enforces the reviewer is invoked before a coding agent can finish a turn.

This command operates on the current working directory's Trabuco project.`,
}

var reviewStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current review state",
	RunE:  runReviewStatus,
}

var reviewEnableCmd = &cobra.Command{
	Use:   "enable",
	Short: "Enable the Stop-hook review guard",
	Long:  "Sets .trabuco/review.config.json enabled=true. Does not re-install missing artifacts — use 'trabuco review install' for that.",
	RunE:  runReviewEnable,
}

var reviewDisableCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable the Stop-hook review guard (keeps artifacts on disk)",
	Long:  "Sets .trabuco/review.config.json enabled=false. The hook early-returns; subagents and skills remain available for manual invocation.",
	RunE:  runReviewDisable,
}

var reviewRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Delete all review artifacts from the project",
	Long:  "Removes .claude/agents/{code,performance,prompt}-reviewer.md, .claude/hooks/require-review.sh, .claude/skills/review-{performance,prompts}/, .claude/HOOKS.md, and .trabuco/review.config.json. Does NOT remove .claude/hooks/format.sh (Spotless formatting) or the general /review skill.",
	RunE:  runReviewRemove,
}

var reviewInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "(Re)install review artifacts into an existing project",
	Long:  "Scaffolds review artifacts based on the current project's modules. Overwrites existing review files. Use after 'trabuco review remove' or in projects that were generated with --review=off.",
	RunE:  runReviewInstall,
}

var reviewInstallMode string

func init() {
	reviewInstallCmd.Flags().StringVar(&reviewInstallMode, "mode", config.ReviewModeFull, "Review mode to install: full (subagents + hooks + skills), minimal (no Stop hook guard)")

	reviewCmd.AddCommand(reviewStatusCmd)
	reviewCmd.AddCommand(reviewEnableCmd)
	reviewCmd.AddCommand(reviewDisableCmd)
	reviewCmd.AddCommand(reviewRemoveCmd)
	reviewCmd.AddCommand(reviewInstallCmd)
}

// --- shared helpers --------------------------------------------------------

// loadProjectConfig reads .trabuco.json from the current directory and returns
// the ProjectConfig needed to (re)generate review artifacts. This is how
// 'install' knows which module-gated subagents to emit without re-prompting.
func loadProjectConfig() (*config.ProjectConfig, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", fmt.Errorf("cannot determine working directory: %w", err)
	}
	metadataPath := filepath.Join(cwd, ".trabuco.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, "", fmt.Errorf("no .trabuco.json found in %s — run this command from the root of a Trabuco-generated project", cwd)
	}
	meta, err := config.LoadMetadata(cwd)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read .trabuco.json: %w", err)
	}
	return meta.ToProjectConfig(), cwd, nil
}

// readReviewConfig loads the on-disk review config. Missing file is treated as
// "no review installed" — callers decide whether that's an error.
func readReviewConfig(projectDir string) (*reviewConfigFile, error) {
	path := filepath.Join(projectDir, reviewConfigRelPath)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg reviewConfigFile
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("malformed %s: %w", reviewConfigRelPath, err)
	}
	return &cfg, nil
}

func writeReviewConfig(projectDir string, cfg *reviewConfigFile) error {
	path := filepath.Join(projectDir, reviewConfigRelPath)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	// Preserve trailing newline to match what the template emits.
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

// --- subcommand impls ------------------------------------------------------

func runReviewStatus(cmd *cobra.Command, args []string) error {
	_, projectDir, err := loadProjectConfig()
	if err != nil {
		return err
	}
	cfg, err := readReviewConfig(projectDir)
	if err != nil {
		return err
	}
	if cfg == nil {
		color.Yellow("Review automation is NOT installed in this project.")
		fmt.Println("Run `trabuco review install` to scaffold it.")
		return nil
	}
	state := color.GreenString("enabled")
	if !cfg.Enabled {
		state = color.YellowString("disabled")
	}
	fmt.Printf("Review automation: %s\n", state)
	fmt.Printf("  Mode:             %s\n", cfg.Mode)
	fmt.Printf("  Generated at:     %s\n", cfg.GeneratedAt)
	fmt.Printf("  Max review cycles: %d\n", cfg.MaxReviewCycles)
	fmt.Printf("  Agents:           %v\n", cfg.Agents)

	hookPath := filepath.Join(projectDir, ".claude/hooks/require-review.sh")
	if _, err := os.Stat(hookPath); err == nil {
		fmt.Println("  Stop hook:        present")
	} else {
		fmt.Println("  Stop hook:        not present (minimal mode or uninstalled)")
	}
	return nil
}

func runReviewEnable(cmd *cobra.Command, args []string) error {
	_, projectDir, err := loadProjectConfig()
	if err != nil {
		return err
	}
	cfg, err := readReviewConfig(projectDir)
	if err != nil {
		return err
	}
	if cfg == nil {
		return fmt.Errorf("review is not installed — run `trabuco review install` first")
	}
	cfg.Enabled = true
	if err := writeReviewConfig(projectDir, cfg); err != nil {
		return err
	}
	color.Green("✓ Review automation enabled.")
	return nil
}

func runReviewDisable(cmd *cobra.Command, args []string) error {
	_, projectDir, err := loadProjectConfig()
	if err != nil {
		return err
	}
	cfg, err := readReviewConfig(projectDir)
	if err != nil {
		return err
	}
	if cfg == nil {
		color.Yellow("Review automation is not installed — nothing to disable.")
		return nil
	}
	cfg.Enabled = false
	if err := writeReviewConfig(projectDir, cfg); err != nil {
		return err
	}
	color.Green("✓ Review automation disabled. Subagents and skills remain available for manual use; the Stop hook will no longer block turns.")
	return nil
}

func runReviewRemove(cmd *cobra.Command, args []string) error {
	_, projectDir, err := loadProjectConfig()
	if err != nil {
		return err
	}
	// Artifacts to remove. Kept intentionally narrow — we do NOT touch
	// format.sh (pure Spotless, useful standalone) or .claude/skills/review
	// (the general /review skill, useful without the automation layer).
	toRemoveFiles := []string{
		".claude/agents/code-reviewer.md",
		".claude/agents/performance-reviewer.md",
		".claude/agents/prompt-reviewer.md",
		".claude/hooks/require-review.sh",
		".claude/HOOKS.md",
		".codex/hooks/require-review.sh",
		".cursor/hooks/require-review.sh",
		".trabuco/review.config.json",
	}
	toRemoveDirs := []string{
		".claude/skills/review-performance",
		".claude/skills/review-prompts",
		".trabuco/review-state",
	}
	removed := 0
	for _, rel := range toRemoveFiles {
		path := filepath.Join(projectDir, rel)
		if err := os.Remove(path); err == nil {
			removed++
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", rel, err)
		}
	}
	for _, rel := range toRemoveDirs {
		path := filepath.Join(projectDir, rel)
		if err := os.RemoveAll(path); err == nil {
			info, _ := os.Stat(path)
			if info == nil {
				removed++
			}
		} else {
			return fmt.Errorf("failed to remove %s: %w", rel, err)
		}
	}
	color.Green("✓ Removed %d review artifacts.", removed)
	fmt.Println("Note: .claude/hooks/format.sh and the general /review skill were kept — they are independent of the review automation.")
	fmt.Println("Note: You may need to manually remove the `Stop` hook entry from .claude/settings.json.")
	return nil
}

func runReviewInstall(cmd *cobra.Command, args []string) error {
	validModes := map[string]bool{
		config.ReviewModeFull:    true,
		config.ReviewModeMinimal: true,
	}
	if !validModes[reviewInstallMode] {
		return fmt.Errorf("invalid --mode '%s'. Valid: full, minimal", reviewInstallMode)
	}

	cfg, projectDir, err := loadProjectConfig()
	if err != nil {
		return err
	}
	// Review artifacts primarily install for Claude (subagents + skills + hooks)
	// but cross-tool Stop adapters (Codex, Cursor) are also emitted when those
	// agents are selected. Only refuse when no supported tool is present.
	if !cfg.HasAIAgent("claude") && !cfg.HasAIAgent("codex") && !cfg.HasAIAgent("cursor") {
		return fmt.Errorf("no compatible AI agent found in this project. Review artifacts install for Claude, Codex, or Cursor — at least one must be present in .trabuco.json")
	}

	cfg.Review = config.ReviewConfig{
		Mode:        reviewInstallMode,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Clean up stale per-mode artifacts before regenerating. Specifically, a
	// minimal install must not leave a require-review.sh from a prior full
	// install on disk — users expect `install --mode=minimal` to match a fresh
	// minimal scaffold, not a superset of whatever was there.
	if reviewInstallMode == config.ReviewModeMinimal {
		stale := filepath.Join(projectDir, ".claude/hooks/require-review.sh")
		if err := os.Remove(stale); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove stale Stop hook: %w", err)
		}
	}

	gen, err := generator.NewWithVersionAt(cfg, "", projectDir)
	if err != nil {
		return err
	}
	if err := gen.GenerateReviewArtifactsOnly(); err != nil {
		return err
	}
	color.Green("✓ Installed review artifacts (mode=%s).", reviewInstallMode)
	if reviewInstallMode == config.ReviewModeFull {
		fmt.Println("To activate the Stop hook, ensure .claude/settings.json includes:")
		fmt.Println(`  "Stop": [{"hooks":[{"type":"command","command":"\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/require-review.sh"}]}]`)
	} else {
		fmt.Println("Note: minimal mode does not emit the Stop hook guard. The `code-reviewer` subagent's description still drives auto-delegation, and CLAUDE.md still carries the directive.")
	}
	return nil
}

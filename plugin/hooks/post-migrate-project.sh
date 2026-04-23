#!/usr/bin/env bash
# PostToolUse hook for mcp__trabuco__migrate_project.
#
# Fires after a legacy-project migration completes. Reinforces the "source
# is untouched, verify output before trusting" posture that /trabuco:migrate
# already establishes — this is belt-and-braces for conversations where
# migrate_project was called outside the skill flow.

set -u
cat >/dev/null 2>&1 || true

cat <<'EOF'
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "A Trabuco migration just ran via migrate_project. Next steps for the user:\n\n1. The SOURCE project was not modified. Migration output lives in a NEW directory (tool output lists it).\n2. Immediately run run_doctor on the output directory (or invoke /trabuco:doctor). Critical findings must be resolved before trusting the migration.\n3. `cd` into the output directory and run `mvn clean install` — LLM-generated code sometimes compiles locally but breaks under strict Maven settings.\n4. Review file-by-file: the AI placed controllers in API, entities in Model, services in Shared, etc. Spot-check a few critical files for misplacement.\n5. Commit migration output to a NEW branch (e.g. `trabuco-migration`) — never directly to main. Treat it like a large-scale refactor PR.\n6. If the source used libraries Trabuco doesn't support (identified in the prior scan), those code paths were likely replaced or stubbed. Search the output for TODO markers and migration-notes comments before relying on the code.\n7. If dry_run was never executed, warn the user: they should have run with dry_run:true first; LLM credits are non-refundable."
  }
}
EOF
exit 0

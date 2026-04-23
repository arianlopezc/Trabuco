#!/usr/bin/env bash
# PostToolUse hook for mcp__trabuco__init_project.
#
# When the assistant generates a new Trabuco project, this hook injects
# next-steps context so the assistant can guide the user through verification
# and first-code-change steps without having to re-derive them.
#
# The hook reads the tool result from stdin (standard PostToolUse contract)
# and emits `hookSpecificOutput.additionalContext`. It does not block the
# tool call — tool already succeeded by the time this fires.

set -u

# Drain stdin — Claude Code passes a JSON payload on stdin for PostToolUse
# hooks. We don't need to parse the fields; the tool has already run. We
# inject general next-steps guidance that applies to every init_project.
cat >/dev/null 2>&1 || true

cat <<'EOF'
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "A Trabuco project was just generated via init_project. Guide the user through these next steps:\n\n1. `cd` into the generated directory (the tool output lists it).\n2. Run `mvn clean install -DskipTests` to confirm the multi-module build succeeds.\n3. If AIAgent module is present, set ANTHROPIC_API_KEY (or the configured provider's key) in the environment before running.\n4. Run `docker compose up -d` if the project pulled in Postgres/Kafka/Mongo — the local compose file spins dependencies up.\n5. Inside the generated project, surface the project-local skills: /add-entity, /add-endpoint, /add-migration, /add-job (if Worker), /add-event-handler (if EventConsumer), /add-tool + /add-guardrail-rule + /add-a2a-skill + /add-knowledge-entry (if AIAgent). These are DIFFERENT from the plugin's /trabuco:* skills — they know the project's conventions.\n6. Remind the user: placeholder entities/migrations exist to show structure; they MUST be replaced before production use.\n7. Before any commit in the generated project, `./review-checks.sh` enforces Trabuco's conventions (keyset pagination, no FK constraints, constructor injection, no DB mocks)."
  }
}
EOF
exit 0

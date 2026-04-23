#!/usr/bin/env bash
# PostToolUse hook for mcp__trabuco__generate_workspace.
#
# Fires after a multi-service workspace is generated. Injects next-steps
# context scoped to multi-service concerns (build order, cross-service
# contracts, local orchestration).

set -u
cat >/dev/null 2>&1 || true

cat <<'EOF'
{
  "hookSpecificOutput": {
    "hookEventName": "PostToolUse",
    "additionalContext": "A Trabuco workspace (multiple services) was just generated via generate_workspace. Guide the user through these next steps:\n\n1. `cd` into the workspace root (tool output lists it). Each service is a sibling Maven project.\n2. If a top-level docker-compose.yml was emitted, `docker compose up -d` starts shared infra (DB, broker) for all services.\n3. Build services in dependency order — services that publish events must be built before services that consume them, so their Events module is available. `mvn -T 1C clean install -DskipTests` from the workspace root handles this if a parent POM exists; otherwise build each service individually.\n4. Cross-service contracts live in each service's Events module. When adding a new event in service A, service B must update its dependency or copy the contract; there is no shared contracts module by default.\n5. Remind the user: each service has its own CLAUDE.md and review-checks.sh. Changes in one service do NOT trigger other services' review tooling — that's their responsibility.\n6. For local dev across services, recommend a single terminal per service (or a tool like tmuxinator / overmind) rather than running all services in one process."
  }
}
EOF
exit 0

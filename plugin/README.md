# Trabuco — Claude Code plugin

Production-ready Java/Spring Boot scaffolding and orchestrated migration of
legacy projects, exposed as Claude Code skills + subagents and an MCP server.

This is the plugin half of [Trabuco](https://github.com/arianlopezc/Trabuco).
The CLI half (the `trabuco` binary) provides scaffolding, the architecture
advisor, and the 14-phase migration; this plugin gives Claude Code first-
class access to all of it.

## What you get

- **8 skills** — `/trabuco:new-project`, `/trabuco:design-system`,
  `/trabuco:add-module`, `/trabuco:extend`, `/trabuco:migrate`,
  `/trabuco:doctor`, `/trabuco:suggest`, `/trabuco:sync`.
- **17 subagents** — the architect (`trabuco-architect`), AI agent expert
  (`trabuco-ai-agent-expert`), the migration orchestrator
  (`trabuco-migration-orchestrator`), and 14 phase specialists for the
  migration (assessor, skeleton-builder, model, datastore, shared, api,
  worker, eventconsumer, aiagent, config, deployment, tests, activator,
  finalizer).
- **MCP server** — 25 tools exposed by the `trabuco` CLI: scaffolding
  (`init_project`, `add_module`, `suggest_architecture`, `design_system`,
  `generate_workspace`, `run_doctor`, `get_project_info`, `list_modules`,
  `list_providers`, `check_docker`, `get_version`, `auth_status`,
  `sync_project`) and the 14-phase migration
  (`migrate_assess`, `migrate_skeleton`, `migrate_module`, `migrate_config`,
  `migrate_deployment`, `migrate_tests`, `migrate_activate`,
  `migrate_finalize`, `migrate_status`, `migrate_rollback`,
  `migrate_decision`, `migrate_resume`).
- **Hooks** — a `SessionStart` hook that verifies the `trabuco` CLI is
  installed and on PATH; `PostToolUse` hooks that follow up after
  `init_project` and `generate_workspace` to set the user up correctly.

## Prerequisites

The plugin is a thin layer over the `trabuco` CLI. Without the CLI, the
MCP server can't start and the `/trabuco:*` skills will fail.

Install the CLI before installing the plugin:

```bash
curl -sSL https://github.com/arianlopezc/Trabuco/releases/latest/download/install.sh | bash
```

Other install paths: download a prebuilt binary for your platform from
[the releases page](https://github.com/arianlopezc/Trabuco/releases/latest),
or `go install github.com/arianlopezc/Trabuco/cmd/trabuco@latest` to build
from source. There is no Homebrew formula.

The `SessionStart` hook checks the CLI is at version 1.8.0 or newer and
emits a clear message if it isn't.

## Install the plugin

### From Anthropic's community marketplace (recommended)

Trabuco is listed in
[`anthropics/claude-plugins-community`](https://github.com/anthropics/claude-plugins-community),
Anthropic's community plugin catalog. Inside Claude Code:

```text
/plugin marketplace add anthropics/claude-plugins-community
/plugin install trabuco@claude-community
```

Browse the catalog at [claude.com/plugins](https://claude.com/plugins/).

### From the Trabuco repository

```text
/plugin marketplace add arianlopezc/Trabuco
/plugin install trabuco@trabuco-marketplace
```

This clones the Trabuco repo and installs the plugin from
[`./plugin`](./) inside it. Useful if you want unreleased changes from
`main` ahead of the next community-marketplace sync.

### From a downloaded release tarball

```bash
curl -L https://github.com/arianlopezc/Trabuco/releases/latest/download/trabuco-plugin-v1.11.0.tar.gz \
  | tar -xz -C ~
```

Then in Claude Code:

```text
/plugin marketplace add ~/trabuco-plugin-v1.11.0
/plugin install trabuco@trabuco-marketplace
```

## Quickstart

Once installed, ask Claude Code things like:

- *"Start a new payments service with Postgres and a worker."* → `/trabuco:new-project` walks the assistant through requirements, recommends the right modules, and runs `trabuco init` for you.
- *"Design a system: user service, billing service, notifications worker."* → `/trabuco:design-system` decomposes it into independent Trabuco services.
- *"Migrate this Spring Boot 2.7 monolith to Trabuco."* → `/trabuco:migrate` drives the 14-phase orchestrated flow with per-phase approval gates.
- *"What should I add next to this Trabuco project?"* → `/trabuco:extend` reviews the project and proposes the next module.

## Compatibility

| | Required |
| - | - |
| Claude Code | latest (anything that supports the `/plugin` command) |
| `trabuco` CLI | ≥ 1.8.0 (latest is 1.11.0) |
| Java (for projects) | 21+ |
| Maven (for migration) | 3.9+ |

The CLI runs on macOS (arm64/amd64), Linux (arm64/amd64), and Windows
(amd64). The plugin itself is platform-agnostic.

## Documentation

- [Trabuco overview](https://github.com/arianlopezc/Trabuco) — README, full
  feature list, generated project anatomy.
- [Migration guide](https://github.com/arianlopezc/Trabuco/blob/main/docs/migration-guide.md)
  — how the 14-phase migration works, gates, decisions, rollback,
  troubleshooting.
- [Auth scaffolding guide](https://github.com/arianlopezc/Trabuco/blob/main/docs/auth.md)
  — OIDC Resource Server (auto-generated when API or AIAgent is selected,
  dormant until `trabuco.auth.enabled=true`), per-provider recipes for
  Keycloak / Auth0 / Okta / Cognito / generic OIDC, dual SecurityFilterChain
  pattern, coexistence with the legacy `ApiKeyAuthFilter`.
- [Plugin docs](./docs/) — when to use Trabuco, when not to, and the
  catalog of patterns it generates.

## License

MIT. See [LICENSE](../LICENSE) at the repository root.

## Author / Maintainer

Arian Lopez — [github.com/arianlopezc](https://github.com/arianlopezc)

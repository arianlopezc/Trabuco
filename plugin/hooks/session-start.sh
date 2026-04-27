#!/usr/bin/env bash
# SessionStart hook: verify the trabuco CLI is installed and compatible.
#
# Emits a JSON payload with `hookSpecificOutput.additionalContext` so the
# session's system prompt gets augmented with the binary status. This means
# the assistant knows, up front, whether it can call Trabuco MCP tools or
# whether it needs to instruct the user to install the CLI first.
#
# The plugin's MCP server is spawned via `trabuco mcp`. If the binary is
# missing, MCP tool calls will fail; telling the assistant early avoids
# confusing error traces mid-conversation.

set -u

# Minimum trabuco version this plugin expects. Keep in lockstep with
# plugin.json version when the MCP surface changes.
MIN_VERSION="1.8.0"

emit_context() {
  local msg="$1"
  # Escape for JSON: replace backslashes and double quotes.
  local escaped="${msg//\\/\\\\}"
  escaped="${escaped//\"/\\\"}"
  # Collapse newlines to \n for JSON safety.
  escaped="${escaped//$'\n'/\\n}"
  cat <<EOF
{
  "hookSpecificOutput": {
    "hookEventName": "SessionStart",
    "additionalContext": "${escaped}"
  }
}
EOF
}

if ! command -v trabuco >/dev/null 2>&1; then
  emit_context "Trabuco plugin: the \`trabuco\` CLI is NOT installed on this machine. The plugin's MCP server cannot start and all /trabuco:* skills will fail. Before using any Trabuco feature, instruct the user to install it — see https://github.com/arianlopezc/Trabuco for install instructions (typically: brew install or a prebuilt binary). Do NOT attempt to call mcp__trabuco__* tools until installed."
  exit 0
fi

# Capture version; tolerate any output format.
raw_version="$(trabuco --version 2>/dev/null || trabuco version 2>/dev/null || echo "unknown")"
# Extract the first semver-like token (e.g. "1.8.3", "v1.8.3").
version="$(printf '%s' "$raw_version" | grep -Eo 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -n 1 | sed 's/^v//')"

if [ -z "$version" ]; then
  emit_context "Trabuco plugin: \`trabuco\` is installed but reported an unrecognizable version string ('${raw_version}'). MCP tools may still work — proceed cautiously. Recommend running \`trabuco --version\` manually to confirm."
  exit 0
fi

# Semver comparison via sort -V. If current >= min, ok.
lowest="$(printf '%s\n%s\n' "$MIN_VERSION" "$version" | sort -V | head -n 1)"
if [ "$lowest" != "$MIN_VERSION" ]; then
  emit_context "Trabuco plugin: installed \`trabuco\` is version ${version}, but this plugin expects >= ${MIN_VERSION}. Some MCP tools or prompts may be missing or have different signatures. Recommend the user upgrade before proceeding — see https://github.com/arianlopezc/Trabuco/releases."
  exit 0
fi

emit_context "Trabuco plugin ready: \`trabuco\` ${version} detected (>= ${MIN_VERSION}). MCP tools under mcp__trabuco__* are available. Skills available: /trabuco:new-project, /trabuco:design-system, /trabuco:add-module, /trabuco:extend, /trabuco:doctor, /trabuco:suggest. Specialist subagents: trabuco-architect, trabuco-ai-agent-expert."
exit 0

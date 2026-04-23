#!/usr/bin/env bash
#
# test_full_build.sh — generate representative archetypes and run `mvn clean
# install` with the full test suite on each. This is the e2e gate that should
# have caught v1.8.1/v1.8.2: it exercises the CLI end-to-end (generation +
# build + tests) against real Maven + the user's JDK.
#
# Archetypes are chosen to cover orthogonal dimensions: datastore presence,
# AIAgent presence, and broker family (the three axes that historically produce
# incompatible dependency combinations). Four combinations catch far more than
# one "all modules" mega-test because isolated failures are easier to bisect.
#
# Usage:
#   scripts/e2e/test_full_build.sh                # run all archetypes
#   scripts/e2e/test_full_build.sh api-minimal    # single archetype
#   TRABUCO_BIN=/path/to/trabuco scripts/e2e/test_full_build.sh
#
# Exits non-zero if any archetype fails to generate or `mvn clean install`
# (tests included) fails on it.

set -u
set -o pipefail

# ─── Paths & binary ──────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TRABUCO_BIN="${TRABUCO_BIN:-${REPO_ROOT}/trabuco}"

if [[ ! -x "${TRABUCO_BIN}" ]]; then
  echo "error: trabuco binary not found or not executable at: ${TRABUCO_BIN}" >&2
  echo "hint:  run 'go build -o trabuco ./cmd/trabuco' from the repo root, or set TRABUCO_BIN" >&2
  exit 2
fi

WORKDIR="$(mktemp -d -t trabuco-e2e-XXXXXX)"
trap 'rm -rf "${WORKDIR}"' EXIT

# ─── Archetype definitions ───────────────────────────────────────────────────
# Each archetype is: <name>|<modules>|<extra-flags>
# The orthogonal axes we're covering:
#   - api-minimal:   no datastore, no AI, no broker — smallest surface, baseline
#   - crud-sql:      SQL datastore + worker — exercises perf/N+1 rule surface
#   - aiagent-pubsub: AIAgent + GCP Pub/Sub — the exact shape that broke for the
#                    user (Mockito/Jacoco/javax.annotation transitive bans)
#   - full-fat:      every module + RabbitMQ — max dependency surface
ARCHETYPES=(
  "api-minimal|Model,API|"
  "crud-sql|Model,SQLDatastore,Shared,API,Worker,Jobs|--database=postgresql"
  "aiagent-pubsub|Model,SQLDatastore,Shared,Events,EventConsumer,AIAgent|--database=postgresql --message-broker=pubsub --ai-agents=claude"
  "full-fat|Model,SQLDatastore,Shared,Events,API,Worker,EventConsumer,AIAgent,Jobs|--database=postgresql --message-broker=rabbitmq --ai-agents=claude"
)

FILTER="${1:-}"
FAILED=()
PASSED=()

# ─── Helpers ─────────────────────────────────────────────────────────────────
log()  { printf '\033[1;36m[e2e]\033[0m %s\n' "$*"; }
ok()   { printf '\033[1;32m[ ok ]\033[0m %s\n' "$*"; }
fail() { printf '\033[1;31m[fail]\033[0m %s\n' "$*"; }

run_archetype() {
  local name="$1" modules="$2" extra="$3"
  local proj_dir="${WORKDIR}/${name}"
  local log_file="${WORKDIR}/${name}.log"

  log "Generating ${name} (modules=${modules})"
  mkdir -p "${proj_dir}"

  # Intentionally --skip-build here: we want to run the build ourselves with
  # streaming output so CI logs show progress and errors surface fast.
  if ! ( cd "${proj_dir}" && "${TRABUCO_BIN}" init \
      --name="${name//-/}" \
      --group-id="com.trabuco.e2e" \
      --modules="${modules}" \
      --skip-build \
      ${extra} \
    > "${log_file}" 2>&1 ); then
    fail "${name}: generation failed"
    tail -40 "${log_file}" >&2
    FAILED+=("${name}")
    return 1
  fi

  local generated="${proj_dir}/${name//-/}"
  if [[ ! -d "${generated}" ]]; then
    fail "${name}: expected generated directory ${generated} missing"
    FAILED+=("${name}")
    return 1
  fi

  log "Building ${name} with mvn clean install (tests included)"
  if ! ( cd "${generated}" && mvn -B clean install ) >> "${log_file}" 2>&1; then
    fail "${name}: mvn clean install failed"
    echo "── last 60 lines of ${log_file} ──" >&2
    tail -60 "${log_file}" >&2
    FAILED+=("${name}")
    return 1
  fi

  ok "${name}: generated + built + tested"
  PASSED+=("${name}")
}

# ─── Main loop ───────────────────────────────────────────────────────────────
log "Repo: ${REPO_ROOT}"
log "Workdir: ${WORKDIR}"
log "Binary: ${TRABUCO_BIN}"
log "Java: $(java -version 2>&1 | head -1)"
log "Maven: $(mvn -v 2>&1 | head -1)"

for entry in "${ARCHETYPES[@]}"; do
  IFS='|' read -r name modules extra <<< "${entry}"
  if [[ -n "${FILTER}" && "${FILTER}" != "${name}" ]]; then
    continue
  fi
  run_archetype "${name}" "${modules}" "${extra}" || true
done

# ─── Summary ─────────────────────────────────────────────────────────────────
echo
log "Summary"
for name in "${PASSED[@]}"; do ok "  ${name}"; done
for name in "${FAILED[@]}"; do fail "  ${name}"; done

if [[ -n "${FILTER}" && ${#PASSED[@]} -eq 0 && ${#FAILED[@]} -eq 0 ]]; then
  echo "error: filter '${FILTER}' matched no archetype" >&2
  exit 2
fi

if [[ ${#FAILED[@]} -gt 0 ]]; then
  exit 1
fi

#!/usr/bin/env bash
#
# test_combinations.sh — generate a broad matrix of module / database /
# broker / vector-store combinations with `trabuco init` and run
# `mvn clean install` (tests included) on every one of them.
#
# test_full_build.sh is the fast CI gate (7 archetypes chosen to cover the
# orthogonal axes). This script is the exhaustive local sweep used before a
# dependency-line upgrade (Spring Boot / Spring AI / broker SDK bumps): it
# walks every selectable option — each SQL database, each NoSQL database,
# each message broker, each vector store, Worker with and without a SQL
# datastore, and every AI-agent flavour — so an incompatible dependency
# combination cannot hide behind the CI archetypes.
#
# Usage:
#   scripts/e2e/test_combinations.sh                 # run everything, 2 at a time
#   scripts/e2e/test_combinations.sh -j 3            # 3 builds in parallel
#   scripts/e2e/test_combinations.sh -j 1 sql-mysql  # single combination
#   TRABUCO_BIN=/path/to/trabuco scripts/e2e/test_combinations.sh
#   TRABUCO_JAVA_VERSION=21 scripts/e2e/test_combinations.sh
#
# Requirements: the trabuco binary (go build -o trabuco ./cmd/trabuco), a JDK,
# Maven and a running Docker daemon (Testcontainers-backed tests).
#
# Exits non-zero if any combination fails to generate or build.

set -u
set -o pipefail

if (( BASH_VERSINFO[0] < 4 || (BASH_VERSINFO[0] == 4 && BASH_VERSINFO[1] < 3) )); then
  echo "error: this script needs bash >= 4.3 (wait -n); on macOS install it with 'brew install bash'" >&2
  exit 2
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TRABUCO_BIN="${TRABUCO_BIN:-${REPO_ROOT}/trabuco}"
PARALLEL=2

while getopts ":j:" opt; do
  case "${opt}" in
    j) PARALLEL="${OPTARG}" ;;
    *) echo "usage: $0 [-j N] [combination-name]" >&2; exit 2 ;;
  esac
done
shift $((OPTIND - 1))
FILTER="${1:-}"

if [[ ! -x "${TRABUCO_BIN}" ]]; then
  echo "error: trabuco binary not found or not executable at: ${TRABUCO_BIN}" >&2
  echo "hint:  run 'go build -o trabuco ./cmd/trabuco' from the repo root, or set TRABUCO_BIN" >&2
  exit 2
fi

WORKDIR="${TRABUCO_E2E_WORKDIR:-$(mktemp -d -t trabuco-combos-XXXXXX)}"
mkdir -p "${WORKDIR}"
if [[ -z "${TRABUCO_E2E_WORKDIR:-}" ]]; then
  trap 'rm -rf "${WORKDIR}"' EXIT
fi

if [[ -z "${TRABUCO_JAVA_VERSION:-}" ]]; then
  TRABUCO_JAVA_VERSION="$(java -version 2>&1 | head -1 | sed -E 's/.*version "([0-9]+).*/\1/')"
fi
if ! [[ "${TRABUCO_JAVA_VERSION}" =~ ^[0-9]+$ ]]; then
  echo "error: could not determine Java major version; set TRABUCO_JAVA_VERSION" >&2
  exit 2
fi

# ─── Combination definitions ─────────────────────────────────────────────────
# <name>|<modules>|<extra flags>
#
# Axes covered (each value appears in at least one combination):
#   SQL database:   postgresql, mysql, none (H2)
#   NoSQL database: mongodb, redis
#   Broker:         kafka, rabbitmq, sqs, pubsub
#   Vector store:   (none), pgvector, qdrant, mongodb
#   Worker:         with SQLDatastore, with NoSQL (JobRunr PostgreSQL
#                   fallback), with no datastore at all
#   AI agents:      claude, cursor, copilot, codex (+ --ci=github)
COMBINATIONS=(
  "model-only|Model|"
  "api-minimal|Model,API|"
  "sql-mysql|Model,SQLDatastore,Shared,API|--database=mysql"
  "sql-h2|Model,SQLDatastore,Shared,API|--database=none"
  "nosql-mongo|Model,NoSQLDatastore,Shared,API|--nosql-database=mongodb"
  "nosql-redis-worker|Model,NoSQLDatastore,Shared,API,Worker,Jobs|--nosql-database=redis"
  "worker-no-db|Model,Worker,Jobs|"
  "crud-sql|Model,SQLDatastore,Shared,API,Worker,Jobs|--database=postgresql"
  "events-kafka|Model,SQLDatastore,Shared,API,Events,EventConsumer|--database=postgresql --message-broker=kafka"
  "events-rabbitmq|Model,Shared,API,Events,EventConsumer|--message-broker=rabbitmq"
  "events-sqs|Model,Shared,Events,EventConsumer|--message-broker=sqs"
  "events-pubsub|Model,SQLDatastore,Shared,Events,EventConsumer|--database=postgresql --message-broker=pubsub"
  "aiagent-bare|Model,AIAgent|--ai-agents=claude"
  "aiagent-rag|Model,SQLDatastore,Shared,API,AIAgent|--database=postgresql --vector-store=pgvector --ai-agents=claude"
  "aiagent-qdrant|Model,Shared,API,AIAgent|--vector-store=qdrant --ai-agents=claude"
  "aiagent-mongo|Model,NoSQLDatastore,Shared,API,AIAgent|--nosql-database=mongodb --vector-store=mongodb --ai-agents=claude"
  "aiagent-pubsub|Model,SQLDatastore,Shared,Events,EventConsumer,AIAgent|--database=postgresql --message-broker=pubsub --ai-agents=claude"
  "everything-mysql-kafka|Model,SQLDatastore,Shared,Events,API,Worker,EventConsumer,AIAgent,Jobs|--database=mysql --message-broker=kafka --ai-agents=claude,cursor,copilot,codex --ci=github"
  "full-fat|Model,SQLDatastore,Shared,Events,API,Worker,EventConsumer,AIAgent,Jobs|--database=postgresql --message-broker=rabbitmq --ai-agents=claude"
)

log()  { printf '\033[1;36m[combos]\033[0m %s\n' "$*"; }
ok()   { printf '\033[1;32m[ ok ]\033[0m %s\n' "$*"; }
fail() { printf '\033[1;31m[fail]\033[0m %s\n' "$*"; }

run_combination() {
  local name="$1" modules="$2" extra="$3"
  local proj_dir="${WORKDIR}/${name}"
  local log_file="${WORKDIR}/${name}.log"
  local status_file="${WORKDIR}/${name}.status"
  local project="${name//-/}"

  rm -rf "${proj_dir}"; mkdir -p "${proj_dir}"
  if ! ( cd "${proj_dir}" && "${TRABUCO_BIN}" init \
      --name="${project}" \
      --group-id="com.trabuco.combos" \
      --modules="${modules}" \
      --java-version="${TRABUCO_JAVA_VERSION}" \
      --skip-build \
      ${extra} \
    > "${log_file}" 2>&1 ); then
    echo "generation-failed" > "${status_file}"; return 1
  fi
  local generated="${proj_dir}/${project}"
  if [[ ! -d "${generated}" ]]; then
    echo "generation-failed" > "${status_file}"; return 1
  fi
  # Each combination gets a private local repository: the generated
  # modules share artifact coordinates (com.trabuco.combos:Model, ...) and
  # parallel `mvn install` runs would otherwise overwrite each other's
  # artifacts in ~/.m2. The shared ~/.m2 is still used as a read-through
  # cache for third-party dependencies via the Maven settings mirror below.
  if ! ( cd "${generated}" && mvn -B clean install \
        -Dmaven.repo.local="${WORKDIR}/.m2-${name}" \
        -Dmaven.repo.local.tail="${HOME}/.m2/repository" \
        -Daether.enhancedLocalRepository.split=false ) >> "${log_file}" 2>&1; then
    echo "build-failed" > "${status_file}"; return 1
  fi
  echo "ok" > "${status_file}"
}

log "Repo: ${REPO_ROOT}"
log "Workdir: ${WORKDIR}"
log "Java: $(java -version 2>&1 | head -1) (generating with --java-version=${TRABUCO_JAVA_VERSION})"
log "Maven: $(mvn -v 2>&1 | head -1)"
log "Parallelism: ${PARALLEL}"

SELECTED=()
for entry in "${COMBINATIONS[@]}"; do
  IFS='|' read -r name _ _ <<< "${entry}"
  if [[ -n "${FILTER}" && "${FILTER}" != "${name}" ]]; then continue; fi
  SELECTED+=("${entry}")
done
if [[ ${#SELECTED[@]} -eq 0 ]]; then
  echo "error: no combination named '${FILTER}'" >&2; exit 2
fi

running=0
for entry in "${SELECTED[@]}"; do
  IFS='|' read -r name modules extra <<< "${entry}"
  log "Starting ${name} (modules=${modules} ${extra})"
  run_combination "${name}" "${modules}" "${extra}" &
  running=$((running + 1))
  if [[ ${running} -ge ${PARALLEL} ]]; then
    wait -n || true
    running=$((running - 1))
  fi
done
wait

log "Summary"
FAILED=0
for entry in "${SELECTED[@]}"; do
  IFS='|' read -r name _ _ <<< "${entry}"
  status="$(cat "${WORKDIR}/${name}.status" 2>/dev/null || echo "unknown")"
  if [[ "${status}" == "ok" ]]; then
    ok "  ${name}"
  else
    fail "  ${name}: ${status} (see ${WORKDIR}/${name}.log)"
    FAILED=$((FAILED + 1))
  fi
done

if [[ ${FAILED} -gt 0 ]]; then
  echo
  fail "${FAILED} combination(s) failed"
  exit 1
fi
ok "All ${#SELECTED[@]} combinations generated, built and tested"

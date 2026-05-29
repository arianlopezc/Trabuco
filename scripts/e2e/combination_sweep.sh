#!/usr/bin/env bash
#
# combination_sweep.sh — exhaustive generation + CLI-check sweep.
#
# Enumerates EVERY module-presence combination the CLI can generate
# (all 16 subsets of the four edge modules {API,Worker,EventConsumer,AIAgent}
# crossed with all 5 datastore choices = 80) and runs the REAL `trabuco init`
# on each — i.e. the CLI's own post-generation check (spotless:apply +
# mvn clean install -DskipTests). A combo "fails" if generation fails or the
# CLI reports the Maven build failed.
#
# Scope is generation + the CLI build check only — no Docker, no app boot.
#
# Usage: TRABUCO_BIN=./trabuco JOBS=8 scripts/e2e/combination_sweep.sh
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TRABUCO_BIN="${TRABUCO_BIN:-${REPO_ROOT}/trabuco}"
WORK="${WORK:-/tmp/trabuco-matrix}"
RESULTS="${WORK}/results.tsv"

first_error() {
  grep -m1 -E "\.java:|BUILD FAILURE|Could not resolve|cannot find symbol|package .* does not exist|Non-resolvable|Banned|Dependency convergence|Failed to execute goal|Error: " "$1" 2>/dev/null | sed 's/^[[:space:]]*//' | cut -c1-220
}

# ── worker mode: process exactly one "name|modules|flags" line ───────────────
if [[ "${1:-}" == "__worker" ]]; then
  line="$2"
  name="${line%%|*}"; rest="${line#*|}"; mods="${rest%%|*}"; flags="${rest#*|}"
  dir="${WORK}/${name}"; log="${WORK}/${name}.log"
  # Derive a CLI-valid project name (lowercase alphanumeric only).
  pname="$(printf '%s' "${name}" | tr 'A-Z' 'a-z' | tr -cd 'a-z0-9')"
  mkdir -p "${dir}"
  ( cd "${dir}" && "${TRABUCO_BIN}" init --name="${pname}" --group-id=com.trabuco.matrix \
      --modules="${mods}" --ai-agents=claude ${flags} ) > "${log}" 2>&1
  if grep -q "Maven build completed successfully" "${log}"; then
    st=PASS; err=""
  elif grep -qE "Maven build failed|BUILD FAILURE|^Error: |panic:" "${log}"; then
    st=FAIL; err="$(first_error "${log}")"
  else
    st=UNKNOWN; err="$(first_error "${log}")"
  fi
  printf '%s\t%s\t%s\n' "${name}" "${st}" "${err}" >> "${RESULTS}"
  printf '  %-6s %-26s %s\n' "${st}" "${name}" "${err}"
  exit 0
fi

# ── dispatcher mode ──────────────────────────────────────────────────────────
JOBS="${JOBS:-8}"
rm -rf "${WORK}"; mkdir -p "${WORK}"; : > "${RESULTS}"
echo "binary : ${TRABUCO_BIN}"; echo "workdir: ${WORK}"
echo "java   : $(java -version 2>&1 | head -1)"; echo "jobs   : ${JOBS}"; echo

EDGES=(API Worker EventConsumer AIAgent)
declare -A DS_MOD=( [none]="" [sqlpg]="SQLDatastore" [sqlmy]="SQLDatastore" [nomongo]="NoSQLDatastore" [noredis]="NoSQLDatastore" )
declare -A DS_FLAG=( [none]="" [sqlpg]="--database=postgresql" [sqlmy]="--database=mysql" [nomongo]="--nosql-database=mongodb" [noredis]="--nosql-database=redis" )
DS_KEYS=(none sqlpg sqlmy nomongo noredis)

JOBFILE="${WORK}/jobs.txt"; : > "${JOBFILE}"
for mask in $(seq 0 15); do
  edgesel=(); for i in 0 1 2 3; do (( (mask >> i) & 1 )) && edgesel+=("${EDGES[$i]}"); done
  for ds in "${DS_KEYS[@]}"; do
    mods="Model"; [[ -n "${DS_MOD[$ds]}" ]] && mods="${mods},${DS_MOD[$ds]}"
    flags="${DS_FLAG[$ds]}"
    for e in "${edgesel[@]}"; do mods="${mods},${e}"; done
    [[ " ${edgesel[*]} " == *" EventConsumer "* ]] && flags="${flags} --message-broker=kafka"
    [[ " ${edgesel[*]} " == *" AIAgent "* ]]        && flags="${flags} --vector-store=qdrant"
    en="$(IFS=-; echo "${edgesel[*]:-bare}")"
    printf '%s|%s|%s\n' "${en}__${ds}" "${mods}" "${flags}" >> "${JOBFILE}"
  done
done

total="$(wc -l < "${JOBFILE}" | tr -d ' ')"
echo "=== total combos: ${total} ==="
# Export only plain vars; re-invoke this script in worker mode per line.
export WORK RESULTS TRABUCO_BIN
xargs -P "${JOBS}" -I LINE bash "${BASH_SOURCE[0]}" __worker "LINE" < "${JOBFILE}"

echo; echo "================ SUMMARY ================"; echo "workdir: ${WORK}"
echo "PASS:    $(grep -c $'\tPASS\t' "${RESULTS}")"
echo "FAIL:    $(grep -c $'\tFAIL\t' "${RESULTS}")"
echo "UNKNOWN: $(grep -c $'\tUNKNOWN\t' "${RESULTS}")"
echo "---- non-PASS ----"
grep -v $'\tPASS\t' "${RESULTS}" | sort | awk -F'\t' '{printf "%-28s %-8s %s\n",$1,$2,$3}'
echo "DONE_MATRIX"

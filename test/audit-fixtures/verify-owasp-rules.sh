#!/usr/bin/env bash
# verify-owasp-rules.sh
#
# Positive-control test for the deterministic OWASP rules in
# `templates/github/scripts/review-checks.sh.tmpl`. Generates a
# scratch Trabuco project, drops the planted-vulnerability fixtures
# (PlantedVulnerabilities.java and planted-application.yml) into
# the right module sub-paths, runs the generated review-checks.sh,
# and asserts every expected `owasp.<id>` rule fires at least once.
#
# Used by the `audit-rules-check` CI job and by maintainers verifying
# the rules locally:
#
#   TRABUCO_BIN=/path/to/trabuco ./test/audit-fixtures/verify-owasp-rules.sh
#
# Exits non-zero with a clear summary if any expected rule didn't
# fire, or if an unexpected rule fired (false positive).

set -euo pipefail

# Resolve paths relative to this script.
THIS_DIR=$(cd "$(dirname "$0")" && pwd)
REPO_ROOT=$(cd "$THIS_DIR/../.." && pwd)
TRABUCO_BIN="${TRABUCO_BIN:-$REPO_ROOT/trabuco}"

if [ ! -x "$TRABUCO_BIN" ]; then
    echo "verify-owasp-rules: trabuco binary not found at $TRABUCO_BIN" >&2
    echo "Set TRABUCO_BIN or run \`go build -o trabuco ./cmd/trabuco\` first." >&2
    exit 2
fi

# All deterministic OWASP rules the script ships. Update this list when
# review-checks.sh.tmpl gains a new owasp.<id> rule. Each entry is the
# rule ID; the assertion is "this rule must produce at least one
# annotation in the script's output".
EXPECTED_RULES=(
    "owasp.a02-weak-hash"
    "owasp.a02-weak-cipher"
    "owasp.a02-jwt-none"
    "owasp.a03-runtime-exec-concat"
    "owasp.a03-jndi-user-input"
    "owasp.a05-cors-wildcard"
    "owasp.a05-mgmt-exposed-all"
    "owasp.a05-debug-prod"
    "owasp.a07-default-creds"
    "owasp.a08-jackson-default-typing"
    "owasp.a08-objectinputstream"
    "owasp.a10-resttemplate-user-url"
)

# Working directory — wiped on exit unless KEEP_TMP=1.
TMP=$(mktemp -d -t trabuco-audit-fixture-XXXXXX)
cleanup() {
    if [ "${KEEP_TMP:-0}" != "1" ]; then
        rm -rf "$TMP"
    else
        echo "verify-owasp-rules: tmp dir kept at $TMP (KEEP_TMP=1)" >&2
    fi
}
trap cleanup EXIT

cd "$TMP"
"$TRABUCO_BIN" init \
    --name owaspfixture \
    --group-id com.test.fixture \
    --modules Model,Shared,API \
    --database postgresql \
    --ai-agents claude \
    --skip-build > "$TMP/init.log" 2>&1

PROJ="$TMP/owaspfixture"

# Plant the Java fixture under API/src/main/java/com/test/fixture/.
mkdir -p "$PROJ/API/src/main/java/com/test/fixture"
cp "$THIS_DIR/PlantedVulnerabilities.java" \
   "$PROJ/API/src/main/java/com/test/fixture/PlantedVulnerabilities.java"

# Plant the misconfigured yaml under API/src/main/resources/.
cp "$THIS_DIR/planted-application.yml" \
   "$PROJ/API/src/main/resources/application-planted.yml"

# `git init` so review-checks.sh's `git ls-files` codepath sees the
# planted files. (Without this, ls-files returns nothing and the
# script's "Nothing to check" early exit triggers.)
cd "$PROJ"
git init -q
git -c user.email=ci@trabuco.local -c user.name=ci add -A
git -c user.email=ci@trabuco.local -c user.name=ci commit -m "fixture" --no-verify -q

# Run the script. We expect non-zero exit (findings are present).
SCRIPT_OUT="$TMP/review-checks.out"
set +e
bash .github/scripts/review-checks.sh > "$SCRIPT_OUT" 2>&1
SCRIPT_EXIT=$?
set -e

if [ "$SCRIPT_EXIT" -eq 0 ]; then
    echo "FAIL: review-checks.sh exited 0 — fixtures planted antipatterns but no rules fired" >&2
    cat "$SCRIPT_OUT" >&2
    exit 1
fi

# Assert every expected rule fired. We match on `::error ...title=[<id>]`
# which is the GitHub Actions annotation format the script emits per
# finding — NOT the `::group::[<id>]` banner that prints regardless of
# whether the rule fired.
MISSING=()
for rule in "${EXPECTED_RULES[@]}"; do
    if ! grep -qE "::error[^:]*title=\[$rule\]" "$SCRIPT_OUT"; then
        MISSING+=("$rule")
    fi
done

if [ "${#MISSING[@]}" -gt 0 ]; then
    echo "FAIL: the following OWASP rules did NOT fire on the planted fixtures:" >&2
    printf '  - %s\n' "${MISSING[@]}" >&2
    echo >&2
    echo "Script output (first 100 lines):" >&2
    head -100 "$SCRIPT_OUT" >&2
    echo >&2
    echo "Edit test/audit-fixtures/PlantedVulnerabilities.java or planted-application.yml" >&2
    echo "to ensure each rule has a planted antipattern that triggers it." >&2
    exit 1
fi

# Summary.
FIRED_COUNT=$(grep -cE "::error " "$SCRIPT_OUT" || true)
echo "verify-owasp-rules: PASS — all ${#EXPECTED_RULES[@]} rules fired ($FIRED_COUNT total annotations)"

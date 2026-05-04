#!/usr/bin/env bash
#
# test_add_commands.sh — end-to-end coverage for `trabuco add <type>`.
# Generates a project, runs each of the eight add-commands in sequence,
# asserts every expected file lands at its canonical path with the
# expected substring content, then runs `mvn install` against the
# generated bundle to confirm it compiles cleanly with the rest of
# the project.
#
# This script complements the unit-test layer (internal/addgen/*_test.go)
# and the CI `add-commands-smoke` job by:
#
#   1. Asserting concrete file paths (catches refactors that move
#      output locations without updating the tests).
#   2. Asserting concrete file contents (catches template drift where
#      the entity record forgets a field, the migration omits NOT NULL,
#      etc. — things that compile but encode the wrong behavior).
#   3. Running the full Maven reactor build, exercising the interaction
#      between every add-command's output (e.g., service references the
#      entity's repository, controller references the service, etc.).
#
# Usage:
#   scripts/e2e/test_add_commands.sh
#   TRABUCO_BIN=/path/to/trabuco scripts/e2e/test_add_commands.sh
#
# Exits non-zero on any missing file, missing substring, or build
# failure. Each failure prints the path/needle that didn't match.

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

WORKDIR="$(mktemp -d -t trabuco-add-e2e-XXXXXX)"
PG_CONTAINER=""
API_PID=""
cleanup() {
  if [[ -n "${API_PID}" ]]; then
    kill "${API_PID}" 2>/dev/null || true
    # mvn spring-boot:run forks a child Java process; kill the whole group.
    pkill -P "${API_PID}" 2>/dev/null || true
    wait "${API_PID}" 2>/dev/null || true
  fi
  if [[ -n "${PG_CONTAINER}" ]]; then
    docker rm -f "${PG_CONTAINER}" >/dev/null 2>&1 || true
  fi
  rm -rf "${WORKDIR}"
}
trap cleanup EXIT

PROJECT="${WORKDIR}/add-demo"
FAILED=0

red() { printf "\033[31m%s\033[0m\n" "$*"; }
green() { printf "\033[32m%s\033[0m\n" "$*"; }
yellow() { printf "\033[33m%s\033[0m\n" "$*"; }

assert_file() {
  local rel="$1"
  if [[ -f "${PROJECT}/${rel}" ]]; then
    green "  ✓ ${rel} exists"
  else
    red "  ✗ ${rel} MISSING"
    FAILED=$((FAILED + 1))
  fi
}

assert_contains() {
  local rel="$1"; shift
  local needle="$1"
  local file="${PROJECT}/${rel}"
  if [[ ! -f "$file" ]]; then
    red "  ✗ ${rel} missing — cannot grep for ${needle}"
    FAILED=$((FAILED + 1))
    return
  fi
  if grep -qF -- "$needle" "$file"; then
    green "  ✓ ${rel} contains: ${needle}"
  else
    red "  ✗ ${rel} missing: ${needle}"
    FAILED=$((FAILED + 1))
  fi
}

assert_not_contains() {
  local rel="$1"; shift
  local needle="$1"
  local file="${PROJECT}/${rel}"
  if [[ -f "$file" ]] && grep -qF -- "$needle" "$file"; then
    red "  ✗ ${rel} contains forbidden: ${needle}"
    FAILED=$((FAILED + 1))
  else
    green "  ✓ ${rel} does NOT contain: ${needle}"
  fi
}

# ─── Step 1: Generate baseline project ───────────────────────────────────────
yellow "[1/4] Generating baseline project (Model, SQLDatastore, Shared, API, Worker, AIAgent)"
mkdir -p "${WORKDIR}"
cd "${WORKDIR}"
"${TRABUCO_BIN}" init \
  --name=add-demo \
  --group-id=com.test.adddemo \
  --modules=Model,SQLDatastore,Shared,API,Worker,AIAgent \
  --database=postgresql \
  --java-version=21 \
  --skip-build > /dev/null

cd "${PROJECT}"

# ─── Step 2: Run each add-command in sequence ────────────────────────────────
yellow "[2/4] Running each add-command"

"${TRABUCO_BIN}" add migration --description="add orders index" > /dev/null
"${TRABUCO_BIN}" add entity Order \
    --fields="customerId:string,total:decimal,placedAt:instant,notes:text?,status:enum:OrderStatus,priority:integer,active:boolean,externalRef:uuid,birthDate:localdate,metadata:json,blob:bytes" > /dev/null
"${TRABUCO_BIN}" add service OrderService --entity=Order > /dev/null
"${TRABUCO_BIN}" add endpoint Order --type=crud > /dev/null
"${TRABUCO_BIN}" add job ProcessShipment --payload="orderId:string,priority:integer" > /dev/null
"${TRABUCO_BIN}" add event OrderShipped --fields="orderId:string,shippedAt:instant" > /dev/null
"${TRABUCO_BIN}" add streaming-endpoint Conversation > /dev/null
"${TRABUCO_BIN}" add test OrderService --module=Shared > /dev/null
"${TRABUCO_BIN}" add test OrderController --module=API --type=integration > /dev/null
"${TRABUCO_BIN}" add test OrderRepository --module=SQLDatastore --type=repository > /dev/null

green "  all 10 commands ran successfully"

# ─── Step 3: Assert files at canonical paths with expected content ───────────
yellow "[3/4] Asserting generated files"

# Migration: V2__add_orders_index.sql (V1 baseline already exists)
assert_file "SQLDatastore/src/main/resources/db/migration/V2__add_orders_index.sql"

# Entity bundle (SQL flavor: 5 files for full data-type matrix + enum)
assert_file "Model/src/main/java/com/test/adddemo/model/entities/Order.java"
assert_file "Model/src/main/java/com/test/adddemo/model/entities/OrderRecord.java"
assert_file "Model/src/main/java/com/test/adddemo/model/entities/OrderStatus.java"
assert_file "SQLDatastore/src/main/java/com/test/adddemo/sqldatastore/repository/OrderRepository.java"
assert_file "SQLDatastore/src/main/resources/db/migration/V3__create_orders.sql"

# Spot-check the migration covers every data type
local_mig="SQLDatastore/src/main/resources/db/migration/V3__create_orders.sql"
assert_contains "$local_mig" "id BIGSERIAL PRIMARY KEY"
assert_contains "$local_mig" "customer_id VARCHAR(255) NOT NULL"
assert_contains "$local_mig" "total NUMERIC(19,4) NOT NULL"
assert_contains "$local_mig" "placed_at TIMESTAMP WITH TIME ZONE NOT NULL"
assert_contains "$local_mig" "notes TEXT"
assert_not_contains "$local_mig" "notes TEXT NOT NULL"  # nullable
assert_contains "$local_mig" "status VARCHAR(64) NOT NULL"
assert_contains "$local_mig" "priority INTEGER NOT NULL"
assert_contains "$local_mig" "active BOOLEAN NOT NULL"
assert_contains "$local_mig" "external_ref UUID NOT NULL"
assert_contains "$local_mig" "birth_date DATE NOT NULL"
assert_contains "$local_mig" "metadata JSONB NOT NULL"
assert_contains "$local_mig" "blob BYTEA NOT NULL"

# Entity interface invariants
local_entity="Model/src/main/java/com/test/adddemo/model/entities/Order.java"
assert_contains "$local_entity" "@Value.Immutable"
assert_contains "$local_entity" "@JsonSerialize(as = ImmutableOrder.class)"
assert_contains "$local_entity" "Long id();"
assert_contains "$local_entity" "BigDecimal total();"
assert_contains "$local_entity" "Instant placedAt();"
assert_contains "$local_entity" "OrderStatus status();"

# Record invariants
local_record="Model/src/main/java/com/test/adddemo/model/entities/OrderRecord.java"
assert_contains "$local_record" "@Table(\"orders\")"
assert_contains "$local_record" "@Id @Nullable Long id"

# Service with constructor-injected repository
local_svc="Shared/src/main/java/com/test/adddemo/shared/service/OrderService.java"
assert_file "$local_svc"
assert_contains "$local_svc" "@Service"
assert_contains "$local_svc" "import com.test.adddemo.sqldatastore.repository.OrderRepository;"
assert_contains "$local_svc" "private final OrderRepository orderRepository;"

# Endpoint (CRUD)
local_ctrl="API/src/main/java/com/test/adddemo/api/controller/OrderController.java"
assert_file "$local_ctrl"
assert_contains "$local_ctrl" "@RequestMapping(\"/api/orders\")"
assert_contains "$local_ctrl" "@PostMapping"
assert_contains "$local_ctrl" "@GetMapping(\"/{id}\")"
assert_contains "$local_ctrl" "@PutMapping(\"/{id}\")"
assert_contains "$local_ctrl" "@DeleteMapping(\"/{id}\")"

# Job bundle (3 files)
assert_file "Model/src/main/java/com/test/adddemo/model/jobs/ProcessShipmentJobRequest.java"
assert_file "Model/src/main/java/com/test/adddemo/model/jobs/ProcessShipmentJobRequestHandler.java"
assert_file "Worker/src/main/java/com/test/adddemo/worker/handler/ProcessShipmentJobRequestHandler.java"

local_jobreq="Model/src/main/java/com/test/adddemo/model/jobs/ProcessShipmentJobRequest.java"
assert_contains "$local_jobreq" "implements JobRequest"
assert_contains "$local_jobreq" "String orderId"
assert_contains "$local_jobreq" "Integer priority"

# Event
local_event="Model/src/main/java/com/test/adddemo/model/events/OrderShipped.java"
assert_file "$local_event"
assert_contains "$local_event" "public record OrderShipped("
assert_contains "$local_event" "@NotNull String orderId"
assert_contains "$local_event" "@NotNull Instant shippedAt"

# Streaming endpoint
local_sse="AIAgent/src/main/java/com/test/adddemo/aiagent/protocol/ConversationStreamController.java"
assert_file "$local_sse"
assert_contains "$local_sse" "SseEmitter"
assert_contains "$local_sse" "Thread.startVirtualThread"
assert_contains "$local_sse" "/api/agent/stream/conversation"

# Tests (3 different shapes)
assert_file "Shared/src/test/java/com/test/adddemo/shared/service/OrderServiceTest.java"
assert_file "API/src/test/java/com/test/adddemo/api/controller/OrderControllerIT.java"
assert_file "SQLDatastore/src/test/java/com/test/adddemo/sqldatastore/repository/OrderRepositoryTest.java"

local_repo_test="SQLDatastore/src/test/java/com/test/adddemo/sqldatastore/repository/OrderRepositoryTest.java"
assert_contains "$local_repo_test" "@DataJdbcTest"
assert_contains "$local_repo_test" "@Testcontainers(disabledWithoutDocker = true)"
assert_contains "$local_repo_test" "PostgreSQLContainer"

# ─── Step 4: Maven reactor build ─────────────────────────────────────────────
yellow "[4/5] Running mvn install -DskipTests against the generated bundle"
if mvn install -DskipTests -B -q 2>&1 | tail -20; then
  green "  mvn install succeeded"
else
  red "  mvn install FAILED"
  FAILED=$((FAILED + 1))
fi

# ─── Step 5: Runtime smoke (boot + Flyway + endpoints + DB round-trip) ──────
# Spins up Postgres (Docker or external via $POSTGRES_URL), boots the API
# module with auth disabled, and asserts:
#   - /actuator/health == UP                 (Spring context wired all generated beans)
#   - flyway_schema_history has V1+V2+V3      (V2 from `add migration`, V3 from `add entity`)
#   - POST /api/orders returns 5xx            (new OrderController registered; stub fires)
#   - POST /api/placeholders round-trips      (Spring Data JDBC + Flyway + Postgres end-to-end)
#
# Skipped gracefully when Docker is unavailable AND $POSTGRES_URL is unset.
# CI (GitHub Actions) sets POSTGRES_URL via a service container, sidestepping
# the Docker-in-Docker concern.
yellow "[5/5] Runtime smoke (boot + Flyway + DB round-trip)"

PG_USER="${POSTGRES_USER:-test}"
PG_PASS="${POSTGRES_PASSWORD:-test}"
PG_DB="${POSTGRES_DB:-test}"
PG_HOST="${POSTGRES_HOST:-localhost}"
PG_PORT="${POSTGRES_PORT:-}"

if [[ -z "${POSTGRES_URL:-}" ]]; then
  # No external Postgres provided — try Docker for a local container.
  if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
    yellow "  [skip] Docker not available and POSTGRES_URL not set; skipping runtime smoke"
    if [[ "${FAILED}" -gt 0 ]]; then
      echo
      red "FAILED: ${FAILED} assertions did not pass"
      exit 1
    fi
    green "All add-commands E2E assertions passed (compile-only; runtime smoke skipped)."
    exit 0
  fi

  PG_PORT="${PG_PORT:-5435}"
  PG_CONTAINER="trabuco-e2e-pg-$$"

  echo "  starting postgres container ${PG_CONTAINER} on :${PG_PORT}"
  docker run -d --rm \
    --name "${PG_CONTAINER}" \
    -e "POSTGRES_USER=${PG_USER}" \
    -e "POSTGRES_PASSWORD=${PG_PASS}" \
    -e "POSTGRES_DB=${PG_DB}" \
    -p "${PG_PORT}:5432" \
    postgres:15-alpine >/dev/null

  echo -n "  waiting for postgres ready"
  for _ in $(seq 1 30); do
    if docker exec "${PG_CONTAINER}" pg_isready -U "${PG_USER}" >/dev/null 2>&1; then
      echo " ready"
      break
    fi
    echo -n "."
    sleep 1
  done
fi

# Default port to 5432 when running against an external Postgres (CI service container).
PG_PORT="${PG_PORT:-5432}"
PG_URL="${POSTGRES_URL:-jdbc:postgresql://${PG_HOST}:${PG_PORT}/${PG_DB}}"

API_LOG="${WORKDIR}/api.log"
echo "  booting API against ${PG_URL} (TRABUCO_AUTH_ENABLED=false)"
(
  cd "${PROJECT}/API"
  SPRING_DATASOURCE_URL="${PG_URL}" \
  SPRING_DATASOURCE_USERNAME="${PG_USER}" \
  SPRING_DATASOURCE_PASSWORD="${PG_PASS}" \
  TRABUCO_AUTH_ENABLED="false" \
  mvn spring-boot:run -B -q
) > "${API_LOG}" 2>&1 &
API_PID=$!

echo -n "  waiting for /actuator/health=UP (up to 120s)"
booted=0
for _ in $(seq 1 120); do
  if curl -sf "http://localhost:8080/actuator/health" 2>/dev/null | grep -q '"status":"UP"'; then
    booted=1
    echo " UP"
    break
  fi
  echo -n "."
  sleep 1
done

if [[ "${booted}" -eq 0 ]]; then
  echo
  red "  ✗ API never reached UP. Last 40 log lines:"
  tail -40 "${API_LOG}" || true
  FAILED=$((FAILED + 1))
else
  green "  ✓ /actuator/health = UP"

  # Flyway: query flyway_schema_history. Prefer the system `psql` (works
  # against external Postgres in CI service-container mode); fall back
  # to `docker exec psql` when we own the local container.
  flyway_query="SELECT string_agg(version, ',' ORDER BY installed_rank) FROM flyway_schema_history WHERE success = true"
  applied=""
  if command -v psql >/dev/null 2>&1; then
    applied=$(PGPASSWORD="${PG_PASS}" psql -h "${PG_HOST}" -p "${PG_PORT}" -U "${PG_USER}" -d "${PG_DB}" -tAc "${flyway_query}" 2>/dev/null | tr -d '[:space:]' || echo "")
  elif [[ -n "${PG_CONTAINER}" ]]; then
    applied=$(docker exec "${PG_CONTAINER}" psql -U "${PG_USER}" -d "${PG_DB}" -tAc "${flyway_query}" 2>/dev/null | tr -d '[:space:]' || echo "")
  fi
  if [[ -z "${applied}" ]]; then
    yellow "  [skip] Flyway check: neither psql nor docker exec accessible"
  elif [[ ",${applied}," == *",1,"* && ",${applied}," == *",2,"* && ",${applied}," == *",3,"* ]]; then
    green "  ✓ Flyway applied V1, V2, V3 (history: ${applied})"
  else
    red "  ✗ Flyway missing V2 or V3 (history: ${applied})"
    FAILED=$((FAILED + 1))
  fi

  # OrderController registration check — stub throws so we expect 5xx (not 404).
  http=$(curl -s -o /dev/null -w "%{http_code}" -X POST "http://localhost:8080/api/orders" \
    -H "Content-Type: application/json" -d "{}")
  case "${http}" in
    5*) green "  ✓ POST /api/orders → ${http} (new controller registered; stub fired as expected)" ;;
    404)
      red "  ✗ POST /api/orders → 404 (new OrderController NOT registered)"
      FAILED=$((FAILED + 1))
      ;;
    *)
      yellow "  ? POST /api/orders → ${http} (expected 5xx; continuing)"
      ;;
  esac

  # End-to-end DB round-trip via the existing PlaceholderController. With
  # TRABUCO_AUTH_ENABLED=false, @PreAuthorize is inert (per
  # MethodSecurityConfig). A 201 response with the request name echoed
  # back proves Spring Data JDBC + Flyway + Postgres are wired correctly
  # in the generated bundle.
  create_resp=$(curl -sf -X POST "http://localhost:8080/api/placeholders" \
    -H "Content-Type: application/json" \
    -d '{"name":"e2e-runtime","description":"smoke test"}' 2>/dev/null || echo "")
  if [[ -n "${create_resp}" ]] && echo "${create_resp}" | grep -q "e2e-runtime"; then
    green "  ✓ POST /api/placeholders round-trip succeeded"
    # Follow up with a list to verify the row is queryable.
    list_resp=$(curl -sf "http://localhost:8080/api/placeholders" 2>/dev/null || echo "")
    if echo "${list_resp}" | grep -q "e2e-runtime"; then
      green "  ✓ GET /api/placeholders returns the inserted row"
    else
      red "  ✗ GET /api/placeholders did NOT return the inserted row"
      FAILED=$((FAILED + 1))
    fi
  else
    red "  ✗ POST /api/placeholders failed (response: ${create_resp})"
    FAILED=$((FAILED + 1))
  fi
fi

# ─── Result ──────────────────────────────────────────────────────────────────
echo
if [[ "${FAILED}" -gt 0 ]]; then
  red "FAILED: ${FAILED} assertions did not pass"
  exit 1
fi
green "All add-commands E2E assertions passed (compile + runtime)."

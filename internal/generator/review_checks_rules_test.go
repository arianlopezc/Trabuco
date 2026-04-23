//go:build integration

package generator

import (
	"testing"
)

// One rule per table-driven subtest, with three mandatory cases each:
//   - positive:    a minimal fixture that MUST trigger the rule
//   - negative:    a fixture that looks superficially similar but MUST NOT
//                  trigger the rule (guards against overly-greedy regexes)
//   - suppression: the positive fixture with `// trabuco-allow: <rule>`
//                  so we assert is_suppressed() actually works
//
// Kept as separate sub-tests (not one mega-fixture) because a regression in
// a single rule should isolate to a single subtest failure.

// ─── Spring patterns ─────────────────────────────────────────────────────────

func TestReviewChecks_Spring_FieldInjection(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "API"))
		writeJava(t, dir, "API/src/main/java/Bad.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Bad {
  @Autowired
  private Thing t;
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertFinding(t, findings, "spring.field-injection", "API/src/main/java/Bad.java")
	})
	t.Run("negative_constructor_injection", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "API"))
		// Constructor injection with @Autowired on the constructor is a known
		// idiom some codebases still use and is NOT what we flag.
		writeJava(t, dir, "API/src/main/java/Good.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Good {
  private final Thing t;
  @Autowired public Good(Thing t) { this.t = t; }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "spring.field-injection", "API/src/main/java/Good.java")
	})
	t.Run("suppression_same_line", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "API"))
		writeJava(t, dir, "API/src/main/java/Suppressed.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Suppressed {
  @Autowired // trabuco-allow: spring.field-injection
  private Thing t;
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "spring.field-injection", "API/src/main/java/Suppressed.java")
	})
}

func TestReviewChecks_Spring_TransactionalPrivate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/TxBad.java", `
package x;
import org.springframework.transaction.annotation.Transactional;
public class TxBad {
  @Transactional
  private void doWork() { /* ... */ }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertFinding(t, findings, "spring.tx-private", "Shared/src/main/java/TxBad.java")
	})
	t.Run("negative_public_method", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/TxGood.java", `
package x;
import org.springframework.transaction.annotation.Transactional;
public class TxGood {
  @Transactional
  public void doWork() { /* ... */ }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "spring.tx-private", "Shared/src/main/java/TxGood.java")
	})
	t.Run("suppression_previous_line", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "Shared"))
		// Suppression on the @Transactional line (which is where the script
		// reports the finding) should silence it.
		writeJava(t, dir, "Shared/src/main/java/TxSup.java", `
package x;
import org.springframework.transaction.annotation.Transactional;
public class TxSup {
  // trabuco-allow: spring.tx-private
  @Transactional
  private void doWork() { /* ... */ }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "spring.tx-private", "Shared/src/main/java/TxSup.java")
	})
}

// ─── Secrets ────────────────────────────────────────────────────────────────

func TestReviewChecks_Secrets_AWSKey(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model"))
	writeJava(t, dir, "Model/src/main/java/Keys.java",
		`package x; class K { static final String AWS = "AKIAIOSFODNN7EXAMPLE"; }`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "sec.aws-key", "Model/src/main/java/Keys.java")
}

func TestReviewChecks_Secrets_OpenAIKey(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model"))
	writeJava(t, dir, "Model/src/main/java/Keys.java",
		`package x; class K { static final String OAI = "sk-abcdefghijklmnopqrstuvwxyz012345"; }`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "sec.openai-key", "Model/src/main/java/Keys.java")
}

func TestReviewChecks_Secrets_HardcodedLiteral(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model"))
		writeJava(t, dir, "Model/src/main/java/Config.java",
			`package x; class C { static final String password = "supersecretvalue1234"; }`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertFinding(t, findings, "sec.hardcoded-secret", "Model/src/main/java/Config.java")
	})
	t.Run("negative_short_value", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model"))
		// Rule requires 16+ chars; 8-char literal is likely a test dummy.
		writeJava(t, dir, "Model/src/main/java/ShortC.java",
			`package x; class C { static final String password = "short"; }`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "sec.hardcoded-secret", "Model/src/main/java/ShortC.java")
	})
	t.Run("scans_test_sources_too", func(t *testing.T) {
		// Secrets checks deliberately scan main+test — a leaked AWS key in
		// a test file is still a leak.
		dir := setupReviewProject(t, projectCfg("Model"))
		writeJava(t, dir, "Model/src/test/java/KeysTest.java",
			`package x; class K { static final String AWS = "AKIAIOSFODNN7EXAMPLE"; }`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertFinding(t, findings, "sec.aws-key", "Model/src/test/java/KeysTest.java")
	})
}

// ─── Datastore performance (SQLDatastore-gated) ──────────────────────────────

func TestReviewChecks_Perf_NPlusOne(t *testing.T) {
	t.Run("positive_stream", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/NPO.java", `
package x;
import java.util.List;
public class NPO {
  java.util.List<Thing> hydrate(java.util.List<Long> ids) {
    return ids.stream()
      .map(id -> repo.findById(id).orElseThrow())
      .toList();
  }
  Repo repo;
  interface Repo { java.util.Optional<Thing> findById(Long id); }
  static class Thing {}
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertFinding(t, findings, "perf.n-plus-one", "Shared/src/main/java/NPO.java")
	})
	t.Run("positive_for_loop", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/NPOFor.java", `
package x;
public class NPOFor {
  void run(java.util.List<Long> ids) {
    for (Long id : ids) {
      repo.findById(id);
    }
  }
  Repo repo;
  interface Repo { Object findById(Long id); }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertFinding(t, findings, "perf.n-plus-one", "Shared/src/main/java/NPOFor.java")
	})
	t.Run("negative_standalone_findById", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/Single.java", `
package x;
public class Single {
  Object one(Long id) { return repo.findById(id); }
  Repo repo;
  interface Repo { Object findById(Long id); }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "perf.n-plus-one", "Shared/src/main/java/Single.java")
	})
}

func TestReviewChecks_Perf_UnboundedScan(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/Scan.java", `
package x;
public class Scan {
  Object all() { return repo.findAll(); }
  Repo repo;
  interface Repo { Object findAll(); }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertFinding(t, findings, "perf.unbounded-scan", "Shared/src/main/java/Scan.java")
	})
	t.Run("suppression_same_line", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/ScanSup.java", `
package x;
public class ScanSup {
  Object all() { return repo.findAll(); // trabuco-allow: perf.unbounded-scan
  }
  Repo repo;
  interface Repo { Object findAll(); }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "perf.unbounded-scan", "Shared/src/main/java/ScanSup.java")
	})
	t.Run("suppression_all_rules", func(t *testing.T) {
		dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
		writeJava(t, dir, "Shared/src/main/java/ScanAll.java", `
package x;
public class ScanAll {
  // trabuco-allow: all
  Object all() { return repo.findAll(); }
  Repo repo;
  interface Repo { Object findAll(); }
}
`)
		findings, _ := runReviewChecks(t, dir, "all")
		assertNoFinding(t, findings, "perf.unbounded-scan", "Shared/src/main/java/ScanAll.java")
	})
}

func TestReviewChecks_Perf_OffsetPagination(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
	writeJava(t, dir, "Shared/src/main/java/Page.java", `
package x;
import org.springframework.data.domain.Pageable;
public class Page {
  Object fetch(Pageable p) { return null; }
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "perf.offset-pagination", "Shared/src/main/java/Page.java")
}

func TestReviewChecks_Perf_SQLConcat(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore"))
	writeJava(t, dir, "SQLDatastore/src/main/java/Repo.java", `
package x;
import org.springframework.data.jpa.repository.Query;
public interface Repo {
  @Query("SELECT u FROM User u WHERE u.name = '" + "foo" + "'")
  Object find();
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "perf.sql-concat", "SQLDatastore/src/main/java/Repo.java")
}

func TestReviewChecks_Perf_UnboundedDML(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore"))
	writeJava(t, dir, "SQLDatastore/src/main/java/R.java", `
package x;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
public interface R {
  @Modifying
  @Query("DELETE FROM User u WHERE u.active = false")
  int purge();
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "perf.unbounded-dml", "SQLDatastore/src/main/java/R.java")
}

// ─── SQL migration checks ────────────────────────────────────────────────────

func TestReviewChecks_SQL_ForeignKey(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore"))
	// The script short-circuits with "No Java files in scope" when it finds
	// zero Java sources — real projects always ship Java next to their
	// migrations, so the stub keeps the fixture realistic.
	writeJava(t, dir, "SQLDatastore/src/main/java/Stub.java",
		`package x; public class Stub {}`)
	writeSQL(t, dir,
		"SQLDatastore/src/main/resources/db/migration/V1__init.sql",
		`CREATE TABLE orders (id BIGINT PRIMARY KEY, user_id BIGINT REFERENCES users(id));`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "sql.foreign-key",
		"SQLDatastore/src/main/resources/db/migration/V1__init.sql")
}

func TestReviewChecks_SQL_UnboundedDML(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore"))
	writeJava(t, dir, "SQLDatastore/src/main/java/Stub.java",
		`package x; public class Stub {}`)
	writeSQL(t, dir,
		"SQLDatastore/src/main/resources/db/migration/V2__purge.sql",
		`DELETE FROM users WHERE created_at < NOW() - INTERVAL '1 year';`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "sql.unbounded-dml",
		"SQLDatastore/src/main/resources/db/migration/V2__purge.sql")
}

// ─── NoSQL checks (NoSQLDatastore-gated) ────────────────────────────────────

func TestReviewChecks_NoSQL_RedisKeys(t *testing.T) {
	cfg := projectCfg("Model", "NoSQLDatastore", "Shared")
	cfg.NoSQLDatabase = "redis"
	dir := setupReviewProject(t, cfg)
	writeJava(t, dir, "Shared/src/main/java/R.java", `
package x;
public class R {
  void scan() { redis.execute("KEYS *"); }
  Redis redis;
  interface Redis { Object execute(String cmd); }
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "nosql.redis-keys", "Shared/src/main/java/R.java")
}

func TestReviewChecks_NoSQL_MongoWhereEval(t *testing.T) {
	cfg := projectCfg("Model", "NoSQLDatastore", "Shared")
	cfg.NoSQLDatabase = "mongodb"
	dir := setupReviewProject(t, cfg)
	writeJava(t, dir, "Shared/src/main/java/M.java", `
package x;
import org.springframework.data.mongodb.repository.Query;
public interface M {
  @Query("{ $where: 'this.name.length > 5' }")
  Object find();
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "nosql.mongo-where-eval", "Shared/src/main/java/M.java")
}

// ─── API controller checks (API-gated) ──────────────────────────────────────

func TestReviewChecks_API_RedundantTryCatch(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "API"))
	// The rule is scoped to `/api/controller/*.java` paths only — worker code
	// legitimately catches/logs/returns, so we must not flag those.
	writeJava(t, dir, "API/src/main/java/api/controller/Ctrl.java", `
package x.api.controller;
import org.springframework.http.ResponseEntity;
public class Ctrl {
  ResponseEntity<Object> h() {
    try { return ResponseEntity.ok(null); }
    catch (Exception e) { return ResponseEntity.status(500).build(); }
  }
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "api.redundant-try-catch",
		"API/src/main/java/api/controller/Ctrl.java")
}

func TestReviewChecks_API_RedundantTryCatch_WorkerExempt(t *testing.T) {
	// Identical pattern in a worker handler must not fire — retry/DLQ wiring
	// often requires exactly this shape.
	dir := setupReviewProject(t, projectCfg("Model", "API", "Worker"))
	writeJava(t, dir, "Worker/src/main/java/worker/handler/H.java", `
package x.worker.handler;
import org.springframework.http.ResponseEntity;
public class H {
  ResponseEntity<Object> h() {
    try { return ResponseEntity.ok(null); }
    catch (Exception e) { return ResponseEntity.status(500).build(); }
  }
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertNoFinding(t, findings, "api.redundant-try-catch",
		"Worker/src/main/java/worker/handler/H.java")
}

// ─── Boundary checks ────────────────────────────────────────────────────────

func TestReviewChecks_Boundary_APIImportsWorker(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "API", "Worker"))
	writeJava(t, dir, "API/src/main/java/Leak.java", `
package x;
import com.trabuco.fixture.worker.Handler;
public class Leak { Handler h; }
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "boundary.api-imports-worker", "API/src/main/java/Leak.java")
}

// ─── Module gating ──────────────────────────────────────────────────────────

// Module gating is asserted by *runtime behavior*, not by grepping the script
// body — the rule IDs appear in the script's user-facing suppression hint
// regardless of gating (e.g., `# trabuco-allow: perf.unbounded-scan`), so a
// textual search produces false positives.
//
// Instead: drop a fixture that WOULD trigger the rule if the rule were active,
// run the script, and assert zero findings for that rule. That's the actual
// property users care about — "rule doesn't fire on minimal projects" —
// rather than an implementation detail of the template.
func TestReviewChecks_ModuleGating_NoDatastore_RulesSilent(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "API"))

	// This fixture would trigger perf.unbounded-scan and perf.n-plus-one if
	// those rules were emitted. They must not be emitted for API-only projects.
	writeJava(t, dir, "API/src/main/java/Would.java", `
package x;
public class Would {
  Object all() { return repo.findAll(); }
  void loop(java.util.List<Long> ids) {
    for (Long id : ids) { repo.findById(id); }
  }
  Repo repo;
  interface Repo { Object findAll(); Object findById(Long id); }
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	for _, rule := range []string{
		"perf.n-plus-one",
		"perf.unbounded-scan",
		"perf.offset-pagination",
		"perf.sql-concat",
		"sql.foreign-key",
		"nosql.redis-keys",
	} {
		assertNoFinding(t, findings, rule, "API/src/main/java/Would.java")
	}
}

func TestReviewChecks_ModuleGating_SQLDatastore_RulesActive(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))

	// Same shape as above — but in a datastore project, the rule MUST fire.
	writeJava(t, dir, "Shared/src/main/java/Active.java", `
package x;
public class Active {
  Object all() { return repo.findAll(); }
  Repo repo;
  interface Repo { Object findAll(); }
}
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertFinding(t, findings, "perf.unbounded-scan", "Shared/src/main/java/Active.java")
}

func TestReviewChecks_ModuleGating_NoSQLExclusion(t *testing.T) {
	// SQL-only projects must not carry NoSQL rules (avoids false positives
	// on code that legitimately uses `KEYS *` in a non-Redis context).
	dir := setupReviewProject(t, projectCfg("Model", "SQLDatastore", "Shared"))
	writeJava(t, dir, "Shared/src/main/java/Keys.java", `
package x;
public class Keys { void scan() { redis.execute("KEYS *"); } Redis redis; interface Redis { Object execute(String s); } }
`)
	findings, _ := runReviewChecks(t, dir, "all")
	assertNoFinding(t, findings, "nosql.redis-keys", "Shared/src/main/java/Keys.java")
}

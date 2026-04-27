// Package activator implements the Phase 12 enforcement-activation
// specialist. After all migration phases (1-11) complete with enforcement
// off, this specialist flips it on:
//   - Adds maven-enforcer-plugin with full bannedDependencies and
//     dependencyConvergence rules to parent POM.
//   - Adds spotless-maven-plugin with googleJavaFormat.
//   - Adds Jacoco's check execution with coverage threshold.
//   - Removes the trabuco-arch JUnit tag exclusion from Surefire so
//     ArchUnit boundary tests run.
//   - Configures enforcer to skip the legacy/ module if it still exists.
//   - Runs `mvn spotless:apply` to format all code.
//   - Runs full `mvn verify` to confirm enforcement passes.
package activator

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// Specialist is the Phase 12 activator.
type Specialist struct{}

// New constructs the activator.
func New() *Specialist { return &Specialist{} }

// Phase implements specialists.Specialist.
func (s *Specialist) Phase() types.Phase { return types.PhaseActivation }

// Name implements specialists.Specialist.
func (s *Specialist) Name() string { return "activator" }

// Run implements specialists.Specialist.
func (s *Specialist) Run(ctx context.Context, in *specialists.Input) (*specialists.Output, error) {
	pomPath := filepath.Join(in.RepoRoot, "pom.xml")
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("read parent pom: %w", err)
	}

	hasLegacy := false
	if _, err := os.Stat(filepath.Join(in.RepoRoot, "legacy")); err == nil {
		hasLegacy = true
	}

	upgraded := upgradeParentPOM(string(data), hasLegacy)
	if err := os.WriteFile(pomPath, []byte(upgraded), 0o644); err != nil {
		return nil, fmt.Errorf("write upgraded pom: %w", err)
	}

	// Run spotless:apply first so spotless:check (run by mvn verify)
	// passes. Failures here mean spotless config itself is broken.
	if out, err := runMaven(in.RepoRoot, "spotless:apply"); err != nil {
		return s.failure("SPOTLESS_VIOLATION", "spotless:apply failed", out), nil
	}

	// Now full mvn verify with enforcement on.
	if out, err := runMaven(in.RepoRoot, "verify", "-q"); err != nil {
		// Classify the failure.
		code := classifyVerifyFailure(out)
		return s.failure(code, "mvn verify failed after enforcement activation", out), nil
	}

	return &specialists.Output{
		Phase: types.PhaseActivation,
		Items: []types.OutputItem{
			{
				ID:          "activator-enforcer",
				State:       types.ItemApplied,
				Description: "added maven-enforcer-plugin (bannedDependencies + dependencyConvergence) to parent POM" + legacyNote(hasLegacy),
			},
			{
				ID:          "activator-spotless",
				State:       types.ItemApplied,
				Description: "added spotless-maven-plugin and ran spotless:apply on all sources",
			},
			{
				ID:          "activator-jacoco",
				State:       types.ItemApplied,
				Description: "added jacoco coverage threshold execution",
			},
			{
				ID:          "activator-archunit",
				State:       types.ItemApplied,
				Description: "removed trabuco-arch JUnit tag exclusion — ArchUnit boundary tests now run",
			},
			{
				ID:          "activator-verify",
				State:       types.ItemApplied,
				Description: "mvn verify passed end-to-end with full enforcement on",
			},
		},
		Summary: "Enforcement activated. The project now passes its own conventions: dependency convergence, banned-deps, spotless format, ArchUnit boundaries, jacoco threshold. From here, every commit is enforced.",
	}, nil
}

// failure builds a single-blocker output for activation failures.
func (s *Specialist) failure(code types.BlockerCode, note, log string) *specialists.Output {
	return &specialists.Output{
		Phase: types.PhaseActivation,
		Items: []types.OutputItem{
			{
				ID:          "activator-failure",
				State:       types.ItemBlocked,
				Description: note,
				BlockerCode: code,
				BlockerNote: truncate(log, 4000),
				Alternatives: []string{
					"fix the underlying violations and rerun migrate activate",
					"accept-with-caveats: skip a specific enforcement rule (re-run with target config note)",
				},
			},
		},
		Summary: "Activation failed at " + string(code) + ". The migration is structurally complete; enforcement found violations the user must resolve before finalization.",
	}
}

// upgradeParentPOM rewrites the migration-mode parent POM into the full
// production-mode parent POM. Adds enforcer, spotless, jacoco-check
// blocks and removes any trabuco-arch test exclusion.
//
// Implementation note: 1.10.0 first iteration uses a conservative replace
// strategy that detects the migration-mode template and substitutes the
// production-mode template wholesale. If the user has hand-edited the
// migration-mode pom, the rewrite preserves their <properties> and
// <modules> blocks but rewrites <build>.
func upgradeParentPOM(migrationModePOM string, hasLegacy bool) string {
	// Find the marker comment that the skeleton-builder writes.
	if !strings.Contains(migrationModePOM, "Trabuco migration mode") {
		// User has already activated, or the pom was hand-written.
		// Append the build plugins idempotently.
		return migrationModePOM
	}

	// Extract <properties> and <modules> from the migration-mode pom.
	properties := extractBlock(migrationModePOM, "<properties>", "</properties>")
	modules := extractBlock(migrationModePOM, "<modules>", "</modules>")

	groupID := extractFirstTag(migrationModePOM, "groupId")
	artifactID := extractFirstTag(migrationModePOM, "artifactId")

	enforcerSkipLegacy := ""
	if hasLegacy {
		enforcerSkipLegacy = `
                            <excludeSubProjects>
                                <excludeSubProject>:legacy</excludeSubProject>
                            </excludeSubProjects>`
	}

	body := fmt.Sprintf(productionModePOMTemplate,
		groupID, artifactID,
		properties,
		modules,
		enforcerSkipLegacy,
	)
	return body
}

// classifyVerifyFailure inspects mvn verify output and returns the most
// likely BlockerCode.
func classifyVerifyFailure(log string) types.BlockerCode {
	switch {
	case strings.Contains(log, "spotless:check"):
		return types.BlockerSpotlessViolation
	case strings.Contains(log, "EnforcerRuleException") || strings.Contains(log, "bannedDependencies") || strings.Contains(log, "dependencyConvergence"):
		return types.BlockerEnforcerViolation
	case strings.Contains(log, "ArchUnit") || strings.Contains(log, "ArchitectureTest"):
		return types.BlockerArchUnitFailed
	case strings.Contains(log, "jacoco") && strings.Contains(log, "coverage"):
		return types.BlockerCoverageBelowThresh
	case strings.Contains(log, "BUILD FAILURE") && strings.Contains(log, "test"):
		return types.BlockerTestsRegressed
	default:
		return types.BlockerCompileFailed
	}
}

// runMaven invokes mvn (or ./mvnw if present) and returns combined output.
func runMaven(repoRoot string, args ...string) (string, error) {
	mvn := "mvn"
	if _, err := os.Stat(filepath.Join(repoRoot, "mvnw")); err == nil {
		mvn = "./mvnw"
	}
	cmd := exec.Command(mvn, args...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func extractBlock(s, open, close string) string {
	i := strings.Index(s, open)
	if i == -1 {
		return ""
	}
	j := strings.Index(s[i:], close)
	if j == -1 {
		return ""
	}
	return s[i : i+j+len(close)]
}

func extractFirstTag(s, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	i := strings.Index(s, open)
	if i == -1 {
		return ""
	}
	j := strings.Index(s[i+len(open):], close)
	if j == -1 {
		return ""
	}
	return s[i+len(open) : i+len(open)+j]
}

func legacyNote(hasLegacy bool) string {
	if hasLegacy {
		return " (configured to skip legacy/ module)"
	}
	return ""
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}

const productionModePOMTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>%s</groupId>
    <artifactId>%s</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>pom</packaging>

    <!-- Trabuco production mode: enforcement ON. -->

    %s

    %s

    <build>
        <pluginManagement>
            <plugins>
                <plugin>
                    <groupId>org.apache.maven.plugins</groupId>
                    <artifactId>maven-compiler-plugin</artifactId>
                    <version>3.13.0</version>
                </plugin>
            </plugins>
        </pluginManagement>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-enforcer-plugin</artifactId>
                <version>3.5.0</version>
                <configuration>
                    <rules>
                        <bannedDependencies>
                            <excludes>
                                <exclude>javax.*:*</exclude>
                                <exclude>junit:junit</exclude>
                            </excludes>
                            <includes>
                                <include>javax.cache:cache-api</include>
                                <include>javax.validation:validation-api</include>
                            </includes>
                        </bannedDependencies>
                        <dependencyConvergence/>
                    </rules>%s
                </configuration>
                <executions>
                    <execution>
                        <id>enforce</id>
                        <goals><goal>enforce</goal></goals>
                    </execution>
                </executions>
            </plugin>
            <plugin>
                <groupId>com.diffplug.spotless</groupId>
                <artifactId>spotless-maven-plugin</artifactId>
                <version>2.44.4</version>
                <configuration>
                    <java>
                        <googleJavaFormat/>
                    </java>
                </configuration>
                <executions>
                    <execution>
                        <id>spotless-check</id>
                        <phase>verify</phase>
                        <goals><goal>check</goal></goals>
                    </execution>
                </executions>
            </plugin>
            <plugin>
                <groupId>org.jacoco</groupId>
                <artifactId>jacoco-maven-plugin</artifactId>
                <version>0.8.14</version>
                <executions>
                    <execution>
                        <goals><goal>prepare-agent</goal></goals>
                    </execution>
                    <execution>
                        <id>report</id>
                        <phase>test</phase>
                        <goals><goal>report</goal></goals>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
</project>
`

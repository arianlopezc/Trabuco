// Package types defines the shared vocabulary used across the migration
// subsystem: phases, output-item states, blocker reason codes, and the
// source_evidence struct that backs the no-out-of-scope contract (see
// docs/MIGRATION_REDESIGN_PLAN.md §4).
package types

// Phase identifies a migration phase. The numeric values are the canonical
// phase indices used in state.json, git tags, and gate UX.
type Phase int

const (
	PhaseAssessment    Phase = 0
	PhaseSkeleton      Phase = 1
	PhaseModel         Phase = 2
	PhaseDatastore     Phase = 3
	PhaseShared        Phase = 4
	PhaseAPI           Phase = 5
	PhaseWorker        Phase = 6
	PhaseEventConsumer Phase = 7
	PhaseAIAgent       Phase = 8
	PhaseConfiguration Phase = 9
	PhaseDeployment    Phase = 10
	PhaseTests         Phase = 11
	PhaseActivation    Phase = 12
	PhaseFinalization  Phase = 13
)

// String returns the canonical short name for the phase, used in state.json
// keys and git tag names.
func (p Phase) String() string {
	switch p {
	case PhaseAssessment:
		return "assessment"
	case PhaseSkeleton:
		return "skeleton"
	case PhaseModel:
		return "model"
	case PhaseDatastore:
		return "datastore"
	case PhaseShared:
		return "shared"
	case PhaseAPI:
		return "api"
	case PhaseWorker:
		return "worker"
	case PhaseEventConsumer:
		return "eventconsumer"
	case PhaseAIAgent:
		return "aiagent"
	case PhaseConfiguration:
		return "configuration"
	case PhaseDeployment:
		return "deployment"
	case PhaseTests:
		return "tests"
	case PhaseActivation:
		return "activation"
	case PhaseFinalization:
		return "finalization"
	default:
		return "unknown"
	}
}

// AllPhases returns every phase in execution order.
func AllPhases() []Phase {
	return []Phase{
		PhaseAssessment, PhaseSkeleton, PhaseModel, PhaseDatastore,
		PhaseShared, PhaseAPI, PhaseWorker, PhaseEventConsumer,
		PhaseAIAgent, PhaseConfiguration, PhaseDeployment, PhaseTests,
		PhaseActivation, PhaseFinalization,
	}
}

// PhaseStateLabel is the lifecycle label for a single phase in state.json.
type PhaseStateLabel string

const (
	PhasePending       PhaseStateLabel = "pending"
	PhaseInProgress    PhaseStateLabel = "in_progress"
	PhaseCompleted     PhaseStateLabel = "completed"
	PhaseUserRejected  PhaseStateLabel = "user_rejected"
	PhaseNotApplicable PhaseStateLabel = "not_applicable"
	PhaseFailed        PhaseStateLabel = "failed"
)

// ItemState is the state of a single specialist output item (see §3 of plan
// — every patch is in one of these five states).
type ItemState string

const (
	// ItemApplied is a patch ready to commit.
	ItemApplied ItemState = "applied"
	// ItemBlocked is a piece the specialist could not migrate; carries a
	// BlockerCode and at least one alternative.
	ItemBlocked ItemState = "blocked"
	// ItemRequiresDecision means multiple legal Trabuco-shaped outcomes
	// exist; the user must choose.
	ItemRequiresDecision ItemState = "requires_decision"
	// ItemNotApplicable is the explicit happy path for "this phase has
	// nothing to do because nothing in source matches its scope."
	ItemNotApplicable ItemState = "not_applicable"
	// ItemRetainedLegacy means the user explicitly accepted that this
	// artifact stays in the legacy/ module.
	ItemRetainedLegacy ItemState = "retained_legacy"
)

// SourceEvidence is the no-out-of-scope guard. Every applied patch must
// carry one of these proving the change is grounded in real source content.
// The orchestrator validates: file exists, lines exist, content hash matches.
type SourceEvidence struct {
	File        string `json:"file"`         // path relative to repo root
	Lines       string `json:"lines"`        // e.g., "12-58"
	ContentHash string `json:"content_hash"` // sha256 of the byte range
}

// OutputItem is a single unit of specialist output. A specialist's full
// output is a slice of these.
type OutputItem struct {
	ID             string          `json:"id"`
	State          ItemState       `json:"state"`
	Description    string          `json:"description"`
	SourceEvidence *SourceEvidence `json:"source_evidence,omitempty"`

	// FileWrites are the file-system changes that, taken together, make up
	// this item's patch. Specialists declare what files to create / replace
	// / delete; the orchestrator applies them after parse-validation. This
	// is more reliable than asking the LLM to produce well-formed unified
	// diffs (which it often gets wrong on context lines and line counts).
	FileWrites []FileWrite `json:"file_writes,omitempty"`

	// Patch is a free-form string payload the specialist can use for
	// out-of-band data delivery. The assessor uses it to embed the full
	// Assessment JSON; module specialists generally don't use it.
	Patch string `json:"patch,omitempty"`

	// BlockerCode is set when State is ItemBlocked.
	BlockerCode  BlockerCode `json:"blocker_code,omitempty"`
	BlockerNote  string      `json:"blocker_note,omitempty"`
	Alternatives []string    `json:"alternatives,omitempty"`

	// Question is set when State is ItemRequiresDecision.
	Question string   `json:"question,omitempty"`
	Choices  []string `json:"choices,omitempty"`

	// Reason is the human-readable rationale for ItemNotApplicable.
	Reason string `json:"reason,omitempty"`
}

// FileWrite is one file-system change. The orchestrator applies these
// after the specialist returns. Path is relative to repo root and must
// not traverse outside the repo (orchestrator enforces).
type FileWrite struct {
	Path      string        `json:"path"`
	Operation FileOperation `json:"operation"`
	Content   string        `json:"content,omitempty"` // present for create/replace
}

// FileOperation enumerates the allowed file-write actions.
type FileOperation string

const (
	OpCreate  FileOperation = "create"  // create a new file (fails if exists)
	OpReplace FileOperation = "replace" // replace an existing file's content
	OpDelete  FileOperation = "delete"  // delete an existing file
)

// BlockerCode is the fixed enum of reasons a specialist could not migrate
// an artifact. New codes require an explicit code change — the orchestrator
// rejects unknown codes (see §7 of plan).
type BlockerCode string

// Schema / data model
const (
	BlockerFKRequired                 BlockerCode = "FK_REQUIRED"
	BlockerOffsetPaginationIncompat   BlockerCode = "OFFSET_PAGINATION_INCOMPATIBLE"
	BlockerStatefulDTO                BlockerCode = "STATEFUL_DTO"
	BlockerCompositePKNoNaturalOrder  BlockerCode = "COMPOSITE_PK_NO_NATURAL_ORDER"
	BlockerMutableEntityGraph         BlockerCode = "MUTABLE_ENTITY_GRAPH"
	BlockerEmbeddedDBDialect          BlockerCode = "EMBEDDED_DB_DIALECT"
)

// Coupling / runtime
const (
	BlockerStaticGlobalState     BlockerCode = "STATIC_GLOBAL_STATE"
	BlockerAppContextLookup      BlockerCode = "APPCONTEXT_LOOKUP"
	BlockerServiceLoader         BlockerCode = "SERVICELOADER"
	BlockerFieldInjectionComplex BlockerCode = "FIELD_INJECTION_COMPLEX"
	BlockerThreadLocalLifecycle  BlockerCode = "THREADLOCAL_LIFECYCLE"
	BlockerNonVirtualThreadSafe  BlockerCode = "NON_VIRTUAL_THREAD_SAFE"
	BlockerBlockingReactiveMix   BlockerCode = "BLOCKING_REACTIVE_MIX"
)

// Build / framework
const (
	BlockerGradleParentAsArtifact   BlockerCode = "GRADLE_PARENT_AS_ARTIFACT"
	BlockerBuildPluginNotPortable   BlockerCode = "BUILD_PLUGIN_NOT_PORTABLE"
	BlockerJavaVersionIncompatible  BlockerCode = "JAVA_VERSION_INCOMPATIBLE"
	BlockerNonJakartaNoReplacement  BlockerCode = "NON_JAKARTA_DEP_NO_REPLACEMENT"
	BlockerNonSpringFramework       BlockerCode = "NON_SPRING_FRAMEWORK"
)

// Wire format / contract
const (
	BlockerLegacyErrorFormatRequired BlockerCode = "LEGACY_ERROR_FORMAT_REQUIRED"
	BlockerBespokeAuthProtocol       BlockerCode = "BESPOKE_AUTH_PROTOCOL"
	BlockerBinaryProtocol            BlockerCode = "BINARY_PROTOCOL"
)

// Tests
const (
	BlockerPowerMockLegacy           BlockerCode = "POWERMOCK_LEGACY"
	BlockerMissingCharacterization   BlockerCode = "MISSING_CHARACTERIZATION_BASIS"
	BlockerBroadTestSuiteSlow        BlockerCode = "BROAD_TEST_SUITE_SLOW"
	BlockerSpockTests                BlockerCode = "SPOCK_TESTS"
)

// Repo shape
const (
	BlockerNonJVMCodeSubstantial BlockerCode = "NON_JVM_CODE_SUBSTANTIAL"
	BlockerMultiLanguageBuild    BlockerCode = "MULTI_LANGUAGE_BUILD"
	BlockerKotlinPartial         BlockerCode = "KOTLIN_PARTIAL"
	BlockerSecretInSource        BlockerCode = "SECRET_IN_SOURCE"
)

// Deployment / CI-CD
const (
	BlockerDockerfileGranularityChange BlockerCode = "DOCKERFILE_GRANULARITY_CHANGE"
	BlockerDeploymentTopologyChange    BlockerCode = "DEPLOYMENT_TOPOLOGY_CHANGE"
	BlockerJavaVersionMismatchCI       BlockerCode = "JAVA_VERSION_MISMATCH_CI"
	BlockerExternalScriptReferenced    BlockerCode = "EXTERNAL_SCRIPT_REFERENCED"
	BlockerDeployTargetUnresolvable    BlockerCode = "DEPLOY_TARGET_UNRESOLVABLE"
)

// Local build environment (raised by orchestrator preflight, not by a
// specialist). Distinct from BlockerJavaVersionMismatchCI which is about
// the user's CI workflow declaring a wrong JDK.
const (
	BlockerJavaVersionMismatchRuntime BlockerCode = "JAVA_VERSION_MISMATCH_RUNTIME"
)

// Validation funnel failures (auto-fed back to specialist for retry)
const (
	BlockerCompileFailed   BlockerCode = "COMPILE_FAILED"
	BlockerArchUnitFailed  BlockerCode = "ARCHUNIT_VIOLATED"
	BlockerTestsRegressed  BlockerCode = "TESTS_REGRESSED"
	BlockerEvidenceInvalid BlockerCode = "EVIDENCE_INVALID"
)

// Activation-phase failures (only surfaced in Phase 12)
const (
	BlockerEnforcerViolation     BlockerCode = "ENFORCER_VIOLATION"
	BlockerSpotlessViolation     BlockerCode = "SPOTLESS_VIOLATION"
	BlockerCoverageBelowThresh   BlockerCode = "COVERAGE_BELOW_THRESHOLD"
)

// IsKnown reports whether the BlockerCode is in the fixed enum. Specialists
// that emit unknown codes have their output rejected by the orchestrator.
func (b BlockerCode) IsKnown() bool {
	_, ok := knownBlockerCodes[b]
	return ok
}

var knownBlockerCodes = map[BlockerCode]struct{}{
	BlockerFKRequired:                  {},
	BlockerOffsetPaginationIncompat:    {},
	BlockerStatefulDTO:                 {},
	BlockerCompositePKNoNaturalOrder:   {},
	BlockerMutableEntityGraph:          {},
	BlockerEmbeddedDBDialect:           {},
	BlockerStaticGlobalState:           {},
	BlockerAppContextLookup:            {},
	BlockerServiceLoader:               {},
	BlockerFieldInjectionComplex:       {},
	BlockerThreadLocalLifecycle:        {},
	BlockerNonVirtualThreadSafe:        {},
	BlockerBlockingReactiveMix:         {},
	BlockerGradleParentAsArtifact:      {},
	BlockerBuildPluginNotPortable:      {},
	BlockerJavaVersionIncompatible:     {},
	BlockerNonJakartaNoReplacement:     {},
	BlockerNonSpringFramework:          {},
	BlockerLegacyErrorFormatRequired:   {},
	BlockerBespokeAuthProtocol:         {},
	BlockerBinaryProtocol:              {},
	BlockerPowerMockLegacy:             {},
	BlockerMissingCharacterization:     {},
	BlockerBroadTestSuiteSlow:          {},
	BlockerSpockTests:                  {},
	BlockerNonJVMCodeSubstantial:       {},
	BlockerMultiLanguageBuild:          {},
	BlockerKotlinPartial:               {},
	BlockerSecretInSource:              {},
	BlockerDockerfileGranularityChange: {},
	BlockerDeploymentTopologyChange:    {},
	BlockerJavaVersionMismatchCI:       {},
	BlockerExternalScriptReferenced:    {},
	BlockerDeployTargetUnresolvable:    {},
	BlockerCompileFailed:               {},
	BlockerArchUnitFailed:              {},
	BlockerTestsRegressed:              {},
	BlockerEvidenceInvalid:             {},
	BlockerEnforcerViolation:           {},
	BlockerSpotlessViolation:           {},
	BlockerCoverageBelowThresh:         {},
	BlockerJavaVersionMismatchRuntime:  {},
}

// GateAction is the user's response at a phase approval gate.
type GateAction string

const (
	GateApprove        GateAction = "approve"
	GateEditAndApprove GateAction = "edit_and_approve"
	GateReject         GateAction = "reject"
)

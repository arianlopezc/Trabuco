package migrate

// Config holds configuration for the migration process
type Config struct {
	// SourcePath is the path to the source project to migrate
	SourcePath string

	// OutputPath is the path where the migrated project will be created
	OutputPath string

	// DryRun if true, analyzes without generating files
	DryRun bool

	// Interactive if true, prompts for confirmation on major decisions
	Interactive bool

	// Resume if true, resumes from last checkpoint
	Resume bool

	// IncludeTests if true, migrates test files
	IncludeTests bool

	// Verbose if true, shows detailed output
	Verbose bool

	// Debug if true, saves all AI interactions
	Debug bool

	// SkipBuild if true, skips Maven build after migration
	SkipBuild bool

	// TrabucoVersion is the version of Trabuco being used
	TrabucoVersion string

	// PreScannedProjectInfo holds pre-scanned project info to avoid duplicate scanning
	// If set, the migrator will skip the discovery scan phase
	PreScannedProjectInfo *ProjectInfo

	// PreAnalyzedDependencies holds pre-analyzed dependency report
	// If set, the migrator will skip dependency analysis
	PreAnalyzedDependencies *DependencyReport
}

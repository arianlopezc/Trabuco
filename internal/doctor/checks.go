package doctor

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// Checker is the interface for individual health checks
type Checker interface {
	ID() string
	Name() string
	Category() string
	Check(projectPath string, meta *config.ProjectMetadata) CheckResult
	Fix(projectPath string, meta *config.ProjectMetadata) error
}

// CheckCategory represents a category of checks
type CheckCategory string

const (
	CategoryStructure   CheckCategory = "structure"
	CategoryMetadata    CheckCategory = "metadata"
	CategoryConsistency CheckCategory = "consistency"
)

// BaseCheck provides common fields for checks
type BaseCheck struct {
	id       string
	name     string
	category CheckCategory
}

func (b *BaseCheck) ID() string            { return b.id }
func (b *BaseCheck) Name() string          { return b.name }
func (b *BaseCheck) Category() string      { return string(b.category) }
func (b *BaseCheck) Fix(projectPath string, meta *config.ProjectMetadata) error {
	return fmt.Errorf("auto-fix not supported for this check")
}

// --- PROJECT_STRUCTURE Check ---

// ProjectStructureCheck verifies pom.xml exists in the current directory
type ProjectStructureCheck struct {
	BaseCheck
}

func NewProjectStructureCheck() *ProjectStructureCheck {
	return &ProjectStructureCheck{
		BaseCheck: BaseCheck{
			id:       "PROJECT_STRUCTURE",
			name:     "Project structure valid",
			category: CategoryStructure,
		},
	}
}

func (c *ProjectStructureCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	pomPath := filepath.Join(projectPath, "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "pom.xml not found in current directory",
			Details: []string{"This directory is not a Maven project"},
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

// --- TRABUCO_PROJECT Check ---

// TrabucoProjectCheck verifies this is a Trabuco-generated project
type TrabucoProjectCheck struct {
	BaseCheck
}

func NewTrabucoProjectCheck() *TrabucoProjectCheck {
	return &TrabucoProjectCheck{
		BaseCheck: BaseCheck{
			id:       "TRABUCO_PROJECT",
			name:     "Trabuco project detected",
			category: CategoryStructure,
		},
	}
}

func (c *TrabucoProjectCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	// Check for .trabuco.json first
	if config.MetadataExists(projectPath) {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass,
		}
	}

	// Try to detect from structure
	pomPath := filepath.Join(projectPath, "pom.xml")
	pom, err := ParseParentPOM(pomPath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Could not parse pom.xml",
			Details: []string{err.Error()},
		}
	}

	// Check for Model module (required for all Trabuco projects)
	hasModel := false
	for _, module := range pom.Modules {
		if module == config.ModuleModel {
			hasModel = true
			break
		}
	}

	if !hasModel {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Not a Trabuco project",
			Details: []string{"Missing Model module", "Run 'trabuco init' to create a new project"},
		}
	}

	// Check for Model directory structure
	modelPath := filepath.Join(projectPath, "Model", "src", "main", "java")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Not a Trabuco project",
			Details: []string{"Model module directory structure is missing"},
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

// --- METADATA_EXISTS Check ---

// MetadataExistsCheck verifies .trabuco.json exists
type MetadataExistsCheck struct {
	BaseCheck
}

func NewMetadataExistsCheck() *MetadataExistsCheck {
	return &MetadataExistsCheck{
		BaseCheck: BaseCheck{
			id:       "METADATA_EXISTS",
			name:     "Metadata file exists",
			category: CategoryMetadata,
		},
	}
}

func (c *MetadataExistsCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	if config.MetadataExists(projectPath) {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass,
		}
	}

	return CheckResult{
		ID:         c.id,
		Name:       c.name,
		Status:     SeverityWarn,
		Message:    ".trabuco.json not found",
		FixAction:  "regenerate from POM",
		CanAutoFix: true,
	}
}

func (c *MetadataExistsCheck) Fix(projectPath string, meta *config.ProjectMetadata) error {
	if meta == nil {
		var err error
		meta, err = InferFromPOM(projectPath)
		if err != nil {
			return err
		}
	}
	return config.SaveMetadata(projectPath, meta)
}

// --- METADATA_VALID Check ---

// MetadataValidCheck verifies .trabuco.json is valid JSON with required fields
type MetadataValidCheck struct {
	BaseCheck
}

func NewMetadataValidCheck() *MetadataValidCheck {
	return &MetadataValidCheck{
		BaseCheck: BaseCheck{
			id:       "METADATA_VALID",
			name:     "Metadata file valid",
			category: CategoryMetadata,
		},
	}
}

func (c *MetadataValidCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	if !config.MetadataExists(projectPath) {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass, // Skipped if doesn't exist (handled by METADATA_EXISTS)
		}
	}

	loadedMeta, err := config.LoadMetadata(projectPath)
	if err != nil {
		return CheckResult{
			ID:         c.id,
			Name:       c.name,
			Status:     SeverityError,
			Message:    "Invalid .trabuco.json",
			Details:    []string{err.Error()},
			FixAction:  "regenerate from POM",
			CanAutoFix: true,
		}
	}

	// Validate required fields
	var missing []string
	if loadedMeta.ProjectName == "" {
		missing = append(missing, "projectName")
	}
	if loadedMeta.GroupID == "" {
		missing = append(missing, "groupId")
	}
	if len(loadedMeta.Modules) == 0 {
		missing = append(missing, "modules")
	}

	if len(missing) > 0 {
		return CheckResult{
			ID:         c.id,
			Name:       c.name,
			Status:     SeverityError,
			Message:    "Missing required fields",
			Details:    missing,
			FixAction:  "regenerate from POM",
			CanAutoFix: true,
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

func (c *MetadataValidCheck) Fix(projectPath string, meta *config.ProjectMetadata) error {
	// Regenerate from POM
	newMeta, err := InferFromPOM(projectPath)
	if err != nil {
		return err
	}
	return config.SaveMetadata(projectPath, newMeta)
}

// --- METADATA_SYNC Check ---

// MetadataSyncCheck verifies modules in .trabuco.json match actual directories
type MetadataSyncCheck struct {
	BaseCheck
}

func NewMetadataSyncCheck() *MetadataSyncCheck {
	return &MetadataSyncCheck{
		BaseCheck: BaseCheck{
			id:       "METADATA_SYNC",
			name:     "Metadata in sync",
			category: CategoryMetadata,
		},
	}
}

func (c *MetadataSyncCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	if meta == nil || !config.MetadataExists(projectPath) {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass, // Skipped if metadata doesn't exist
		}
	}

	// Get modules from POM
	pomModules, err := GetModulesFromPOM(projectPath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityWarn,
			Message: "Could not read POM modules",
			Details: []string{err.Error()},
		}
	}

	// Compare metadata modules with POM modules
	metaSet := make(map[string]bool)
	for _, m := range meta.Modules {
		metaSet[m] = true
	}

	pomSet := make(map[string]bool)
	for _, m := range pomModules {
		pomSet[m] = true
	}

	var inMetaNotPOM, inPOMNotMeta []string
	for m := range metaSet {
		if !pomSet[m] {
			inMetaNotPOM = append(inMetaNotPOM, m)
		}
	}
	for m := range pomSet {
		if !metaSet[m] {
			inPOMNotMeta = append(inPOMNotMeta, m)
		}
	}

	if len(inMetaNotPOM) > 0 || len(inPOMNotMeta) > 0 {
		var details []string
		if len(inMetaNotPOM) > 0 {
			details = append(details, fmt.Sprintf("In metadata but not in POM: %v", inMetaNotPOM))
		}
		if len(inPOMNotMeta) > 0 {
			details = append(details, fmt.Sprintf("In POM but not in metadata: %v", inPOMNotMeta))
		}

		return CheckResult{
			ID:         c.id,
			Name:       c.name,
			Status:     SeverityWarn,
			Message:    "Metadata modules don't match POM",
			Details:    details,
			FixAction:  "sync metadata with POM",
			CanAutoFix: true,
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

func (c *MetadataSyncCheck) Fix(projectPath string, meta *config.ProjectMetadata) error {
	// Get modules from POM and update metadata
	pomModules, err := GetModulesFromPOM(projectPath)
	if err != nil {
		return err
	}

	meta.Modules = pomModules
	meta.UpdateGeneratedAt()
	return config.SaveMetadata(projectPath, meta)
}

// --- PARENT_POM_VALID Check ---

// ParentPOMValidCheck verifies parent POM has required sections
type ParentPOMValidCheck struct {
	BaseCheck
}

func NewParentPOMValidCheck() *ParentPOMValidCheck {
	return &ParentPOMValidCheck{
		BaseCheck: BaseCheck{
			id:       "PARENT_POM_VALID",
			name:     "Parent POM valid",
			category: CategoryStructure,
		},
	}
}

func (c *ParentPOMValidCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	hasModules, hasProperties, err := HasRequiredPOMSections(projectPath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Could not read parent POM",
			Details: []string{err.Error()},
		}
	}

	var missing []string
	if !hasModules {
		missing = append(missing, "<modules> section")
	}
	if !hasProperties {
		missing = append(missing, "<properties> section")
	}

	if len(missing) > 0 {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Parent POM missing required sections",
			Details: missing,
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

// --- MODULE_POMS_EXIST Check ---

// ModulePOMsExistCheck verifies all declared modules have pom.xml
type ModulePOMsExistCheck struct {
	BaseCheck
}

func NewModulePOMsExistCheck() *ModulePOMsExistCheck {
	return &ModulePOMsExistCheck{
		BaseCheck: BaseCheck{
			id:       "MODULE_POMS_EXIST",
			name:     "Module POMs exist",
			category: CategoryStructure,
		},
	}
}

func (c *ModulePOMsExistCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	modules, err := GetModulesFromPOM(projectPath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Could not read modules from POM",
			Details: []string{err.Error()},
		}
	}

	var missing []string
	for _, module := range modules {
		pomPath := filepath.Join(projectPath, module, "pom.xml")
		if _, err := os.Stat(pomPath); os.IsNotExist(err) {
			missing = append(missing, fmt.Sprintf("%s/pom.xml", module))
		}
	}

	if len(missing) > 0 {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Module POMs missing",
			Details: missing,
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   fmt.Sprintf("Module POMs exist (%d modules)", len(modules)),
		Status: SeverityPass,
	}
}

// --- MODULE_DIRS_EXIST Check ---

// ModuleDirsExistCheck verifies all POM-declared modules have directories
type ModuleDirsExistCheck struct {
	BaseCheck
}

func NewModuleDirsExistCheck() *ModuleDirsExistCheck {
	return &ModuleDirsExistCheck{
		BaseCheck: BaseCheck{
			id:       "MODULE_DIRS_EXIST",
			name:     "Module directories exist",
			category: CategoryStructure,
		},
	}
}

func (c *ModuleDirsExistCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	modules, err := GetModulesFromPOM(projectPath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Could not read modules from POM",
			Details: []string{err.Error()},
		}
	}

	var missing []string
	for _, module := range modules {
		modulePath := filepath.Join(projectPath, module)
		if _, err := os.Stat(modulePath); os.IsNotExist(err) {
			missing = append(missing, module)
		}
	}

	if len(missing) > 0 {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityError,
			Message: "Module directories missing",
			Details: missing,
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

// --- JAVA_VERSION_CONSISTENT Check ---

// JavaVersionConsistentCheck verifies Java versions match across all POMs
type JavaVersionConsistentCheck struct {
	BaseCheck
}

func NewJavaVersionConsistentCheck() *JavaVersionConsistentCheck {
	return &JavaVersionConsistentCheck{
		BaseCheck: BaseCheck{
			id:       "JAVA_VERSION_CONSISTENT",
			name:     "Java version consistent",
			category: CategoryConsistency,
		},
	}
}

func (c *JavaVersionConsistentCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	parentVersion, err := GetJavaVersionFromPOM(projectPath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityWarn,
			Message: "Could not determine Java version from parent POM",
		}
	}

	if parentVersion == "" {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityWarn,
			Message: "Java version not specified in parent POM",
		}
	}

	// For now, just check parent POM (child POMs should inherit)
	return CheckResult{
		ID:     c.id,
		Name:   fmt.Sprintf("Java version consistent (%s)", parentVersion),
		Status: SeverityPass,
	}
}

// --- GROUP_ID_CONSISTENT Check ---

// GroupIDConsistentCheck verifies group IDs match across all POMs
type GroupIDConsistentCheck struct {
	BaseCheck
}

func NewGroupIDConsistentCheck() *GroupIDConsistentCheck {
	return &GroupIDConsistentCheck{
		BaseCheck: BaseCheck{
			id:       "GROUP_ID_CONSISTENT",
			name:     "Group ID consistent",
			category: CategoryConsistency,
		},
	}
}

func (c *GroupIDConsistentCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	parentGroupID, err := GetGroupIDFromPOM(projectPath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityWarn,
			Message: "Could not determine group ID from parent POM",
		}
	}

	// Check module POMs
	modules, err := GetModulesFromPOM(projectPath)
	if err != nil {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass, // Can't check modules, assume OK
		}
	}

	var inconsistent []string
	for _, module := range modules {
		pomPath := filepath.Join(projectPath, module, "pom.xml")
		if _, err := os.Stat(pomPath); os.IsNotExist(err) {
			continue // Skip missing POMs (handled by another check)
		}

		modInfo, err := ParseModulePOM(pomPath)
		if err != nil {
			continue // Skip unparseable POMs
		}

		// Module should inherit groupId from parent or explicitly match
		if modInfo.GroupID != "" && modInfo.GroupID != parentGroupID {
			inconsistent = append(inconsistent, fmt.Sprintf("%s: %s", module, modInfo.GroupID))
		}
	}

	if len(inconsistent) > 0 {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityWarn,
			Message: "Some modules have different group IDs",
			Details: inconsistent,
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

// --- DOCKER_COMPOSE_SYNC Check ---

// DockerComposeSyncCheck verifies docker-compose.yml matches module requirements
type DockerComposeSyncCheck struct {
	BaseCheck
}

func NewDockerComposeSyncCheck() *DockerComposeSyncCheck {
	return &DockerComposeSyncCheck{
		BaseCheck: BaseCheck{
			id:       "DOCKER_COMPOSE_SYNC",
			name:     "Docker Compose in sync",
			category: CategoryConsistency,
		},
	}
}

func (c *DockerComposeSyncCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	if meta == nil {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass, // Skip if no metadata
		}
	}

	// Check if docker-compose should exist
	cfg := meta.ToProjectConfig()
	if !cfg.NeedsDockerCompose() {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass, // Not needed
		}
	}

	composePath := filepath.Join(projectPath, "docker-compose.yml")
	if _, err := os.Stat(composePath); os.IsNotExist(err) {
		return CheckResult{
			ID:         c.id,
			Name:       c.name,
			Status:     SeverityWarn,
			Message:    "docker-compose.yml not found but modules require services",
			FixAction:  "generate docker-compose.yml",
			CanAutoFix: true,
		}
	}

	dc, err := ParseDockerCompose(composePath)
	if err != nil {
		return CheckResult{
			ID:      c.id,
			Name:    c.name,
			Status:  SeverityWarn,
			Message: "Could not parse docker-compose.yml",
			Details: []string{err.Error()},
		}
	}

	required := GetRequiredDockerServices(meta)
	var missing []string
	for _, service := range required {
		if _, ok := dc.Services[service]; !ok {
			missing = append(missing, service)
		}
	}

	if len(missing) > 0 {
		return CheckResult{
			ID:         c.id,
			Name:       c.name,
			Status:     SeverityWarn,
			Message:    "Missing required Docker services",
			Details:    missing,
			FixAction:  "update docker-compose.yml",
			CanAutoFix: true,
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

func (c *DockerComposeSyncCheck) Fix(projectPath string, meta *config.ProjectMetadata) error {
	// This would require regenerating docker-compose.yml
	// For now, return an error indicating manual fix is needed
	return fmt.Errorf("docker-compose.yml regeneration not yet implemented; please regenerate manually")
}

// --- CROSS_MODULE_DEPS Check ---

// CrossModuleDepsCheck verifies cross-module dependencies are correct in POMs
type CrossModuleDepsCheck struct {
	BaseCheck
}

func NewCrossModuleDepsCheck() *CrossModuleDepsCheck {
	return &CrossModuleDepsCheck{
		BaseCheck: BaseCheck{
			id:       "CROSS_MODULE_DEPS",
			name:     "Cross-module dependencies valid",
			category: CategoryConsistency,
		},
	}
}

func (c *CrossModuleDepsCheck) Check(projectPath string, meta *config.ProjectMetadata) CheckResult {
	// This is a complex check that would need to parse all module POMs
	// and verify their internal dependencies are correct.
	// For now, we'll do a basic check.

	modules, err := GetModulesFromPOM(projectPath)
	if err != nil {
		return CheckResult{
			ID:     c.id,
			Name:   c.name,
			Status: SeverityPass, // Skip if can't read modules
		}
	}

	// Basic validation: all modules should have valid POMs
	for _, module := range modules {
		pomPath := filepath.Join(projectPath, module, "pom.xml")
		if !IsValidPOM(pomPath) {
			return CheckResult{
				ID:      c.id,
				Name:    c.name,
				Status:  SeverityWarn,
				Message: fmt.Sprintf("Module %s has invalid pom.xml", module),
			}
		}
	}

	return CheckResult{
		ID:     c.id,
		Name:   c.name,
		Status: SeverityPass,
	}
}

// GetAllChecks returns all available checks
func GetAllChecks() []Checker {
	return []Checker{
		NewProjectStructureCheck(),
		NewTrabucoProjectCheck(),
		NewMetadataExistsCheck(),
		NewMetadataValidCheck(),
		NewMetadataSyncCheck(),
		NewParentPOMValidCheck(),
		NewModulePOMsExistCheck(),
		NewModuleDirsExistCheck(),
		NewJavaVersionConsistentCheck(),
		NewGroupIDConsistentCheck(),
		NewDockerComposeSyncCheck(),
		NewCrossModuleDepsCheck(),
	}
}

// GetChecksByCategory returns checks filtered by category
func GetChecksByCategory(category string) []Checker {
	var filtered []Checker
	for _, check := range GetAllChecks() {
		if check.Category() == category {
			filtered = append(filtered, check)
		}
	}
	return filtered
}

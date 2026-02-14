package generator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// BackupManager handles backup and restore of files before modification
type BackupManager struct {
	projectPath     string
	backupDir       string
	timestamp       string
	files           []string
	createdDirs     []string // Track directories created during add operation
	enabled         bool
}

// BackupDirName is the name of the backup directory
const BackupDirName = ".trabuco-backup"

// NewBackupManager creates a new BackupManager
func NewBackupManager(projectPath string, enabled bool) *BackupManager {
	timestamp := time.Now().Format("20060102-150405")
	backupDir := filepath.Join(projectPath, BackupDirName, timestamp)

	return &BackupManager{
		projectPath: projectPath,
		backupDir:   backupDir,
		timestamp:   timestamp,
		files:       []string{},
		createdDirs: []string{},
		enabled:     enabled,
	}
}

// TrackCreatedDir records a directory that was created during the add operation
// so it can be removed during rollback
func (b *BackupManager) TrackCreatedDir(dir string) {
	// Only track if enabled and if the directory is within the project
	if !b.enabled {
		return
	}
	b.createdDirs = append(b.createdDirs, dir)
}

// Backup creates a backup of a single file
func (b *BackupManager) Backup(relativePath string) error {
	if !b.enabled {
		return nil
	}

	srcPath := filepath.Join(b.projectPath, relativePath)

	// Check if source file exists
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return nil // Nothing to backup
	}

	// Create backup directory if needed
	if err := os.MkdirAll(b.backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create destination path
	dstPath := filepath.Join(b.backupDir, relativePath)

	// Create parent directories in backup
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create backup subdirectory: %w", err)
	}

	// Copy file
	if err := copyFile(srcPath, dstPath); err != nil {
		return fmt.Errorf("failed to backup %s: %w", relativePath, err)
	}

	b.files = append(b.files, relativePath)
	return nil
}

// BackupAll creates backups of multiple files
func (b *BackupManager) BackupAll(relativePaths []string) error {
	for _, path := range relativePaths {
		if err := b.Backup(path); err != nil {
			return err
		}
	}
	return nil
}

// Restore restores all backed up files and removes created directories
func (b *BackupManager) Restore() error {
	if !b.enabled {
		return nil
	}

	var restoreErrors []error

	// First, restore backed up files
	for _, relativePath := range b.files {
		srcPath := filepath.Join(b.backupDir, relativePath)
		dstPath := filepath.Join(b.projectPath, relativePath)

		if err := copyFile(srcPath, dstPath); err != nil {
			restoreErrors = append(restoreErrors, fmt.Errorf("failed to restore %s: %w", relativePath, err))
		}
	}

	// Then, remove created directories in reverse order (deepest first)
	// This ensures parent directories are removed after their children
	for i := len(b.createdDirs) - 1; i >= 0; i-- {
		dir := b.createdDirs[i]
		// Only remove if the directory exists and is empty or was fully created by us
		if err := os.RemoveAll(dir); err != nil {
			restoreErrors = append(restoreErrors, fmt.Errorf("failed to remove created directory %s: %w", dir, err))
		}
	}

	if len(restoreErrors) > 0 {
		// Combine all errors for better debugging
		errMsgs := make([]string, len(restoreErrors))
		for i, err := range restoreErrors {
			errMsgs[i] = err.Error()
		}
		return fmt.Errorf("restore failed with %d errors: %s", len(restoreErrors), strings.Join(errMsgs, "; "))
	}

	return nil
}

// Cleanup removes the backup directory and the parent if empty
func (b *BackupManager) Cleanup() error {
	if !b.enabled {
		return nil
	}

	// Remove the timestamped backup directory
	if err := os.RemoveAll(b.backupDir); err != nil {
		return err
	}

	// Try to remove the parent .trabuco-backup directory if empty
	backupRoot := filepath.Join(b.projectPath, BackupDirName)
	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		// Directory might not exist, that's fine
		return nil
	}

	if len(entries) == 0 {
		// Directory is empty, remove it
		return os.Remove(backupRoot)
	}

	return nil
}

// CleanupOldBackups removes all but the most recent backup
func (b *BackupManager) CleanupOldBackups() error {
	backupRoot := filepath.Join(b.projectPath, BackupDirName)

	entries, err := os.ReadDir(backupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// Keep only the most recent backup (current one)
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != b.timestamp {
			oldBackupPath := filepath.Join(backupRoot, entry.Name())
			if err := os.RemoveAll(oldBackupPath); err != nil {
				return fmt.Errorf("failed to remove old backup %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// GetBackupPath returns the path to the backup directory
func (b *BackupManager) GetBackupPath() string {
	return b.backupDir
}

// GetBackedUpFiles returns the list of backed up files
func (b *BackupManager) GetBackedUpFiles() []string {
	return b.files
}

// HasBackups returns true if any files were backed up
func (b *BackupManager) HasBackups() bool {
	return len(b.files) > 0
}

// PrintRestoreInstructions prints instructions for manual restore
func (b *BackupManager) PrintRestoreInstructions() {
	if !b.enabled || !b.HasBackups() {
		return
	}

	fmt.Println("\nTo restore from backup:")
	fmt.Printf("  cp -r %s/* %s/\n", b.backupDir, b.projectPath)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Preserve file permissions
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, srcInfo.Mode())
}

// GetFilesToBackup returns a list of files that should be backed up before adding a module
func GetFilesToBackup(module string) []string {
	files := []string{
		"pom.xml",
		".trabuco.json",
		// Documentation files that will be regenerated
		"README.md",
		// AI agent context files (backup handles nonexistent files gracefully)
		"CLAUDE.md",
		".cursorrules",
		".github/copilot-instructions.md",
		".windsurf/rules/project.md",
		".clinerules/project.md",
	}

	// Docker-related files
	if needsDockerComposeUpdate(module) {
		files = append(files, "docker-compose.yml", ".env.example")
	}

	// Model module files that might be updated
	if module == config.ModuleSQLDatastore || module == config.ModuleNoSQLDatastore || module == config.ModuleWorker || module == config.ModuleEventConsumer {
		files = append(files, config.ModuleModel+"/pom.xml")
	}

	return files
}

// needsDockerComposeUpdate returns true if adding this module might need docker-compose updates
func needsDockerComposeUpdate(module string) bool {
	switch module {
	case config.ModuleSQLDatastore, config.ModuleNoSQLDatastore, config.ModuleWorker, config.ModuleEventConsumer:
		return true
	default:
		return false
	}
}

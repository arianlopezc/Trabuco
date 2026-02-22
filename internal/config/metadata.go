package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MetadataFileName is the name of the Trabuco metadata file
const MetadataFileName = ".trabuco.json"

// ProjectMetadata holds metadata about a Trabuco-generated project
type ProjectMetadata struct {
	Version       string   `json:"version"`
	GeneratedAt   string   `json:"generatedAt"`
	ProjectName   string   `json:"projectName"`
	GroupID       string   `json:"groupId"`
	ArtifactID    string   `json:"artifactId"`
	JavaVersion   string   `json:"javaVersion"`
	Modules       []string `json:"modules"`
	Database      string   `json:"database,omitempty"`
	NoSQLDatabase string   `json:"noSqlDatabase,omitempty"`
	MessageBroker string   `json:"messageBroker,omitempty"`
	AIAgents      []string `json:"aiAgents,omitempty"`
	CIProvider    string   `json:"ciProvider,omitempty"`
}

// LoadMetadata loads project metadata from .trabuco.json in the specified directory
func LoadMetadata(projectPath string) (*ProjectMetadata, error) {
	metadataPath := filepath.Join(projectPath, MetadataFileName)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("metadata file not found: %s", metadataPath)
		}
		return nil, fmt.Errorf("failed to read metadata file: %w", err)
	}

	var metadata ProjectMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata file: %w", err)
	}

	return &metadata, nil
}

// SaveMetadata saves project metadata to .trabuco.json in the specified directory
func SaveMetadata(projectPath string, meta *ProjectMetadata) error {
	metadataPath := filepath.Join(projectPath, MetadataFileName)

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}

	return nil
}

// MetadataExists checks if .trabuco.json exists in the specified directory
func MetadataExists(projectPath string) bool {
	metadataPath := filepath.Join(projectPath, MetadataFileName)
	_, err := os.Stat(metadataPath)
	return err == nil
}

// NewMetadataFromConfig creates a ProjectMetadata from a ProjectConfig
func NewMetadataFromConfig(cfg *ProjectConfig, version string) *ProjectMetadata {
	return &ProjectMetadata{
		Version:       version,
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		ProjectName:   cfg.ProjectName,
		GroupID:       cfg.GroupID,
		ArtifactID:    cfg.ArtifactID,
		JavaVersion:   cfg.JavaVersion,
		Modules:       cfg.Modules,
		Database:      cfg.Database,
		NoSQLDatabase: cfg.NoSQLDatabase,
		MessageBroker: cfg.MessageBroker,
		AIAgents:      cfg.AIAgents,
		CIProvider:    cfg.CIProvider,
	}
}

// ToProjectConfig converts ProjectMetadata to a ProjectConfig
// This is useful when inferring configuration from an existing project
func (m *ProjectMetadata) ToProjectConfig() *ProjectConfig {
	return &ProjectConfig{
		ProjectName:   m.ProjectName,
		GroupID:       m.GroupID,
		ArtifactID:    m.ArtifactID,
		JavaVersion:   m.JavaVersion,
		Modules:       m.Modules,
		Database:      m.Database,
		NoSQLDatabase: m.NoSQLDatabase,
		MessageBroker: m.MessageBroker,
		AIAgents:      m.AIAgents,
		CIProvider:    m.CIProvider,
	}
}

// HasModule checks if a specific module is in the metadata
func (m *ProjectMetadata) HasModule(name string) bool {
	for _, mod := range m.Modules {
		if mod == name {
			return true
		}
	}
	return false
}

// AddModule adds a module to the metadata if not already present
func (m *ProjectMetadata) AddModule(name string) {
	if !m.HasModule(name) {
		m.Modules = append(m.Modules, name)
	}
}

// UpdateGeneratedAt updates the generatedAt timestamp to now
func (m *ProjectMetadata) UpdateGeneratedAt() {
	m.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
}

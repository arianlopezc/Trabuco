package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// CacheVersion is incremented when cache format changes
	CacheVersion = "1"

	// DefaultCacheDir is the default cache directory
	DefaultCacheDir = ".trabuco/cache"
)

// Cache provides caching for migration results
type Cache struct {
	baseDir string
	enabled bool
}

// CacheEntry represents a cached item with metadata
type CacheEntry struct {
	Version   string    `json:"version"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Checksum  string    `json:"checksum"`
	Data      []byte    `json:"data"`
}

// NewCache creates a new cache instance
func NewCache(baseDir string) *Cache {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		baseDir = filepath.Join(home, DefaultCacheDir)
	}

	return &Cache{
		baseDir: baseDir,
		enabled: true,
	}
}

// SetEnabled enables or disables the cache
func (c *Cache) SetEnabled(enabled bool) {
	c.enabled = enabled
}

// IsEnabled returns whether caching is enabled
func (c *Cache) IsEnabled() bool {
	return c.enabled
}

// Get retrieves a cached entry
func (c *Cache) Get(category, key string) ([]byte, bool) {
	if !c.enabled {
		return nil, false
	}

	path := c.entryPath(category, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Check version
	if entry.Version != CacheVersion {
		return nil, false
	}

	// Check expiration
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		c.Delete(category, key)
		return nil, false
	}

	return entry.Data, true
}

// GetJSON retrieves and unmarshals a cached JSON entry
func (c *Cache) GetJSON(category, key string, v interface{}) bool {
	data, ok := c.Get(category, key)
	if !ok {
		return false
	}

	if err := json.Unmarshal(data, v); err != nil {
		return false
	}

	return true
}

// Set stores an entry in the cache
func (c *Cache) Set(category, key string, data []byte, ttl time.Duration) error {
	if !c.enabled {
		return nil
	}

	entry := CacheEntry{
		Version:   CacheVersion,
		Key:       key,
		CreatedAt: time.Now(),
		Checksum:  checksum(data),
		Data:      data,
	}

	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	entryData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal cache entry: %w", err)
	}

	path := c.entryPath(category, key)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	if err := os.WriteFile(path, entryData, 0644); err != nil {
		return fmt.Errorf("write cache entry: %w", err)
	}

	return nil
}

// SetJSON marshals and stores a JSON entry
func (c *Cache) SetJSON(category, key string, v interface{}, ttl time.Duration) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	return c.Set(category, key, data, ttl)
}

// Delete removes a cached entry
func (c *Cache) Delete(category, key string) error {
	path := c.entryPath(category, key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Clear removes all entries in a category
func (c *Cache) Clear(category string) error {
	dir := filepath.Join(c.baseDir, category)
	if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ClearAll removes all cached entries
func (c *Cache) ClearAll() error {
	if err := os.RemoveAll(c.baseDir); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// HasValidEntry checks if a valid cache entry exists
func (c *Cache) HasValidEntry(category, key string) bool {
	_, ok := c.Get(category, key)
	return ok
}

// GetStats returns cache statistics
func (c *Cache) GetStats() (*CacheStats, error) {
	stats := &CacheStats{
		Categories: make(map[string]CategoryStats),
	}

	err := filepath.Walk(c.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Ignore errors
		}
		if info.IsDir() {
			return nil
		}

		// Get category from path
		rel, _ := filepath.Rel(c.baseDir, path)
		parts := filepath.SplitList(rel)
		category := "unknown"
		if len(parts) > 0 {
			category = filepath.Dir(rel)
		}

		// Update stats
		catStats := stats.Categories[category]
		catStats.EntryCount++
		catStats.TotalSize += info.Size()
		stats.Categories[category] = catStats

		stats.TotalEntries++
		stats.TotalSize += info.Size()

		return nil
	})

	return stats, err
}

// CacheStats contains cache statistics
type CacheStats struct {
	TotalEntries int64
	TotalSize    int64
	Categories   map[string]CategoryStats
}

// CategoryStats contains stats for a cache category
type CategoryStats struct {
	EntryCount int64
	TotalSize  int64
}

// entryPath returns the file path for a cache entry
func (c *Cache) entryPath(category, key string) string {
	// Hash the key to avoid filesystem issues
	hash := hashKey(key)
	return filepath.Join(c.baseDir, category, hash+".json")
}

// hashKey creates a filesystem-safe hash of a key
func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:16]) // Use first 16 bytes
}

// checksum calculates a checksum for data
func checksum(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// Categories for different types of cached data
const (
	CategoryAnalysis    = "analysis"    // Project analysis results
	CategoryConversion  = "conversion"  // Converted file results
	CategoryDependency  = "dependency"  // Dependency analysis
	CategoryPattern     = "pattern"     // Migration patterns
)

// ProjectAnalysisKey creates a cache key for project analysis
func ProjectAnalysisKey(projectPath string, configHash string) string {
	return fmt.Sprintf("%s:%s", filepath.Base(projectPath), configHash)
}

// FileConversionKey creates a cache key for file conversion
func FileConversionKey(filePath string, contentHash string) string {
	return fmt.Sprintf("%s:%s", filepath.Base(filePath), contentHash)
}

// ContentHash creates a hash of file content
func ContentHash(content []byte) string {
	h := sha256.Sum256(content)
	return hex.EncodeToString(h[:])
}

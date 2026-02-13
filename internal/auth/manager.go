package auth

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Manager handles credential operations
type Manager struct {
	storage Storage
	store   *CredentialStore
}

// NewManager creates a new credential manager
func NewManager() (*Manager, error) {
	storage := GetPreferredStorage()
	store, err := storage.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	return &Manager{
		storage: storage,
		store:   store,
	}, nil
}

// NewManagerWithStorage creates a manager with a specific storage backend
func NewManagerWithStorage(storage Storage) (*Manager, error) {
	store, err := storage.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	return &Manager{
		storage: storage,
		store:   store,
	}, nil
}

// StorageBackend returns the name of the storage backend being used
func (m *Manager) StorageBackend() string {
	return m.storage.Name()
}

// SetCredential stores a credential and optionally validates it
func (m *Manager) SetCredential(cred *Credential, validate bool) error {
	// Basic validation
	if err := ValidateAPIKey(cred.Provider, cred.APIKey); err != nil {
		return err
	}

	// Store the credential
	m.store.SetCredential(cred)

	// Save to storage
	if err := m.storage.Save(m.store); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	return nil
}

// GetCredential retrieves a credential for a provider
func (m *Manager) GetCredential(provider Provider) (*Credential, error) {
	cred, ok := m.store.GetCredential(provider)
	if !ok {
		return nil, ErrProviderNotFound
	}
	return cred, nil
}

// GetDefaultCredential returns the default provider's credential
func (m *Manager) GetDefaultCredential() (*Credential, error) {
	cred, ok := m.store.GetDefaultCredential()
	if !ok {
		return nil, ErrNoCredentials
	}
	return cred, nil
}

// GetCredentialWithFallback returns credentials in order of preference:
// 1. Environment variable for the specified provider
// 2. Stored credential for the specified provider
// 3. Default stored credential
// 4. Any environment variable
func (m *Manager) GetCredentialWithFallback(preferredProvider Provider) (*Credential, error) {
	// 1. Check environment variable for preferred provider
	if preferredProvider != "" {
		if info, ok := SupportedProviders[preferredProvider]; ok && info.EnvVar != "" {
			if key := os.Getenv(info.EnvVar); key != "" {
				return &Credential{
					Provider: preferredProvider,
					APIKey:   key,
				}, nil
			}
		}
	}

	// 2. Check stored credential for preferred provider
	if preferredProvider != "" {
		if cred, err := m.GetCredential(preferredProvider); err == nil {
			return cred, nil
		}
	}

	// 3. Check default stored credential
	if cred, err := m.GetDefaultCredential(); err == nil {
		return cred, nil
	}

	// 4. Check any environment variable
	for provider, info := range SupportedProviders {
		if info.EnvVar != "" {
			if key := os.Getenv(info.EnvVar); key != "" {
				return &Credential{
					Provider: provider,
					APIKey:   key,
				}, nil
			}
		}
	}

	return nil, ErrNoCredentials
}

// RemoveCredential removes a credential
func (m *Manager) RemoveCredential(provider Provider) error {
	m.store.RemoveCredential(provider)
	return m.storage.Save(m.store)
}

// SetDefault sets the default provider
func (m *Manager) SetDefault(provider Provider) error {
	if err := m.store.SetDefault(provider); err != nil {
		return err
	}
	return m.storage.Save(m.store)
}

// GetDefault returns the default provider
func (m *Manager) GetDefault() Provider {
	return m.store.DefaultProvider
}

// ListConfigured returns all configured providers with their status
func (m *Manager) ListConfigured() []ProviderStatus {
	var statuses []ProviderStatus

	for provider, info := range SupportedProviders {
		status := ProviderStatus{
			Provider:  provider,
			Info:      info,
			IsDefault: provider == m.store.DefaultProvider,
		}

		// Check stored credential
		if cred, ok := m.store.GetCredential(provider); ok {
			status.Configured = true
			status.ValidatedAt = cred.ValidatedAt
			status.Source = "stored"
			if cred.Model != "" {
				status.Model = cred.Model
			}
		}

		// Check environment variable (overrides stored)
		if info.EnvVar != "" {
			if key := os.Getenv(info.EnvVar); key != "" {
				status.Configured = true
				status.Source = "env:" + info.EnvVar
			}
		}

		// Only include providers that require keys or are configured
		if info.RequiresKey || status.Configured {
			statuses = append(statuses, status)
		}
	}

	return statuses
}

// Clear removes all stored credentials
func (m *Manager) Clear() error {
	m.store = NewCredentialStore()
	return m.storage.Clear()
}

// MarkValidated marks a credential as validated
func (m *Manager) MarkValidated(provider Provider) error {
	cred, ok := m.store.GetCredential(provider)
	if !ok {
		return ErrProviderNotFound
	}
	cred.ValidatedAt = time.Now()
	return m.storage.Save(m.store)
}

// HasAnyCredentials returns true if any credentials are configured
func (m *Manager) HasAnyCredentials() bool {
	// Check stored credentials
	if len(m.store.Credentials) > 0 {
		return true
	}

	// Check environment variables
	for _, info := range SupportedProviders {
		if info.EnvVar != "" && os.Getenv(info.EnvVar) != "" {
			return true
		}
	}

	return false
}

// ProviderStatus represents the status of a provider's credentials
type ProviderStatus struct {
	Provider    Provider
	Info        ProviderInfo
	Configured  bool
	IsDefault   bool
	Source      string    // "stored", "env:VAR_NAME"
	ValidatedAt time.Time
	Model       string
}

// Validator defines an interface for validating credentials with an API
type Validator interface {
	ValidateCredential(ctx context.Context, cred *Credential) error
}

// ValidateWithAPI validates a credential by making a test API call
func (m *Manager) ValidateWithAPI(ctx context.Context, provider Provider, validator Validator) error {
	cred, err := m.GetCredential(provider)
	if err != nil {
		return err
	}

	if err := validator.ValidateCredential(ctx, cred); err != nil {
		return fmt.Errorf("API validation failed: %w", err)
	}

	return m.MarkValidated(provider)
}

package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	serviceName = "trabuco"
	accountName = "credentials"
)

// Storage defines the interface for credential storage backends
type Storage interface {
	// Load retrieves the credential store
	Load() (*CredentialStore, error)

	// Save persists the credential store
	Save(store *CredentialStore) error

	// Clear removes all stored credentials
	Clear() error

	// Name returns the storage backend name
	Name() string
}

// KeychainStorage uses the system keychain for credential storage
type KeychainStorage struct{}

// FileStorage uses an encrypted file for credential storage
type FileStorage struct {
	path string
}

// NewKeychainStorage creates a new keychain storage backend
func NewKeychainStorage() *KeychainStorage {
	return &KeychainStorage{}
}

// NewFileStorage creates a new file storage backend
func NewFileStorage(path string) *FileStorage {
	if path == "" {
		path = defaultCredentialPath()
	}
	return &FileStorage{path: path}
}

func defaultCredentialPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".trabuco", "credentials.enc")
}

// Name returns the storage backend name
func (k *KeychainStorage) Name() string {
	return "system keychain"
}

// Load retrieves credentials from the system keychain
func (k *KeychainStorage) Load() (*CredentialStore, error) {
	data, err := keychainGet(serviceName, accountName)
	if err != nil {
		if errors.Is(err, errKeychainItemNotFound) {
			return NewCredentialStore(), nil
		}
		return nil, fmt.Errorf("keychain load: %w", err)
	}

	var store CredentialStore
	if err := json.Unmarshal([]byte(data), &store); err != nil {
		return nil, fmt.Errorf("keychain unmarshal: %w", err)
	}

	return &store, nil
}

// Save persists credentials to the system keychain
func (k *KeychainStorage) Save(store *CredentialStore) error {
	data, err := json.Marshal(store)
	if err != nil {
		return fmt.Errorf("keychain marshal: %w", err)
	}

	if err := keychainSet(serviceName, accountName, string(data)); err != nil {
		return fmt.Errorf("keychain save: %w", err)
	}

	return nil
}

// Clear removes credentials from the system keychain
func (k *KeychainStorage) Clear() error {
	if err := keychainDelete(serviceName, accountName); err != nil {
		if errors.Is(err, errKeychainItemNotFound) {
			return nil
		}
		return fmt.Errorf("keychain clear: %w", err)
	}
	return nil
}

// Name returns the storage backend name
func (f *FileStorage) Name() string {
	return "encrypted file"
}

// Load retrieves credentials from the encrypted file
func (f *FileStorage) Load() (*CredentialStore, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewCredentialStore(), nil
		}
		return nil, fmt.Errorf("file load: %w", err)
	}

	decrypted, err := decrypt(data, getMachineID())
	if err != nil {
		return nil, fmt.Errorf("file decrypt: %w", err)
	}

	var store CredentialStore
	if err := json.Unmarshal(decrypted, &store); err != nil {
		return nil, fmt.Errorf("file unmarshal: %w", err)
	}

	return &store, nil
}

// Save persists credentials to the encrypted file
func (f *FileStorage) Save(store *CredentialStore) error {
	data, err := json.Marshal(store)
	if err != nil {
		return fmt.Errorf("file marshal: %w", err)
	}

	encrypted, err := encrypt(data, getMachineID())
	if err != nil {
		return fmt.Errorf("file encrypt: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(f.path), 0700); err != nil {
		return fmt.Errorf("file mkdir: %w", err)
	}

	if err := os.WriteFile(f.path, encrypted, 0600); err != nil {
		return fmt.Errorf("file write: %w", err)
	}

	return nil
}

// Clear removes the encrypted file
func (f *FileStorage) Clear() error {
	if err := os.Remove(f.path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("file clear: %w", err)
	}
	return nil
}

// Encryption helpers using AES-GCM

func encrypt(plaintext []byte, key string) ([]byte, error) {
	keyHash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return []byte(base64.StdEncoding.EncodeToString(ciphertext)), nil
}

func decrypt(ciphertext []byte, key string) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(string(ciphertext))
	if err != nil {
		return nil, err
	}

	keyHash := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyHash[:])
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertextBytes, nil)
}

// getMachineID returns a machine-specific identifier for encryption
func getMachineID() string {
	// Try to get a stable machine identifier
	switch runtime.GOOS {
	case "darwin":
		if id, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output(); err == nil {
			// Extract UUID from output
			lines := strings.Split(string(id), "\n")
			for _, line := range lines {
				if strings.Contains(line, "IOPlatformUUID") {
					parts := strings.Split(line, "=")
					if len(parts) == 2 {
						return strings.Trim(strings.TrimSpace(parts[1]), "\"")
					}
				}
			}
		}
	case "linux":
		if id, err := os.ReadFile("/etc/machine-id"); err == nil {
			return strings.TrimSpace(string(id))
		}
		if id, err := os.ReadFile("/var/lib/dbus/machine-id"); err == nil {
			return strings.TrimSpace(string(id))
		}
	case "windows":
		if id, err := exec.Command("wmic", "csproduct", "get", "UUID").Output(); err == nil {
			lines := strings.Split(string(id), "\n")
			if len(lines) > 1 {
				return strings.TrimSpace(lines[1])
			}
		}
	}

	// Fallback to hostname + username
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	return hostname + "-" + username
}

// Platform-specific keychain operations

var errKeychainItemNotFound = errors.New("keychain item not found")

func keychainGet(service, account string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return macKeychainGet(service, account)
	case "linux":
		return linuxSecretGet(service, account)
	case "windows":
		return windowsCredentialGet(service, account)
	default:
		return "", errors.New("keychain not supported on this platform")
	}
}

func keychainSet(service, account, password string) error {
	switch runtime.GOOS {
	case "darwin":
		return macKeychainSet(service, account, password)
	case "linux":
		return linuxSecretSet(service, account, password)
	case "windows":
		return windowsCredentialSet(service, account, password)
	default:
		return errors.New("keychain not supported on this platform")
	}
}

func keychainDelete(service, account string) error {
	switch runtime.GOOS {
	case "darwin":
		return macKeychainDelete(service, account)
	case "linux":
		return linuxSecretDelete(service, account)
	case "windows":
		return windowsCredentialDelete(service, account)
	default:
		return errors.New("keychain not supported on this platform")
	}
}

// macOS Keychain using security command
func macKeychainGet(service, account string) (string, error) {
	cmd := exec.Command("security", "find-generic-password", "-s", service, "-a", account, "-w")
	out, err := cmd.Output()
	if err != nil {
		if strings.Contains(err.Error(), "could not be found") || strings.Contains(string(out), "could not be found") {
			return "", errKeychainItemNotFound
		}
		// Check exit code for "item not found"
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 44 {
			return "", errKeychainItemNotFound
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func macKeychainSet(service, account, password string) error {
	// First try to delete existing entry
	_ = macKeychainDelete(service, account)

	cmd := exec.Command("security", "add-generic-password", "-s", service, "-a", account, "-w", password)
	return cmd.Run()
}

func macKeychainDelete(service, account string) error {
	cmd := exec.Command("security", "delete-generic-password", "-s", service, "-a", account)
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 44 {
			return errKeychainItemNotFound
		}
	}
	return err
}

// Linux Secret Service using secret-tool
func linuxSecretGet(service, account string) (string, error) {
	cmd := exec.Command("secret-tool", "lookup", "service", service, "account", account)
	out, err := cmd.Output()
	if err != nil {
		return "", errKeychainItemNotFound
	}
	return strings.TrimSpace(string(out)), nil
}

func linuxSecretSet(service, account, password string) error {
	cmd := exec.Command("secret-tool", "store", "--label", service+" credentials", "service", service, "account", account)
	cmd.Stdin = strings.NewReader(password)
	return cmd.Run()
}

func linuxSecretDelete(service, account string) error {
	cmd := exec.Command("secret-tool", "clear", "service", service, "account", account)
	return cmd.Run()
}

// Windows Credential Manager using cmdkey
func windowsCredentialGet(service, account string) (string, error) {
	target := service + "/" + account
	cmd := exec.Command("cmdkey", "/list:"+target)
	out, err := cmd.Output()
	if err != nil || !strings.Contains(string(out), target) {
		return "", errKeychainItemNotFound
	}
	// Windows doesn't expose passwords directly via cmdkey
	// We use a file-based fallback for actual storage
	return "", errors.New("windows credential read not supported, using file fallback")
}

func windowsCredentialSet(service, account, password string) error {
	target := service + "/" + account
	cmd := exec.Command("cmdkey", "/generic:"+target, "/user:"+account, "/pass:"+password)
	return cmd.Run()
}

func windowsCredentialDelete(service, account string) error {
	target := service + "/" + account
	cmd := exec.Command("cmdkey", "/delete:"+target)
	return cmd.Run()
}

// GetPreferredStorage returns the best available storage backend
func GetPreferredStorage() Storage {
	// Try keychain first
	keychain := NewKeychainStorage()
	if _, err := keychain.Load(); err == nil {
		return keychain
	}

	// Fall back to encrypted file
	return NewFileStorage("")
}

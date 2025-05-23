package keyring

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/encryption"
	"github.com/zalando/go-keyring"
)

const (
	// ServiceName is the service name used for keyring storage
	ServiceName = "export-trakt-4-letterboxd"
	// DefaultUsername is the default username for keyring entries
	DefaultUsername = "trakt-api"
)

var (
	// ErrCredentialNotFound is returned when a credential is not found
	ErrCredentialNotFound = errors.New("credential not found")
	// ErrUnsupportedBackend is returned when an unsupported backend is specified
	ErrUnsupportedBackend = errors.New("unsupported keyring backend")
	// ErrPermissionDenied is returned when file permissions are incorrect
	ErrPermissionDenied = errors.New("file permissions must be 0600 for security")
)

// Backend represents different credential storage backends
type Backend string

const (
	// SystemBackend uses the system keyring (keychain on macOS, etc.)
	SystemBackend Backend = "system"
	// EnvBackend uses environment variables
	EnvBackend Backend = "env"
	// FileBackend uses encrypted file storage
	FileBackend Backend = "file"
)

// Credential represents a stored credential
type Credential struct {
	Key   string
	Value string
}

// Manager handles credential storage and retrieval across different backends
type Manager struct {
	backend       Backend
	encryptionKey []byte
	filePath      string
}

// NewManager creates a new credential manager with the specified backend
func NewManager(backend Backend, options ...Option) (*Manager, error) {
	m := &Manager{
		backend: backend,
	}

	// Apply options
	for _, opt := range options {
		if err := opt(m); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Validate backend-specific requirements
	if err := m.validateBackend(); err != nil {
		return nil, fmt.Errorf("backend validation failed: %w", err)
	}

	return m, nil
}

// Option represents a configuration option for the Manager
type Option func(*Manager) error

// WithEncryptionKey sets the encryption key for file backend
func WithEncryptionKey(key []byte) Option {
	return func(m *Manager) error {
		if len(key) != encryption.AESKeySize {
			return encryption.ErrInvalidKeySize
		}
		m.encryptionKey = make([]byte, len(key))
		copy(m.encryptionKey, key)
		return nil
	}
}

// WithFilePath sets the file path for file backend
func WithFilePath(path string) Option {
	return func(m *Manager) error {
		m.filePath = path
		return nil
	}
}

// validateBackend ensures the backend is properly configured
func (m *Manager) validateBackend() error {
	switch m.backend {
	case SystemBackend:
		// System backend should work out of the box
		return nil
	case EnvBackend:
		// Environment backend requires no additional configuration
		return nil
	case FileBackend:
		if m.encryptionKey == nil {
			return errors.New("file backend requires encryption key")
		}
		if m.filePath == "" {
			return errors.New("file backend requires file path")
		}
		return nil
	default:
		return ErrUnsupportedBackend
	}
}

// Store stores a credential using the configured backend
func (m *Manager) Store(key, value string) error {
	switch m.backend {
	case SystemBackend:
		return m.storeSystem(key, value)
	case EnvBackend:
		return m.storeEnv(key, value)
	case FileBackend:
		return m.storeFile(key, value)
	default:
		return ErrUnsupportedBackend
	}
}

// Retrieve retrieves a credential using the configured backend
func (m *Manager) Retrieve(key string) (string, error) {
	switch m.backend {
	case SystemBackend:
		return m.retrieveSystem(key)
	case EnvBackend:
		return m.retrieveEnv(key)
	case FileBackend:
		return m.retrieveFile(key)
	default:
		return "", ErrUnsupportedBackend
	}
}

// Delete deletes a credential using the configured backend
func (m *Manager) Delete(key string) error {
	switch m.backend {
	case SystemBackend:
		return m.deleteSystem(key)
	case EnvBackend:
		return m.deleteEnv(key)
	case FileBackend:
		return m.deleteFile(key)
	default:
		return ErrUnsupportedBackend
	}
}

// List lists all stored credentials (returns keys only for security)
func (m *Manager) List() ([]string, error) {
	switch m.backend {
	case SystemBackend:
		return m.listSystem()
	case EnvBackend:
		return m.listEnv()
	case FileBackend:
		return m.listFile()
	default:
		return nil, ErrUnsupportedBackend
	}
}

// System backend implementations
func (m *Manager) storeSystem(key, value string) error {
	return keyring.Set(ServiceName, key, value)
}

func (m *Manager) retrieveSystem(key string) (string, error) {
	value, err := keyring.Get(ServiceName, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "password not found") {
			return "", ErrCredentialNotFound
		}
		return "", fmt.Errorf("failed to retrieve from system keyring: %w", err)
	}
	return value, nil
}

func (m *Manager) deleteSystem(key string) error {
	return keyring.Delete(ServiceName, key)
}

func (m *Manager) listSystem() ([]string, error) {
	// System keyring doesn't provide a way to list all keys
	// Return known credential keys
	knownKeys := []string{"client_id", "client_secret", "access_token", "encryption_key"}
	var existingKeys []string
	
	for _, key := range knownKeys {
		if _, err := m.retrieveSystem(key); err == nil {
			existingKeys = append(existingKeys, key)
		}
	}
	
	return existingKeys, nil
}

// Environment backend implementations
func (m *Manager) storeEnv(key, value string) error {
	envKey := m.getEnvKey(key)
	return os.Setenv(envKey, value)
}

func (m *Manager) retrieveEnv(key string) (string, error) {
	envKey := m.getEnvKey(key)
	value := os.Getenv(envKey)
	if value == "" {
		return "", ErrCredentialNotFound
	}
	return value, nil
}

func (m *Manager) deleteEnv(key string) error {
	envKey := m.getEnvKey(key)
	return os.Unsetenv(envKey)
}

func (m *Manager) listEnv() ([]string, error) {
	prefix := "TRAKT_"
	var keys []string
	
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, prefix) {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				// Convert back to internal key format
				envKey := parts[0]
				internalKey := strings.ToLower(strings.TrimPrefix(envKey, prefix))
				keys = append(keys, internalKey)
			}
		}
	}
	
	return keys, nil
}

func (m *Manager) getEnvKey(key string) string {
	// Convert internal key to environment variable name
	return "TRAKT_" + strings.ToUpper(key)
}

// File backend implementations (encrypted)
func (m *Manager) storeFile(key, value string) error {
	credentials, err := m.loadCredentialsFile()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load credentials file: %w", err)
	}

	if credentials == nil {
		credentials = make(map[string]string)
	}

	// Encrypt the value
	encryptor, err := encryption.NewEncryptor(m.encryptionKey)
	if err != nil {
		return fmt.Errorf("failed to create encryptor: %w", err)
	}
	defer encryptor.Destroy()

	encryptedValue, err := encryptor.Encrypt(value)
	if err != nil {
		return fmt.Errorf("failed to encrypt value: %w", err)
	}

	credentials[key] = encryptedValue

	return m.saveCredentialsFile(credentials)
}

func (m *Manager) retrieveFile(key string) (string, error) {
	credentials, err := m.loadCredentialsFile()
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrCredentialNotFound
		}
		return "", fmt.Errorf("failed to load credentials file: %w", err)
	}

	encryptedValue, exists := credentials[key]
	if !exists {
		return "", ErrCredentialNotFound
	}

	// Decrypt the value
	encryptor, err := encryption.NewEncryptor(m.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create encryptor: %w", err)
	}
	defer encryptor.Destroy()

	value, err := encryptor.Decrypt(encryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt value: %w", err)
	}

	return value, nil
}

func (m *Manager) deleteFile(key string) error {
	credentials, err := m.loadCredentialsFile()
	if err != nil {
		if os.IsNotExist(err) {
			return ErrCredentialNotFound
		}
		return fmt.Errorf("failed to load credentials file: %w", err)
	}

	if _, exists := credentials[key]; !exists {
		return ErrCredentialNotFound
	}

	delete(credentials, key)

	if len(credentials) == 0 {
		// Remove file if no credentials left
		return os.Remove(m.filePath)
	}

	return m.saveCredentialsFile(credentials)
}

func (m *Manager) listFile() ([]string, error) {
	credentials, err := m.loadCredentialsFile()
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to load credentials file: %w", err)
	}

	keys := make([]string, 0, len(credentials))
	for key := range credentials {
		keys = append(keys, key)
	}

	return keys, nil
}

func (m *Manager) loadCredentialsFile() (map[string]string, error) {
	// Check file permissions
	if err := m.checkFilePermissions(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, err
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode file data: %w", err)
	}

	// For simplicity, we store as encrypted JSON
	// In production, consider using a more structured format
	credentials := make(map[string]string)
	// Parse the decoded data as key=value pairs separated by newlines
	lines := strings.Split(string(decoded), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			credentials[parts[0]] = parts[1]
		}
	}

	return credentials, nil
}

func (m *Manager) saveCredentialsFile(credentials map[string]string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.filePath), 0700); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Build content as key=value pairs
	var lines []string
	for key, value := range credentials {
		lines = append(lines, key+"="+value)
	}
	content := strings.Join(lines, "\n")

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString([]byte(content))

	// Write with secure permissions
	if err := os.WriteFile(m.filePath, []byte(encoded), 0600); err != nil {
		return fmt.Errorf("failed to write credentials file: %w", err)
	}

	return nil
}

func (m *Manager) checkFilePermissions() error {
	info, err := os.Stat(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's ok
		}
		return err
	}

	// Check if permissions are 0600 (rw-------) or more restrictive
	mode := info.Mode()
	if mode&0077 != 0 {
		return ErrPermissionDenied
	}

	return nil
}

// Destroy securely clears sensitive data from memory
func (m *Manager) Destroy() {
	if m.encryptionKey != nil {
		for i := range m.encryptionKey {
			m.encryptionKey[i] = 0
		}
		m.encryptionKey = nil
	}
} 
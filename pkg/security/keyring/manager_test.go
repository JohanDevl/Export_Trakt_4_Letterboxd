package keyring

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/encryption"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name        string
		backend     Backend
		options     []Option
		expectError bool
	}{
		{
			name:        "system backend",
			backend:     SystemBackend,
			expectError: false,
		},
		{
			name:        "env backend",
			backend:     EnvBackend,
			expectError: false,
		},
		{
			name:        "unsupported backend",
			backend:     Backend("invalid"),
			expectError: true,
		},
		{
			name:        "file backend without key",
			backend:     FileBackend,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewManager(tt.backend, tt.options...)
			
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			
			if manager == nil {
				t.Fatal("Manager should not be nil")
			}
			
			if manager.backend != tt.backend {
				t.Errorf("Expected backend %s, got %s", tt.backend, manager.backend)
			}
		})
	}
}

func TestNewManagerWithFileBackend(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "keyring_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Generate encryption key
	key, err := encryption.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	filePath := filepath.Join(tempDir, "credentials.enc")
	
	manager, err := NewManager(FileBackend, 
		WithEncryptionKey(key),
		WithFilePath(filePath))
	
	if err != nil {
		t.Fatalf("Failed to create file backend manager: %v", err)
	}
	
	if manager.backend != FileBackend {
		t.Errorf("Expected backend %s, got %s", FileBackend, manager.backend)
	}
	
	if manager.filePath != filePath {
		t.Errorf("Expected file path %s, got %s", filePath, manager.filePath)
	}
}

func TestWithEncryptionKeyOption(t *testing.T) {
	key, err := encryption.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	manager := &Manager{}
	option := WithEncryptionKey(key)
	
	err = option(manager)
	if err != nil {
		t.Fatalf("WithEncryptionKey option failed: %v", err)
	}
	
	if len(manager.encryptionKey) != encryption.AESKeySize {
		t.Errorf("Expected key size %d, got %d", encryption.AESKeySize, len(manager.encryptionKey))
	}
}

func TestWithEncryptionKeyInvalidSize(t *testing.T) {
	invalidKey := []byte("too_short")
	
	manager := &Manager{}
	option := WithEncryptionKey(invalidKey)
	
	err := option(manager)
	if err == nil {
		t.Error("Expected error for invalid key size")
	}
	
	if err != encryption.ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got %v", err)
	}
}

func TestWithFilePathOption(t *testing.T) {
	testPath := "/tmp/test_credentials.enc"
	
	manager := &Manager{}
	option := WithFilePath(testPath)
	
	err := option(manager)
	if err != nil {
		t.Fatalf("WithFilePath option failed: %v", err)
	}
	
	if manager.filePath != testPath {
		t.Errorf("Expected file path %s, got %s", testPath, manager.filePath)
	}
}

func TestEnvBackendOperations(t *testing.T) {
	manager, err := NewManager(EnvBackend)
	if err != nil {
		t.Fatal(err)
	}

	testKey := "TEST_CREDENTIAL"
	testValue := "test_secret_value"

	// Clean up any existing env var
	oldValue := os.Getenv("TRAKT_" + testKey)
	defer func() {
		if oldValue != "" {
			os.Setenv("TRAKT_"+testKey, oldValue)
		} else {
			os.Unsetenv("TRAKT_" + testKey)
		}
	}()

	// Test Store
	err = manager.Store(testKey, testValue)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Verify environment variable was set
	envKey := "TRAKT_" + testKey
	if os.Getenv(envKey) != testValue {
		t.Errorf("Environment variable %s not set correctly", envKey)
	}

	// Test Retrieve
	retrieved, err := manager.Retrieve(testKey)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	
	if retrieved != testValue {
		t.Errorf("Expected %s, got %s", testValue, retrieved)
	}

	// Test List
	keys, err := manager.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	
	found := false
	expectedKey := strings.ToLower(testKey) // listEnv returns keys in lowercase
	for _, key := range keys {
		if key == expectedKey {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Key %s not found in list", expectedKey)
	}

	// Test Delete
	err = manager.Delete(testKey)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify environment variable was unset
	if os.Getenv(envKey) != "" {
		t.Errorf("Environment variable %s should be unset", envKey)
	}

	// Test retrieve after delete
	_, err = manager.Retrieve(testKey)
	if err != ErrCredentialNotFound {
		t.Errorf("Expected ErrCredentialNotFound, got %v", err)
	}
}

func TestFileBackendOperations(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "keyring_file_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Generate encryption key
	key, err := encryption.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	filePath := filepath.Join(tempDir, "credentials.enc")
	
	manager, err := NewManager(FileBackend, 
		WithEncryptionKey(key),
		WithFilePath(filePath))
	if err != nil {
		t.Fatal(err)
	}

	testKey := "test_key"
	testValue := "test_secret_value"

	// Test Store
	err = manager.Store(testKey, testValue)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Credentials file should have been created")
	}

	// Test Retrieve
	retrieved, err := manager.Retrieve(testKey)
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}
	
	if retrieved != testValue {
		t.Errorf("Expected %s, got %s", testValue, retrieved)
	}

	// Test List
	keys, err := manager.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	
	found := false
	for _, key := range keys {
		if key == testKey {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Key %s not found in list", testKey)
	}

	// Test storing multiple credentials
	err = manager.Store("key2", "value2")
	if err != nil {
		t.Fatalf("Store second credential failed: %v", err)
	}

	keys, err = manager.List()
	if err != nil {
		t.Fatal(err)
	}
	
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}

	// Test Delete
	err = manager.Delete(testKey)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Test retrieve after delete
	_, err = manager.Retrieve(testKey)
	if err != ErrCredentialNotFound {
		t.Errorf("Expected ErrCredentialNotFound, got %v", err)
	}

	// Verify other credential still exists
	retrieved, err = manager.Retrieve("key2")
	if err != nil {
		t.Fatalf("Second credential should still exist: %v", err)
	}
	if retrieved != "value2" {
		t.Errorf("Expected value2, got %s", retrieved)
	}
}

func TestFileBackendFilePermissions(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "keyring_permissions_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Generate encryption key
	key, err := encryption.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	filePath := filepath.Join(tempDir, "credentials.enc")
	
	manager, err := NewManager(FileBackend, 
		WithEncryptionKey(key),
		WithFilePath(filePath))
	if err != nil {
		t.Fatal(err)
	}

	// Store a credential to create the file
	err = manager.Store("test", "value")
	if err != nil {
		t.Fatal(err)
	}

	// Check file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatal(err)
	}

	mode := info.Mode()
	expectedMode := os.FileMode(0600)
	
	if mode != expectedMode {
		t.Errorf("Expected file mode %o, got %o", expectedMode, mode)
	}
}

func TestUnsupportedBackendOperations(t *testing.T) {
	manager := &Manager{backend: Backend("invalid")}

	// Test Store
	err := manager.Store("key", "value")
	if err != ErrUnsupportedBackend {
		t.Errorf("Expected ErrUnsupportedBackend, got %v", err)
	}

	// Test Retrieve
	_, err = manager.Retrieve("key")
	if err != ErrUnsupportedBackend {
		t.Errorf("Expected ErrUnsupportedBackend, got %v", err)
	}

	// Test Delete
	err = manager.Delete("key")
	if err != ErrUnsupportedBackend {
		t.Errorf("Expected ErrUnsupportedBackend, got %v", err)
	}

	// Test List
	_, err = manager.List()
	if err != ErrUnsupportedBackend {
		t.Errorf("Expected ErrUnsupportedBackend, got %v", err)
	}
}

func TestEnvBackendGetEnvKey(t *testing.T) {
	manager := &Manager{backend: EnvBackend}

	tests := []struct {
		input    string
		expected string
	}{
		{"client_id", "TRAKT_CLIENT_ID"},
		{"CLIENT_SECRET", "TRAKT_CLIENT_SECRET"},
		{"access_token", "TRAKT_ACCESS_TOKEN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := manager.getEnvKey(tt.input)
			if result != tt.expected {
				t.Errorf("getEnvKey(%s) = %s, expected %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestRetrieveNonExistentCredential(t *testing.T) {
	tests := []struct {
		name    string
		backend Backend
		setup   func() (*Manager, func(), error)
	}{
		{
			name:    "env backend",
			backend: EnvBackend,
			setup: func() (*Manager, func(), error) {
				manager, err := NewManager(EnvBackend)
				return manager, func() {}, err
			},
		},
		{
			name:    "file backend",
			backend: FileBackend,
			setup: func() (*Manager, func(), error) {
				tempDir, err := os.MkdirTemp("", "keyring_nonexistent_test")
				if err != nil {
					return nil, nil, err
				}
				
				key, err := encryption.GenerateKey()
				if err != nil {
					os.RemoveAll(tempDir)
					return nil, nil, err
				}

				filePath := filepath.Join(tempDir, "credentials.enc")
				manager, err := NewManager(FileBackend, 
					WithEncryptionKey(key),
					WithFilePath(filePath))
				
				cleanup := func() { os.RemoveAll(tempDir) }
				return manager, cleanup, err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, cleanup, err := tt.setup()
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			_, err = manager.Retrieve("nonexistent_key")
			if err != ErrCredentialNotFound {
				t.Errorf("Expected ErrCredentialNotFound, got %v", err)
			}
		})
	}
}

func TestDestroy(t *testing.T) {
	// Generate encryption key
	key, err := encryption.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	manager := &Manager{
		backend:       FileBackend,
		encryptionKey: make([]byte, len(key)),
	}
	copy(manager.encryptionKey, key)

	// Call Destroy
	manager.Destroy()

	// Verify encryption key is cleared
	for i, b := range manager.encryptionKey {
		if b != 0 {
			t.Errorf("Encryption key byte %d should be 0, got %d", i, b)
		}
	}
}

func TestConstants(t *testing.T) {
	if ServiceName == "" {
		t.Error("ServiceName should not be empty")
	}
	
	if DefaultUsername == "" {
		t.Error("DefaultUsername should not be empty")
	}
}

func TestBackendConstants(t *testing.T) {
	backends := []Backend{SystemBackend, EnvBackend, FileBackend}
	
	for _, backend := range backends {
		if string(backend) == "" {
			t.Errorf("Backend %v should not be empty", backend)
		}
	}
}

func TestErrorConstants(t *testing.T) {
	errors := []error{
		ErrCredentialNotFound,
		ErrUnsupportedBackend,
		ErrPermissionDenied,
	}
	
	for _, err := range errors {
		if err.Error() == "" {
			t.Errorf("Error %v should have a message", err)
		}
	}
}

func TestValidateBackendErrors(t *testing.T) {
	tests := []struct {
		name    string
		manager *Manager
		wantErr bool
	}{
		{
			name:    "file backend without encryption key",
			manager: &Manager{backend: FileBackend},
			wantErr: true,
		},
		{
			name: "file backend without file path",
			manager: &Manager{
				backend:       FileBackend,
				encryptionKey: make([]byte, 32),
			},
			wantErr: true,
		},
		{
			name: "file backend valid",
			manager: &Manager{
				backend:       FileBackend,
				encryptionKey: make([]byte, 32),
				filePath:      "/tmp/test.enc",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.manager.validateBackend()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBackend() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Note: System backend tests are not included here because they would require
// actual system keyring access which might not be available in CI environments.
// In a real-world scenario, you might want to mock the keyring library or
// run system backend tests only in specific environments. 
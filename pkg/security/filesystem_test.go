package security

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/audit"
)

// testFileSystemConfig creates a filesystem config suitable for testing
// It removes /var from restricted paths to allow temporary directories on macOS
func testFileSystemConfig(tempDir string) FileSystemConfig {
	config := DefaultFileSystemConfig()
	
	// Add both the original tempDir and its resolved version to allowed paths
	allowedPaths := []string{tempDir}
	if resolvedTempDir, err := filepath.EvalSymlinks(tempDir); err == nil && resolvedTempDir != tempDir {
		allowedPaths = append(allowedPaths, resolvedTempDir)
	}
	config.AllowedBasePaths = allowedPaths
	
	// Remove /var from restricted paths to allow macOS temp directories
	var filteredRestricted []string
	for _, path := range config.RestrictedPaths {
		if path != "/var" {
			filteredRestricted = append(filteredRestricted, path)
		}
	}
	// Add macOS-specific restricted paths
	filteredRestricted = append(filteredRestricted, "/private/etc")
	config.RestrictedPaths = filteredRestricted
	
	return config
}

func TestNewFileSystemSecurity(t *testing.T) {
	config := DefaultFileSystemConfig()
	fs := NewFileSystemSecurity(config, nil)

	if fs == nil {
		t.Fatal("NewFileSystemSecurity returned nil")
	}

	if fs.config.EnforcePermissions != config.EnforcePermissions {
		t.Errorf("Expected EnforcePermissions %v, got %v", config.EnforcePermissions, fs.config.EnforcePermissions)
	}
}

func TestValidatePath(t *testing.T) {
	config := DefaultFileSystemConfig()
	fs := NewFileSystemSecurity(config, nil)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid relative path",
			path:    "./config/test.toml",
			wantErr: false,
		},
		{
			name:    "valid export path",
			path:    "./exports/data.csv",
			wantErr: false,
		},
		{
			name:    "path traversal attack",
			path:    "../../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "windows path traversal",
			path:    "..\\..\\windows\\system32",
			wantErr: true,
		},
		{
			name:    "restricted system path",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "restricted var path",
			path:    "/var/log/system.log",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fs.ValidatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSecureCreateFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := testFileSystemConfig(tempDir)
	fs := NewFileSystemSecurity(config, nil)

	testPath := filepath.Join(tempDir, "test.txt")

	// Test creating a file
	file, err := fs.SecureCreateFile(testPath, 0600)
	if err != nil {
		t.Fatalf("SecureCreateFile failed: %v", err)
	}
	defer file.Close()

	// Check file permissions
	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedMode := os.FileMode(0600)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("Expected file mode %v, got %v", expectedMode, info.Mode().Perm())
	}
}

func TestSecureWriteFile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := testFileSystemConfig(tempDir)
	fs := NewFileSystemSecurity(config, nil)

	testPath := filepath.Join(tempDir, "test.txt")
	testData := []byte("test data")

	// Test writing config file
	err = fs.SecureWriteFile(testPath, testData, true)
	if err != nil {
		t.Fatalf("SecureWriteFile failed: %v", err)
	}

	// Check file permissions for config file
	info, err := os.Stat(testPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	expectedMode := os.FileMode(0600)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("Expected config file mode %v, got %v", expectedMode, info.Mode().Perm())
	}

	// Check file content
	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if string(content) != string(testData) {
		t.Errorf("Expected content %s, got %s", testData, content)
	}

	// Test writing data file
	dataPath := filepath.Join(tempDir, "data.txt")
	err = fs.SecureWriteFile(dataPath, testData, false)
	if err != nil {
		t.Fatalf("SecureWriteFile failed for data file: %v", err)
	}

	// Check file permissions for data file
	info, err = os.Stat(dataPath)
	if err != nil {
		t.Fatalf("Failed to stat data file: %v", err)
	}

	expectedMode = os.FileMode(0644)
	if info.Mode().Perm() != expectedMode {
		t.Errorf("Expected data file mode %v, got %v", expectedMode, info.Mode().Perm())
	}
}

func TestValidateFilePermissions(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := testFileSystemConfig(tempDir)
	fs := NewFileSystemSecurity(config, nil)

	// Create test file with correct permissions
	testPath := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testPath, []byte("test"), 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Should pass validation
	err = fs.ValidateFilePermissions(testPath)
	if err != nil {
		t.Errorf("ValidateFilePermissions failed for correct permissions: %v", err)
	}

	// Create file with incorrect permissions
	badPath := filepath.Join(tempDir, "bad.txt")
	err = os.WriteFile(badPath, []byte("test"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Force world-writable permissions
	err = os.Chmod(badPath, 0666) // rw-rw-rw- (world writable)
	if err != nil {
		t.Fatal(err)
	}



	// Should fail validation for overly permissive file
	err = fs.ValidateFilePermissions(badPath)
	if err == nil {
		t.Error("ValidateFilePermissions should fail for world-writable file")
	}
}

func TestCleanupTempFiles(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := testFileSystemConfig(tempDir)
	fs := NewFileSystemSecurity(config, nil)

	// Create old temp file
	oldFile := filepath.Join(tempDir, "old_temp.txt")
	err = os.WriteFile(oldFile, []byte("old"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Make file old by changing its modification time
	oldTime := time.Now().Add(-2 * time.Hour)
	err = os.Chtimes(oldFile, oldTime, oldTime)
	if err != nil {
		t.Fatal(err)
	}

	// Create new temp file
	newFile := filepath.Join(tempDir, "new_temp.txt")
	err = os.WriteFile(newFile, []byte("new"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Cleanup files older than 1 hour
	err = fs.CleanupTempFiles(tempDir, time.Hour)
	if err != nil {
		t.Fatalf("CleanupTempFiles failed: %v", err)
	}

	// Old file should be deleted
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old temp file should have been deleted")
	}

	// New file should still exist
	if _, err := os.Stat(newFile); err != nil {
		t.Error("New temp file should still exist")
	}
}

func TestSecureDelete(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := testFileSystemConfig(tempDir)
	fs := NewFileSystemSecurity(config, nil)

	// Create test file
	testPath := filepath.Join(tempDir, "test.txt")
	testData := []byte("sensitive data")
	err = os.WriteFile(testPath, testData, 0600)
	if err != nil {
		t.Fatal(err)
	}

	// Delete file securely
	err = fs.SecureDeleteFile(testPath)
	if err != nil {
		t.Fatalf("SecureDeleteFile failed: %v", err)
	}

	// File should be deleted
	if _, err := os.Stat(testPath); !os.IsNotExist(err) {
		t.Error("File should have been deleted")
	}
}

func TestWithAuditLogging(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create audit logger
	auditConfig := audit.Config{
		LogLevel:     "debug",
		OutputFormat: "json",
		LogFile:      filepath.Join(tempDir, "audit.log"),
	}
	auditLogger, err := audit.NewLogger(auditConfig)
	if err != nil {
		t.Fatal(err)
	}

	config := testFileSystemConfig(tempDir)
	fs := NewFileSystemSecurity(config, auditLogger)

	// Test path validation with audit logging
	err = fs.ValidatePath("../../../etc/passwd")
	if err == nil {
		t.Error("Expected path validation to fail")
	}

	// Test secure file creation with audit logging
	testPath := filepath.Join(tempDir, "test.txt")
	file, err := fs.SecureCreateFile(testPath, 0600)
	if err != nil {
		t.Fatalf("SecureCreateFile failed: %v", err)
	}
	file.Close()

	// Check that audit log was created
	if _, err := os.Stat(auditConfig.LogFile); err != nil {
		t.Error("Audit log file should have been created")
	}
}

func TestFileSystemConfig(t *testing.T) {
	config := DefaultFileSystemConfig()

	// Test default values
	if !config.EnforcePermissions {
		t.Error("Expected EnforcePermissions to be true by default")
	}

	if config.ConfigFileMode != 0600 {
		t.Errorf("Expected ConfigFileMode 0600, got %o", config.ConfigFileMode)
	}

	if config.DataFileMode != 0644 {
		t.Errorf("Expected DataFileMode 0644, got %o", config.DataFileMode)
	}

	if config.DirectoryMode != 0750 {
		t.Errorf("Expected DirectoryMode 0750, got %o", config.DirectoryMode)
	}

	if config.MaxFileSize != 100*1024*1024 {
		t.Errorf("Expected MaxFileSize 100MB, got %d", config.MaxFileSize)
	}

	if !config.CheckSymlinks {
		t.Error("Expected CheckSymlinks to be true by default")
	}

	// Test that allowed paths include expected directories
	expectedPaths := []string{"./config", "./exports", "./logs", "./temp"}
	for _, expected := range expectedPaths {
		found := false
		for _, allowed := range config.AllowedBasePaths {
			if allowed == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected allowed path %s not found", expected)
		}
	}

	// Test that restricted paths include system directories
	expectedRestricted := []string{"/etc", "/var", "/usr", "/sys", "/proc", "/dev"}
	for _, expected := range expectedRestricted {
		found := false
		for _, restricted := range config.RestrictedPaths {
			if restricted == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected restricted path %s not found", expected)
		}
	}
}

func TestSymlinkValidation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filesystem_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	config := testFileSystemConfig(tempDir)
	config.CheckSymlinks = true
	fs := NewFileSystemSecurity(config, nil)

	// Create a target file
	targetFile := filepath.Join(tempDir, "target.txt")
	err = os.WriteFile(targetFile, []byte("target"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create a symlink to the target file
	symlinkPath := filepath.Join(tempDir, "symlink.txt")
	err = os.Symlink(targetFile, symlinkPath)
	if err != nil {
		t.Skip("Symlink creation not supported on this system")
	}

	// Valid symlink should pass validation
	err = fs.ValidatePath(symlinkPath)
	if err != nil {
		t.Errorf("Valid symlink should pass validation: %v", err)
	}

	// Create symlink pointing outside allowed paths
	outsideTarget := "/etc"  // Use directory instead of file
	badSymlink := filepath.Join(tempDir, "bad_symlink.txt")
	err = os.Symlink(outsideTarget, badSymlink)
	if err != nil {
		t.Skipf("Symlink creation not supported on this system: %v", err)
	}



	// Bad symlink should fail validation
	err = fs.ValidatePath(badSymlink)
	if err == nil {
		t.Error("Symlink pointing outside allowed paths should fail validation")
	}
} 
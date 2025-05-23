package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security/audit"
)

// FileSystemSecurity provides secure file system operations
type FileSystemSecurity struct {
	auditLog       *audit.Logger
	allowedPaths   []string
	restrictedDirs []string
	config         FileSystemConfig
}

// FileSystemConfig holds configuration for file system security
type FileSystemConfig struct {
	EnforcePermissions bool     `toml:"enforce_permissions"`
	ConfigFileMode     os.FileMode `toml:"config_file_mode"`     // Default: 0600
	DataFileMode       os.FileMode `toml:"data_file_mode"`       // Default: 0644  
	DirectoryMode      os.FileMode `toml:"directory_mode"`       // Default: 0750
	AllowedBasePaths   []string    `toml:"allowed_base_paths"`
	RestrictedPaths    []string    `toml:"restricted_paths"`
	MaxFileSize        int64       `toml:"max_file_size"`        // Maximum file size in bytes
	CheckSymlinks      bool        `toml:"check_symlinks"`       // Check for symlink attacks
}

// DefaultFileSystemConfig returns secure default file system configuration
func DefaultFileSystemConfig() FileSystemConfig {
	return FileSystemConfig{
		EnforcePermissions: true,
		ConfigFileMode:     0600,  // Owner read/write only
		DataFileMode:       0644,  // Owner read/write, group/others read
		DirectoryMode:      0750,  // Owner read/write/execute, group read/execute
		AllowedBasePaths: []string{
			"./config",
			"./exports", 
			"./logs",
			"./temp",
		},
		RestrictedPaths: []string{
			"/etc",
			"/var",
			"/usr",
			"/sys",
			"/proc",
			"/dev",
		},
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		CheckSymlinks: true,
	}
}

// NewFileSystemSecurity creates a new file system security manager
func NewFileSystemSecurity(config FileSystemConfig, auditLog *audit.Logger) *FileSystemSecurity {
	return &FileSystemSecurity{
		auditLog:       auditLog,
		allowedPaths:   config.AllowedBasePaths,
		restrictedDirs: config.RestrictedPaths,
		config:         config,
	}
}

// ValidatePath checks if a file path is safe to access
func (fs *FileSystemSecurity) ValidatePath(path string) error {
	// Clean and resolve the path
	cleanPath := filepath.Clean(path)
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check for path traversal attacks
	if strings.Contains(cleanPath, "..") {
		fs.logSecurityViolation("path_traversal", path, "Attempted path traversal attack")
		return fmt.Errorf("path traversal attempt detected: %s", path)
	}

	// Check if path is in restricted directories
	for _, restricted := range fs.restrictedDirs {
		if strings.HasPrefix(absPath, restricted) {
			fs.logSecurityViolation("restricted_path", path, fmt.Sprintf("Access to restricted path: %s", restricted))
			return fmt.Errorf("access to restricted path denied: %s", path)
		}
	}

	// Check if path is in allowed base paths (if specified)
	if len(fs.allowedPaths) > 0 {
		allowed := false
		for _, allowedPath := range fs.allowedPaths {
			absAllowed, err := filepath.Abs(allowedPath)
			if err != nil {
				continue
			}
			if strings.HasPrefix(absPath, absAllowed) {
				allowed = true
				break
			}
		}
		if !allowed {
			fs.logSecurityViolation("unauthorized_path", path, "Access to unauthorized path")
			return fmt.Errorf("access to unauthorized path denied: %s", path)
		}
	}

	// Check for symlink attacks if enabled
	if fs.config.CheckSymlinks {
		if err := fs.checkSymlinks(absPath); err != nil {
			return fmt.Errorf("symlink security check failed: %w", err)
		}
	}

	return nil
}

// SecureCreateFile creates a file with secure permissions
func (fs *FileSystemSecurity) SecureCreateFile(path string, mode os.FileMode) (*os.File, error) {
	// Validate path first
	if err := fs.ValidatePath(path); err != nil {
		return nil, err
	}

	// Ensure directory exists with secure permissions
	dir := filepath.Dir(path)
	if err := fs.SecureCreateDir(dir); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file with specified permissions
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		fs.logFileOperation("create_file", path, "failed", err.Error())
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	// Verify permissions were set correctly
	if fs.config.EnforcePermissions {
		if err := fs.enforceFilePermissions(path, mode); err != nil {
			file.Close()
			os.Remove(path)
			return nil, fmt.Errorf("failed to enforce permissions: %w", err)
		}
	}

	fs.logFileOperation("create_file", path, "success", "")
	return file, nil
}

// SecureCreateDir creates a directory with secure permissions
func (fs *FileSystemSecurity) SecureCreateDir(path string) error {
	// Validate path first
	if err := fs.ValidatePath(path); err != nil {
		return err
	}

	// Create directory with secure permissions
	if err := os.MkdirAll(path, fs.config.DirectoryMode); err != nil {
		fs.logFileOperation("create_dir", path, "failed", err.Error())
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Enforce permissions if required
	if fs.config.EnforcePermissions {
		if err := fs.enforceFilePermissions(path, fs.config.DirectoryMode); err != nil {
			return fmt.Errorf("failed to enforce directory permissions: %w", err)
		}
	}

	fs.logFileOperation("create_dir", path, "success", "")
	return nil
}

// SecureWriteFile writes data to a file with secure permissions
func (fs *FileSystemSecurity) SecureWriteFile(path string, data []byte, isConfig bool) error {
	// Determine appropriate mode
	mode := fs.config.DataFileMode
	if isConfig {
		mode = fs.config.ConfigFileMode
	}

	// Check file size limit
	if int64(len(data)) > fs.config.MaxFileSize {
		return fmt.Errorf("file size exceeds limit: %d > %d", len(data), fs.config.MaxFileSize)
	}

	// Create file with secure permissions
	file, err := fs.SecureCreateFile(path, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write data
	if _, err := file.Write(data); err != nil {
		fs.logFileOperation("write_file", path, "failed", err.Error())
		return fmt.Errorf("failed to write file: %w", err)
	}

	fs.logFileOperation("write_file", path, "success", "")
	return nil
}

// ValidateFilePermissions checks if a file has secure permissions
func (fs *FileSystemSecurity) ValidateFilePermissions(path string) error {
	if !fs.config.EnforcePermissions {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	mode := info.Mode()
	
	// Check if file is too permissive
	if mode&0002 != 0 { // World writable
		return fmt.Errorf("file is world writable: %s", path)
	}

	// For config files, ensure they're only readable by owner
	if strings.Contains(path, "config") || strings.Contains(path, "credential") {
		if mode&0077 != 0 { // Group or world readable
			return fmt.Errorf("config file has overly permissive permissions: %s", path)
		}
	}

	return nil
}

// SecureDeleteFile securely deletes a file
func (fs *FileSystemSecurity) SecureDeleteFile(path string) error {
	// Validate path first
	if err := fs.ValidatePath(path); err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to do
	}

	// Remove file
	if err := os.Remove(path); err != nil {
		fs.logFileOperation("delete_file", path, "failed", err.Error())
		return fmt.Errorf("failed to delete file: %w", err)
	}

	fs.logFileOperation("delete_file", path, "success", "")
	return nil
}

// CleanupTempFiles removes temporary files older than specified duration
func (fs *FileSystemSecurity) CleanupTempFiles(tempDir string, maxAge time.Duration) error {
	// Validate temp directory path
	if err := fs.ValidatePath(tempDir); err != nil {
		return err
	}

	// Check if directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		return nil // Directory doesn't exist, nothing to do
	}

	cutoff := time.Now().Add(-maxAge)
	var cleanedCount int

	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is old enough to clean
		if info.ModTime().Before(cutoff) {
			if err := fs.SecureDeleteFile(path); err != nil {
				fs.logFileOperation("cleanup_temp", path, "failed", err.Error())
			} else {
				cleanedCount++
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk temp directory: %w", err)
	}

	if cleanedCount > 0 && fs.auditLog != nil {
		fs.auditLog.LogSystemEvent(audit.DataAccess, "filesystem", "cleanup_temp", "success")
	}

	return nil
}

// enforceFilePermissions sets the correct permissions on a file
func (fs *FileSystemSecurity) enforceFilePermissions(path string, mode os.FileMode) error {
	if err := os.Chmod(path, mode); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Verify permissions were set correctly
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to verify permissions: %w", err)
	}

	if info.Mode()&os.ModePerm != mode {
		return fmt.Errorf("permissions not set correctly: expected %o, got %o", 
			mode, info.Mode()&os.ModePerm)
	}

	return nil
}

// checkSymlinks checks for symlink attacks in the path
func (fs *FileSystemSecurity) checkSymlinks(path string) error {
	// First check if the final path itself is a symlink
	info, err := os.Lstat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check symlink: %w", err)
		}
		// File doesn't exist, check parent components
	} else if info.Mode()&os.ModeSymlink != 0 {
		// The final path is a symlink, resolve and validate it
		realPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return fmt.Errorf("failed to resolve symlink: %w", err)
		}

		// Check if symlink points outside allowed paths
		if err := fs.ValidatePath(realPath); err != nil {
			fs.logSecurityViolation("symlink_attack", path, 
				fmt.Sprintf("Symlink points to restricted location: %s -> %s", path, realPath))
			return fmt.Errorf("symlink points to restricted location: %s", realPath)
		}
	}

	// Check each component of the path for symlinks
	components := strings.Split(path, string(filepath.Separator))
	currentPath := ""

	for _, component := range components {
		if component == "" {
			continue
		}

		if currentPath == "" {
			currentPath = component
		} else {
			currentPath = filepath.Join(currentPath, component)
		}

		// Skip the final path since we already checked it above
		if currentPath == path {
			continue
		}

		// Check if current path component is a symlink
		info, err := os.Lstat(currentPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue // Path doesn't exist yet, that's OK
			}
			return fmt.Errorf("failed to check symlink: %w", err)
		}

		if info.Mode()&os.ModeSymlink != 0 {
			// Resolve symlink and check if it's safe
			realPath, err := filepath.EvalSymlinks(currentPath)
			if err != nil {
				return fmt.Errorf("failed to resolve symlink: %w", err)
			}

			// Check if symlink points outside allowed paths
			if err := fs.ValidatePath(realPath); err != nil {
				fs.logSecurityViolation("symlink_attack", path, 
					fmt.Sprintf("Symlink points to restricted location: %s -> %s", currentPath, realPath))
				return fmt.Errorf("symlink points to restricted location: %s", realPath)
			}
		}
	}

	return nil
}

// logFileOperation logs file system operations
func (fs *FileSystemSecurity) logFileOperation(operation, path, result, details string) {
	if fs.auditLog != nil {
		fs.auditLog.LogEvent(audit.AuditEvent{
			EventType: audit.DataAccess,
			Severity:  audit.SeverityLow,
			Source:    "filesystem_security",
			Action:    operation,
			Target:    path,
			Result:    result,
			Message:   fmt.Sprintf("File operation: %s on %s", operation, path),
			Details: map[string]interface{}{
				"operation": operation,
				"path":      path,
				"details":   details,
			},
		})
	}
}

// logSecurityViolation logs security violations
func (fs *FileSystemSecurity) logSecurityViolation(violationType, path, description string) {
	if fs.auditLog != nil {
		fs.auditLog.LogEvent(audit.AuditEvent{
			EventType: audit.SecurityViolation,
			Severity:  audit.SeverityHigh,
			Source:    "filesystem_security",
			Action:    violationType,
			Target:    path,
			Result:    "blocked",
			Message:   fmt.Sprintf("Security violation: %s", description),
			Details: map[string]interface{}{
				"violation_type": violationType,
				"path":          path,
				"description":   description,
			},
		})
	}
} 
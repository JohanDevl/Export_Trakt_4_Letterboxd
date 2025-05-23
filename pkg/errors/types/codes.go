package types

// Error Categories and Codes as defined in issue #13

// Network errors
const (
	ErrNetworkTimeout     = "NET_001"
	ErrNetworkUnavailable = "NET_002"
	ErrRateLimited        = "NET_003"
	ErrConnectionRefused  = "NET_004"
	ErrDNSResolution      = "NET_005"
)

// Authentication errors
const (
	ErrInvalidCredentials = "AUTH_001"
	ErrTokenExpired       = "AUTH_002"
	ErrUnauthorized       = "AUTH_003"
	ErrForbidden          = "AUTH_004"
	ErrAPIKeyMissing      = "AUTH_005"
)

// Validation errors
const (
	ErrInvalidConfig      = "VAL_001"
	ErrInvalidInput       = "VAL_002"
	ErrMissingRequired    = "VAL_003"
	ErrInvalidFormat      = "VAL_004"
	ErrOutOfRange         = "VAL_005"
)

// Operation errors
const (
	ErrExportFailed       = "OP_001"
	ErrImportFailed       = "OP_002"
	ErrFileSystem         = "OP_003"
	ErrProcessingFailed   = "OP_004"
	ErrOperationCanceled  = "OP_005"
	ErrOperationFailed    = "OP_006"
)

// Data errors
const (
	ErrDataCorrupted      = "DATA_001"
	ErrDataMissing        = "DATA_002"
	ErrDataFormat         = "DATA_003"
	ErrDataIntegrity      = "DATA_004"
)

// Configuration errors
const (
	ErrConfigMissing      = "CFG_001"
	ErrConfigCorrupted    = "CFG_002"
	ErrConfigPermissions  = "CFG_003"
	ErrConfigSyntax       = "CFG_004"
)

// System errors
const (
	ErrSystemResource     = "SYS_001"
	ErrSystemPermission   = "SYS_002"
	ErrSystemDisk         = "SYS_003"
	ErrSystemMemory       = "SYS_004"
)

// ErrorCategory represents error categories for classification
type ErrorCategory string

const (
	CategoryNetwork       ErrorCategory = "network"
	CategoryAuthentication ErrorCategory = "authentication"
	CategoryValidation     ErrorCategory = "validation"
	CategoryOperation      ErrorCategory = "operation"
	CategoryData           ErrorCategory = "data"
	CategoryConfiguration  ErrorCategory = "configuration"
	CategorySystem         ErrorCategory = "system"
)

// GetErrorCategory returns the category for a given error code
func GetErrorCategory(code string) ErrorCategory {
	switch {
	case code[:3] == "NET":
		return CategoryNetwork
	case code[:4] == "AUTH":
		return CategoryAuthentication
	case code[:3] == "VAL":
		return CategoryValidation
	case code[:2] == "OP":
		return CategoryOperation
	case code[:4] == "DATA":
		return CategoryData
	case code[:3] == "CFG":
		return CategoryConfiguration
	case code[:3] == "SYS":
		return CategorySystem
	default:
		return CategoryOperation // default category
	}
}

// IsRetryableError determines if an error should be retried
func IsRetryableError(code string) bool {
	retryableCodes := map[string]bool{
		ErrNetworkTimeout:     true,
		ErrNetworkUnavailable: true,
		ErrRateLimited:        true,
		ErrConnectionRefused:  true,
		ErrTokenExpired:       true,
		ErrSystemResource:     true,
		ErrSystemDisk:         false, // Usually not retryable
		ErrSystemMemory:       false, // Usually not retryable
	}
	
	return retryableCodes[code]
}

// IsTemporaryError determines if an error is temporary
func IsTemporaryError(code string) bool {
	temporaryCodes := map[string]bool{
		ErrNetworkTimeout:     true,
		ErrNetworkUnavailable: true,
		ErrRateLimited:        true,
		ErrSystemResource:     true,
		ErrTokenExpired:       true,
	}
	
	return temporaryCodes[code]
} 
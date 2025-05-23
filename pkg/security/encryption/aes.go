package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

const (
	// AESKeySize is the size of the AES-256 key in bytes
	AESKeySize = 32
	// NonceSize is the size of the GCM nonce in bytes
	NonceSize = 12
)

var (
	// ErrInvalidKeySize is returned when the encryption key has an invalid size
	ErrInvalidKeySize = errors.New("invalid key size: must be 32 bytes for AES-256")
	// ErrInvalidCiphertext is returned when the ciphertext is too short or invalid
	ErrInvalidCiphertext = errors.New("invalid ciphertext: too short or malformed")
	// ErrDecryptionFailed is returned when decryption fails
	ErrDecryptionFailed = errors.New("decryption failed: invalid key or corrupted data")
)

// Encryptor provides AES-256-GCM encryption and decryption
type Encryptor struct {
	key []byte
}

// NewEncryptor creates a new AES-256 encryptor with the provided key
func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != AESKeySize {
		return nil, ErrInvalidKeySize
	}

	// Create a copy to prevent external modification
	keyClone := make([]byte, len(key))
	copy(keyClone, key)

	return &Encryptor{
		key: keyClone,
	}, nil
}

// NewEncryptorFromPassword creates a new encryptor using a password-derived key
func NewEncryptorFromPassword(password string, salt []byte) (*Encryptor, error) {
	if len(salt) == 0 {
		return nil, errors.New("salt cannot be empty")
	}

	// Derive key using SHA-256 (for simplicity, in production consider using PBKDF2 or Argon2)
	hash := sha256.New()
	hash.Write([]byte(password))
	hash.Write(salt)
	key := hash.Sum(nil)

	return NewEncryptor(key)
}

// GenerateKey generates a cryptographically secure random key for AES-256
func GenerateKey() ([]byte, error) {
	key := make([]byte, AESKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	return key, nil
}

// GenerateSalt generates a cryptographically secure random salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 16) // 128-bit salt
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}
	return salt, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	return e.EncryptBytes([]byte(plaintext))
}

// EncryptBytes encrypts plaintext bytes using AES-256-GCM and returns base64-encoded ciphertext
func (e *Encryptor) EncryptBytes(plaintext []byte) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64 for safe storage/transmission
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	plaintext, err := e.DecryptBytes(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// DecryptBytes decrypts base64-encoded ciphertext using AES-256-GCM and returns bytes
func (e *Encryptor) DecryptBytes(ciphertext string) ([]byte, error) {
	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	// Check minimum length (nonce + tag)
	if len(data) < NonceSize+16 { // 16 is GCM tag size
		return nil, ErrInvalidCiphertext
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Extract nonce and ciphertext
	nonce := data[:NonceSize]
	ciphertextData := data[NonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// Destroy securely clears the encryption key from memory
func (e *Encryptor) Destroy() {
	// Zero out the key
	for i := range e.key {
		e.key[i] = 0
	}
	e.key = nil
}

// EncryptCredentials is a convenience function to encrypt API credentials
func EncryptCredentials(clientID, clientSecret, accessToken string, key []byte) (map[string]string, error) {
	encryptor, err := NewEncryptor(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}
	defer encryptor.Destroy()

	result := make(map[string]string)

	if clientID != "" {
		encrypted, err := encryptor.Encrypt(clientID)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt client ID: %w", err)
		}
		result["client_id"] = encrypted
	}

	if clientSecret != "" {
		encrypted, err := encryptor.Encrypt(clientSecret)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt client secret: %w", err)
		}
		result["client_secret"] = encrypted
	}

	if accessToken != "" {
		encrypted, err := encryptor.Encrypt(accessToken)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt access token: %w", err)
		}
		result["access_token"] = encrypted
	}

	return result, nil
}

// DecryptCredentials is a convenience function to decrypt API credentials
func DecryptCredentials(encryptedCreds map[string]string, key []byte) (map[string]string, error) {
	encryptor, err := NewEncryptor(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}
	defer encryptor.Destroy()

	result := make(map[string]string)

	for field, encrypted := range encryptedCreds {
		if encrypted == "" {
			continue
		}

		decrypted, err := encryptor.Decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt %s: %w", field, err)
		}
		result[field] = decrypted
	}

	return result, nil
} 
package encryption

import (
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	if len(key) != AESKeySize {
		t.Errorf("Expected key size %d, got %d", AESKeySize, len(key))
	}

	// Generate another key and ensure they're different
	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate second key: %v", err)
	}

	if string(key) == string(key2) {
		t.Error("Generated keys should be different")
	}
}

func TestNewEncryptor(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	if encryptor == nil {
		t.Error("Encryptor should not be nil")
	}

	// Test with invalid key size
	invalidKey := make([]byte, 16) // Wrong size
	_, err = NewEncryptor(invalidKey)
	if err != ErrInvalidKeySize {
		t.Errorf("Expected ErrInvalidKeySize, got %v", err)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}
	defer encryptor.Destroy()

	testCases := []string{
		"",
		"hello world",
		"this is a test string",
		"special chars: !@#$%^&*()_+-=[]{}|;':\",./<>?",
		"unicode: 🔒🔑🛡️",
		"very long string: " + string(make([]byte, 1000)),
	}

	for _, plaintext := range testCases {
		t.Run("plaintext_"+plaintext[:min(10, len(plaintext))], func(t *testing.T) {
			// Encrypt
			ciphertext, err := encryptor.Encrypt(plaintext)
			if err != nil {
				t.Fatalf("Failed to encrypt: %v", err)
			}

			// Ciphertext should be different from plaintext
			if ciphertext == plaintext && plaintext != "" {
				t.Error("Ciphertext should be different from plaintext")
			}

			// Decrypt
			decrypted, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				t.Fatalf("Failed to decrypt: %v", err)
			}

			// Should match original
			if decrypted != plaintext {
				t.Errorf("Decrypted text doesn't match original. Expected: %s, Got: %s", plaintext, decrypted)
			}
		})
	}
}

func TestEncryptSameTextDifferentCiphertext(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}
	defer encryptor.Destroy()

	plaintext := "same text"
	ciphertext1, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt first time: %v", err)
	}

	ciphertext2, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt second time: %v", err)
	}

	// Should be different due to random nonce
	if ciphertext1 == ciphertext2 {
		t.Error("Same plaintext should produce different ciphertexts due to random nonce")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := encryptor.Decrypt(ciphertext1)
	if err != nil {
		t.Fatalf("Failed to decrypt first ciphertext: %v", err)
	}

	decrypted2, err := encryptor.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("Failed to decrypt second ciphertext: %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("Both ciphertexts should decrypt to original plaintext")
	}
}

func TestDecryptInvalidCiphertext(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}
	defer encryptor.Destroy()

	// Test various invalid inputs
	testCases := []struct {
		name       string
		ciphertext string
		expectErr  error
	}{
		{"empty", "", ErrInvalidCiphertext},
		{"too_short", "short", nil}, // Will fail at base64 decode or GCM
		{"invalid_base64", "invalid base64!", nil}, // Will fail at base64 decode
		{"wrong_key", "", nil}, // We'll encrypt with different key
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "wrong_key" {
				// Encrypt with a different key
				wrongKey, _ := GenerateKey()
				wrongEncryptor, _ := NewEncryptor(wrongKey)
				ciphertext, _ := wrongEncryptor.Encrypt("test")
				wrongEncryptor.Destroy()
				
				_, err := encryptor.Decrypt(ciphertext)
				if err != ErrDecryptionFailed {
					t.Errorf("Expected ErrDecryptionFailed, got %v", err)
				}
			} else {
				_, err := encryptor.Decrypt(tc.ciphertext)
				if err == nil {
					t.Error("Expected decryption to fail")
				}
			}
		})
	}
}

func TestEncryptCredentials(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	clientID := "test_client_id"
	clientSecret := "test_client_secret"
	accessToken := "test_access_token"

	encrypted, err := EncryptCredentials(clientID, clientSecret, accessToken, key)
	if err != nil {
		t.Fatalf("Failed to encrypt credentials: %v", err)
	}

	// Check that all fields are encrypted
	expectedFields := []string{"client_id", "client_secret", "access_token"}
	for _, field := range expectedFields {
		if _, exists := encrypted[field]; !exists {
			t.Errorf("Expected encrypted field %s not found", field)
		}
		if encrypted[field] == "" {
			t.Errorf("Encrypted field %s is empty", field)
		}
	}

	// Decrypt and verify
	decrypted, err := DecryptCredentials(encrypted, key)
	if err != nil {
		t.Fatalf("Failed to decrypt credentials: %v", err)
	}

	if decrypted["client_id"] != clientID {
		t.Errorf("Client ID mismatch. Expected: %s, Got: %s", clientID, decrypted["client_id"])
	}
	if decrypted["client_secret"] != clientSecret {
		t.Errorf("Client Secret mismatch. Expected: %s, Got: %s", clientSecret, decrypted["client_secret"])
	}
	if decrypted["access_token"] != accessToken {
		t.Errorf("Access Token mismatch. Expected: %s, Got: %s", accessToken, decrypted["access_token"])
	}
}

func TestEncryptorDestroy(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	// Test that it works before destroy
	_, err = encryptor.Encrypt("test")
	if err != nil {
		t.Fatalf("Encryption should work before destroy: %v", err)
	}

	// Destroy the encryptor
	encryptor.Destroy()

	// The key should be zeroed out
	if encryptor.key != nil {
		t.Error("Key should be nil after destroy")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
} 
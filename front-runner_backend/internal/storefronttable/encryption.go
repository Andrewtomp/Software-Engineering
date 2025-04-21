// front-runner/internal/storefronttable/encryption.go
package storefronttable // Correct package name

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var encryptionKey []byte // Loaded during Setup

// loadEncryptionKey retrieves the AES-256 encryption key from the
// STOREFRONT_KEY environment variable. The key must be a 32-byte
// value encoded in base64. It populates the package-level encryptionKey variable.
// It returns an error if the key is missing, cannot be decoded, or is not 32 bytes long.
func loadEncryptionKey() error {
	keyBase64 := strings.TrimSpace(string(os.Getenv("STOREFRONT_KEY")))

	if keyBase64 == "" {
		return fmt.Errorf("must provide storefront encryption key")
	}

	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64 key from file: %w", err)
	}

	// Ensure key length is suitable for AES (16, 24, or 32 bytes)
	// We'll enforce AES-256 (32 bytes) for strong security.
	if len(key) != 32 {
		return fmt.Errorf("decoded encryption key must be 32 bytes for AES-256, got %d bytes", len(key))
	}
	encryptionKey = key
	log.Printf("Storefront encryption key loaded successfully")
	return nil
}

// encryptCredentials encrypts the given plaintext string using AES-GCM with the
// package-level encryption key. It generates a unique nonce for each encryption.
// Returns a base64 encoded string containing the nonce prepended to the ciphertext.
// Returns an error if the encryption key is not loaded or if any cryptographic operation fails.
func encryptCredentials(plaintext string) (string, error) {
	if len(encryptionKey) == 0 {
		// This should ideally not happen if Setup is called correctly
		log.Println("Error: encryptCredentials called before encryption key was loaded.")
		return "", errors.New("encryption key not available")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		// Log internal error details
		log.Printf("Error creating cipher block during encryption: %v", err)
		return "", errors.New("internal error during encryption setup")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("Error creating GCM during encryption: %v", err)
		return "", errors.New("internal error during encryption setup")
	}

	// Allocate nonce space
	nonce := make([]byte, gcm.NonceSize())
	// Fill nonce with random data
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("Error generating nonce during encryption: %v", err)
		return "", errors.New("internal error during encryption")
	}

	// Seal encrypts the plaintext and prepends the nonce to the ciphertext output.
	// The nonce is passed in as the first argument, and also used for nonce generation internally.
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode the result (nonce + ciphertext) to base64 for safe storage/transfer
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptCredentials decrypts a base64 encoded string (containing nonce + ciphertext)
// using AES-GCM with the package-level encryption key.
// It verifies the integrity of the data during decryption.
// Returns the original plaintext string or an error if the key is not loaded,
// the input format is invalid, or decryption fails (e.g., due to tampering or incorrect key).
func decryptCredentials(ciphertextBase64 string) (string, error) {
	if len(encryptionKey) == 0 {
		log.Println("Error: decryptCredentials called before encryption key was loaded.")
		return "", errors.New("encryption key not available")
	}

	// Decode base64 string back to bytes
	data, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		// Log potentially invalid input format
		log.Printf("Error decoding base64 ciphertext: %v", err)
		return "", errors.New("invalid credentials format")
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		log.Printf("Error creating cipher block during decryption: %v", err)
		return "", errors.New("internal error during decryption setup")
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("Error creating GCM during decryption: %v", err)
		return "", errors.New("internal error during decryption setup")
	}

	nonceSize := gcm.NonceSize()
	// Ensure received data is at least as long as the nonce
	if len(data) < nonceSize {
		log.Println("Error: Ciphertext received is shorter than nonce size.")
		return "", errors.New("invalid credentials format")
	}

	// Extract nonce and actual ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt using gcm.Open
	// The first argument (dst) is usually nil to let Open allocate memory.
	// The nonce, ciphertext, and optional additional authenticated data (nil here) are provided.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// IMPORTANT: Decryption failure could be due to incorrect key OR tampered data.
		// Do NOT reveal specific crypto errors to the client.
		log.Printf("Failed to decrypt credentials (potential tampering or wrong key): %v", err)
		return "", errors.New("failed to decrypt credentials") // Generic error is safer
	}

	return string(plaintext), nil
}

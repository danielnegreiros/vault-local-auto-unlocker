package encryption

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test helper functions
func createTempDir(t *testing.T) string {
	dir, err := os.MkdirTemp("", "crypto_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir
}

func cleanupTempDir(t *testing.T, dir string) {
	if err := os.RemoveAll(dir); err != nil {
		t.Errorf("Failed to cleanup temp dir: %v", err)
	}
}

func createTestKeys(t *testing.T, dir string) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	privateKeyPath := filepath.Join(dir, "private.pem")
	publicKeyPath := filepath.Join(dir, "public.pem")

	// Save private key with correct implementation
	if err := saveTestPEMKey(privateKeyPath, privateKey); err != nil {
		t.Fatalf("Failed to save private key: %v", err)
	}
	// Save public key with correct implementation
	if err := saveTestPublicPEMKey(publicKeyPath, publicKey); err != nil {
		t.Fatalf("Failed to save public key: %v", err)
	}

	return privateKey, publicKey
}

// Fixed version of savePEMKey for testing
func saveTestPEMKey(fileName string, key *rsa.PrivateKey) error {
	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	var privateKey = &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}
	return pem.Encode(outFile, privateKey)
}

// Fixed version of savePublicPEMKey for testing
func saveTestPublicPEMKey(fileName string, pubkey *rsa.PublicKey) error {
	asn1Bytes := x509.MarshalPKCS1PublicKey(pubkey)
	var pemkey = &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}
	outFile, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return pem.Encode(outFile, pemkey)
}

// Test NewCrypto function
func TestNewCrypto(t *testing.T) {
	t.Run("NewCrypto with new keys", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		if crypto == nil {
			t.Fatal("NewCrypto returned nil crypto")
		}
		if crypto.path != dir {
			t.Errorf("Expected path %s, got %s", dir, crypto.path)
		}
		if crypto.privateKey == nil {
			t.Error("Private key should not be nil")
		}
		if crypto.publicKey == nil {
			t.Error("Public key should not be nil")
		}

		// Verify files were created
		privateKeyPath := filepath.Join(dir, "private.pem")
		publicKeyPath := filepath.Join(dir, "public.pem")
		if !fileExists(privateKeyPath) {
			t.Error("Private key file was not created")
		}
		if !fileExists(publicKeyPath) {
			t.Error("Public key file was not created")
		}
	})

	t.Run("NewCrypto with existing keys", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		// Create test keys first
		expectedPrivateKey, expectedPublicKey := createTestKeys(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		if crypto.privateKey.Equal(expectedPrivateKey) == false {
			t.Error("Private key doesn't match expected key")
		}
		if crypto.publicKey.Equal(expectedPublicKey) == false {
			t.Error("Public key doesn't match expected key")
		}
	})

	t.Run("NewCrypto with invalid path", func(t *testing.T) {
		// Use a path that doesn't exist and can't be created
		invalidPath := "/root/nonexistent/path"
		_, err := NewCrypto(invalidPath)
		if err == nil {
			t.Error("Expected error for invalid path, got nil")
		}
	})
}

// Test Encrypt function
func TestCrypto_Encrypt(t *testing.T) {
	t.Run("Successful encryption", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		testText := "Hello, World!"
		encrypted, err := crypto.Encrypt(testText)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		if encrypted == "" {
			t.Error("Encrypted text should not be empty")
		}

		// Verify it's valid base64
		_, err = base64.StdEncoding.DecodeString(encrypted)
		if err != nil {
			t.Errorf("Encrypted text is not valid base64: %v", err)
		}
	})

	t.Run("Encryption with empty string", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		encrypted, err := crypto.Encrypt("")
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		if encrypted == "" {
			t.Error("Encrypted text should not be empty even for empty input")
		}
	})

	t.Run("Encryption without public key", func(t *testing.T) {
		crypto := &Crypto{
			path:       "/tmp",
			publicKey:  nil,
			privateKey: nil,
		}

		_, err := crypto.Encrypt("test")
		if err == nil {
			t.Error("Expected error when public key is nil")
		}
		if !strings.Contains(err.Error(), "public key is not loaded") {
			t.Errorf("Expected 'public key is not loaded' error, got: %v", err)
		}
	})

	t.Run("Encryption with large text", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		// Create text that's close to RSA limit (2048 bit key can encrypt up to 245 bytes)
		largeText := strings.Repeat("A", 200)
		encrypted, err := crypto.Encrypt(largeText)
		if err != nil {
			t.Fatalf("Encrypt failed for large text: %v", err)
		}

		if encrypted == "" {
			t.Error("Encrypted text should not be empty")
		}
	})

	t.Run("Encryption with text too large for RSA", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		// Text larger than RSA can handle
		tooLargeText := strings.Repeat("A", 300)
		_, err = crypto.Encrypt(tooLargeText)
		if err == nil {
			t.Error("Expected error for text too large for RSA encryption")
		}
	})
}

// Test Decrypt function
func TestCrypto_Decrypt(t *testing.T) {
	t.Run("Successful decryption", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		testText := "Hello, World!"
		encrypted, err := crypto.Encrypt(testText)
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		decrypted, err := crypto.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("Decrypt failed: %v", err)
		}

		if decrypted != testText {
			t.Errorf("Expected %s, got %s", testText, decrypted)
		}
	})

	t.Run("Decryption without private key", func(t *testing.T) {
		crypto := &Crypto{
			path:       "/tmp",
			publicKey:  nil,
			privateKey: nil,
		}

		_, err := crypto.Decrypt("test")
		if err == nil {
			t.Error("Expected error when private key is nil")
		}
		if !strings.Contains(err.Error(), "private key is not loaded") {
			t.Errorf("Expected 'private key is not loaded' error, got: %v", err)
		}
	})

	t.Run("Decryption with invalid base64", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		_, err = crypto.Decrypt("invalid-base64!")
		if err == nil {
			t.Error("Expected error for invalid base64")
		}
	})

	t.Run("Decryption with invalid encrypted data", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto, err := NewCrypto(dir)
		if err != nil {
			t.Fatalf("NewCrypto failed: %v", err)
		}

		// Valid base64 but not valid encrypted data
		invalidEncrypted := base64.StdEncoding.EncodeToString([]byte("invalid encrypted data"))
		_, err = crypto.Decrypt(invalidEncrypted)
		if err == nil {
			t.Error("Expected error for invalid encrypted data")
		}
	})
}

// Test round-trip encryption/decryption
func TestCrypto_EncryptDecryptRoundTrip(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	crypto, err := NewCrypto(dir)
	if err != nil {
		t.Fatalf("NewCrypto failed: %v", err)
	}

	testCases := []string{
		"",
		"Hello, World!",
		"Special characters: !@#$%^&*()_+-=[]{}|;:,.<>?",
		"Unicode: 你好世界 🌍",
		strings.Repeat("A", 100),
	}

	for _, testText := range testCases {
		t.Run("RoundTrip_"+testText[:min(len(testText), 20)], func(t *testing.T) {
			encrypted, err := crypto.Encrypt(testText)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			decrypted, err := crypto.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != testText {
				t.Errorf("Round-trip failed: expected %q, got %q", testText, decrypted)
			}
		})
	}
}

// Test loadOrCreateKeys function
func TestCrypto_loadOrCreateKeys(t *testing.T) {
	t.Run("Create new keys when none exist", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto := &Crypto{path: dir}
		err := crypto.loadOrCreateKeys()
		if err != nil {
			t.Fatalf("loadOrCreateKeys failed: %v", err)
		}

		if crypto.privateKey == nil {
			t.Error("Private key should not be nil")
		}
		if crypto.publicKey == nil {
			t.Error("Public key should not be nil")
		}

		privateKeyPath := filepath.Join(dir, "private.pem")
		publicKeyPath := filepath.Join(dir, "public.pem")
		if !fileExists(privateKeyPath) {
			t.Error("Private key file was not created")
		}
		if !fileExists(publicKeyPath) {
			t.Error("Public key file was not created")
		}
	})

	t.Run("Load existing keys", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		expectedPrivateKey, expectedPublicKey := createTestKeys(t, dir)

		crypto := &Crypto{path: dir}
		err := crypto.loadOrCreateKeys()
		if err != nil {
			t.Fatalf("loadOrCreateKeys failed: %v", err)
		}

		if !crypto.privateKey.Equal(expectedPrivateKey) {
			t.Error("Private key doesn't match expected key")
		}
		if !crypto.publicKey.Equal(expectedPublicKey) {
			t.Error("Public key doesn't match expected key")
		}
	})
}

// Test loadKeys function
func TestCrypto_loadKeys(t *testing.T) {
	t.Run("Load valid keys", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		expectedPrivateKey, expectedPublicKey := createTestKeys(t, dir)

		crypto := &Crypto{path: dir}
		privateKeyPath := filepath.Join(dir, "private.pem")
		publicKeyPath := filepath.Join(dir, "public.pem")

		err := crypto.loadKeys(privateKeyPath, publicKeyPath)
		if err != nil {
			t.Fatalf("loadKeys failed: %v", err)
		}

		if !crypto.privateKey.Equal(expectedPrivateKey) {
			t.Error("Private key doesn't match expected key")
		}
		if !crypto.publicKey.Equal(expectedPublicKey) {
			t.Error("Public key doesn't match expected key")
		}
	})

	t.Run("Load with non-existent private key", func(t *testing.T) {
		dir := createTempDir(t)
		defer cleanupTempDir(t, dir)

		crypto := &Crypto{path: dir}
		err := crypto.loadKeys(filepath.Join(dir, "nonexistent.pem"), filepath.Join(dir, "public.pem"))
		if err == nil {
			t.Error("Expected error for non-existent private key")
		}
	})
}

// Test savePEMKey function
func TestSavePEMKey(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	keyPath := filepath.Join(dir, "test_private.pem")
	err = saveTestPEMKey(keyPath, privateKey)
	if err != nil {
		t.Fatalf("savePEMKey failed: %v", err)
	}

	if !fileExists(keyPath) {
		t.Error("Key file was not created")
	}

	// Verify the saved key can be loaded
	data, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read saved key: %v", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		t.Error("Invalid PEM format or type")
	}

	loadedKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse saved key: %v", err)
	}

	if !loadedKey.Equal(privateKey) {
		t.Error("Saved key doesn't match original")
	}
}

// Test savePublicPEMKey function
func TestSavePublicPEMKey(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	keyPath := filepath.Join(dir, "test_public.pem")
	err = saveTestPublicPEMKey(keyPath, publicKey)
	if err != nil {
		t.Fatalf("savePublicPEMKey failed: %v", err)
	}

	if !fileExists(keyPath) {
		t.Error("Key file was not created")
	}

	// Verify the saved key can be loaded
	data, err := os.ReadFile(keyPath)
	if err != nil {
		t.Fatalf("Failed to read saved key: %v", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		t.Error("Invalid PEM format or type")
	}

	loadedKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse saved key: %v", err)
	}

	if !loadedKey.Equal(publicKey) {
		t.Error("Saved key doesn't match original")
	}
}

// Test fileExists function
func TestFileExists(t *testing.T) {
	dir := createTempDir(t)
	defer cleanupTempDir(t, dir)

	t.Run("Existing file", func(t *testing.T) {
		testFile := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if !fileExists(testFile) {
			t.Error("fileExists should return true for existing file")
		}
	})

	t.Run("Non-existent file", func(t *testing.T) {
		nonExistentFile := filepath.Join(dir, "nonexistent.txt")
		if fileExists(nonExistentFile) {
			t.Error("fileExists should return false for non-existent file")
		}
	})

	t.Run("Directory instead of file", func(t *testing.T) {
		if fileExists(dir) {
			t.Error("fileExists should return false for directories")
		}
	})
}

// Helper function for Go versions that don't have min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark tests
func BenchmarkCrypto_Encrypt(b *testing.B) {
	dir := createTempDir(&testing.T{})
	defer cleanupTempDir(&testing.T{}, dir)

	crypto, err := NewCrypto(dir)
	if err != nil {
		b.Fatalf("NewCrypto failed: %v", err)
	}

	testText := "Hello, World! This is a test message for encryption."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.Encrypt(testText)
		if err != nil {
			b.Fatalf("Encrypt failed: %v", err)
		}
	}
}

func BenchmarkCrypto_Decrypt(b *testing.B) {
	dir := createTempDir(&testing.T{})
	defer cleanupTempDir(&testing.T{}, dir)

	crypto, err := NewCrypto(dir)
	if err != nil {
		b.Fatalf("NewCrypto failed: %v", err)
	}

	testText := "Hello, World! This is a test message for encryption."
	encrypted, err := crypto.Encrypt(testText)
	if err != nil {
		b.Fatalf("Encrypt failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.Decrypt(encrypted)
		if err != nil {
			b.Fatalf("Decrypt failed: %v", err)
		}
	}
}
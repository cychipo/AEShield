package encryption

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/aeshield/backend/internal/crypto"
)

func TestNewEncryptionService(t *testing.T) {
	service := NewEncryptionService()
	if service == nil {
		t.Fatal("NewEncryptionService() returned nil")
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	service := NewEncryptionService()
	ctx := context.Background()

	testCases := []struct {
		name      string
		plaintext string
		password  string
		keyBits   crypto.KeyBits
	}{
		{
			name:      "AES-128 roundtrip",
			plaintext: "hello aeshield 128",
			password:  "password123",
			keyBits:   crypto.KeyBits128,
		},
		{
			name:      "AES-192 roundtrip",
			plaintext: "hello aeshield 192",
			password:  "password123",
			keyBits:   crypto.KeyBits192,
		},
		{
			name:      "AES-256 roundtrip (default)",
			plaintext: "hello aeshield 256",
			password:  "password123",
			keyBits:   crypto.KeyBits256,
		},
		{
			name:      "Empty payload",
			plaintext: "",
			password:  "s3cr3t!",
			keyBits:   crypto.KeyBits256,
		},
		{
			name:      "Large payload",
			plaintext: strings.Repeat("A", crypto.ChunkSize*2+100),
			password:  "complexPassword123",
			keyBits:   crypto.KeyBits256,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := strings.NewReader(tc.plaintext)
			var encrypted bytes.Buffer

			encryptParams := EncryptParams{
				Source:      src,
				Destination: &encrypted,
				Password:    tc.password,
				KeyBits:     tc.keyBits,
			}

			_, err := service.Encrypt(ctx, encryptParams)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			var decrypted bytes.Buffer
			decryptParams := DecryptParams{
				Source:      &encrypted,
				Destination: &decrypted,
				Password:    tc.password,
			}

			err = service.Decrypt(ctx, decryptParams)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted.String() != tc.plaintext {
				t.Fatalf("roundtrip mismatch:\n got  %q\n want %q", decrypted.String(), tc.plaintext)
			}
		})
	}
}

func TestEncryptValidations(t *testing.T) {
	service := NewEncryptionService()
	ctx := context.Background()

	t.Run("encrypt with nil source", func(t *testing.T) {
		var dest bytes.Buffer
		params := EncryptParams{
			Source:      nil,
			Destination: &dest,
			Password:    "password",
			KeyBits:     crypto.KeyBits256,
		}

		_, err := service.Encrypt(ctx, params)
		if err == nil {
			t.Fatal("expected error for nil source, got nil")
		}
	})

	t.Run("encrypt with nil destination", func(t *testing.T) {
		src := strings.NewReader("test")
		params := EncryptParams{
			Source:      src,
			Destination: nil,
			Password:    "password",
			KeyBits:     crypto.KeyBits256,
		}

		_, err := service.Encrypt(ctx, params)
		if err == nil {
			t.Fatal("expected error for nil destination, got nil")
		}
	})

	t.Run("encrypt with empty password", func(t *testing.T) {
		var src, dest bytes.Buffer
		params := EncryptParams{
			Source:      &src,
			Destination: &dest,
			Password:    "",
			KeyBits:     crypto.KeyBits256,
		}

		_, err := service.Encrypt(ctx, params)
		if err == nil {
			t.Fatal("expected error for empty password, got nil")
		}
	})

	t.Run("encrypt with invalid key bits", func(t *testing.T) {
		var src, dest bytes.Buffer
		params := EncryptParams{
			Source:      &src,
			Destination: &dest,
			Password:    "password",
			KeyBits:     crypto.KeyBits(99), // Invalid
		}

		_, err := service.Encrypt(ctx, params)
		if err == nil {
			t.Fatal("expected error for invalid key bits, got nil")
		}
	})
}

func TestDecryptValidations(t *testing.T) {
	service := NewEncryptionService()
	ctx := context.Background()

	t.Run("decrypt with nil source", func(t *testing.T) {
		var dest bytes.Buffer
		params := DecryptParams{
			Source:      nil,
			Destination: &dest,
			Password:    "password",
		}

		err := service.Decrypt(ctx, params)
		if err == nil {
			t.Fatal("expected error for nil source, got nil")
		}
	})

	t.Run("decrypt with nil destination", func(t *testing.T) {
		src := strings.NewReader("test")
		params := DecryptParams{
			Source:      src,
			Destination: nil,
			Password:    "password",
		}

		err := service.Decrypt(ctx, params)
		if err == nil {
			t.Fatal("expected error for nil destination, got nil")
		}
	})

	t.Run("decrypt with empty password", func(t *testing.T) {
		var src, dest bytes.Buffer
		params := DecryptParams{
			Source:      &src,
			Destination: &dest,
			Password:    "",
		}

		err := service.Decrypt(ctx, params)
		if err == nil {
			t.Fatal("expected error for empty password, got nil")
		}
	})
}

func TestDeriveKey(t *testing.T) {
	service := NewEncryptionService()
	ctx := context.Background()

	salt := []byte("1234567890123456") // 16 bytes

	testCases := []struct {
		name     string
		password string
		keyBits  crypto.KeyBits
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "derive AES-128 key",
			password: "password123",
			keyBits:  crypto.KeyBits128,
			wantLen:  16,
			wantErr:  false,
		},
		{
			name:     "derive AES-192 key",
			password: "password123",
			keyBits:  crypto.KeyBits192,
			wantLen:  24,
			wantErr:  false,
		},
		{
			name:     "derive AES-256 key",
			password: "password128",
			keyBits:  crypto.KeyBits256,
			wantLen:  32,
			wantErr:  false,
		},
		{
			name:     "invalid key bits",
			password: "password123",
			keyBits:  crypto.KeyBits(99),
			wantLen:  0,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			key, err := service.DeriveKey(ctx, tc.password, salt, tc.keyBits)

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("DeriveKey() error = %v", err)
			}

			if len(key) != tc.wantLen {
				t.Errorf("DeriveKey() returned key of length %d, want %d", len(key), tc.wantLen)
			}

			key2, err := service.DeriveKey(ctx, tc.password, salt, tc.keyBits)
			if err != nil {
				t.Fatalf("DeriveKey() second call error = %v", err)
			}

			if !bytes.Equal(key, key2) {
				t.Error("DeriveKey() should produce consistent results for same inputs")
			}

			key3, err := service.DeriveKey(ctx, tc.password+"different", salt, tc.keyBits)
			if err != nil {
				t.Fatalf("DeriveKey() different password error = %v", err)
			}

			if bytes.Equal(key, key3) {
				t.Error("DeriveKey() should produce different results for different passwords")
			}
		})
	}
}

func TestGetters(t *testing.T) {
	service := NewEncryptionService()

	t.Run("GetDefaultKeyBits should return AES-256", func(t *testing.T) {
		defaultBits := service.GetDefaultKeyBits()
		if defaultBits != crypto.KeyBits256 {
			t.Errorf("GetDefaultKeyBits() = %d, want %d", int(defaultBits), int(crypto.KeyBits256))
		}
	})

	t.Run("GetSupportedKeyBits should return all supported types", func(t *testing.T) {
		supported := service.GetSupportedKeyBits()
		expectedCount := 3 // 128, 192, 256

		if len(supported) != expectedCount {
			t.Errorf("GetSupportedKeyBits() returned %d items, want %d", len(supported), expectedCount)
		}

		has128, has192, has256 := false, false, false
		for _, bits := range supported {
			switch bits {
			case crypto.KeyBits128:
				has128 = true
			case crypto.KeyBits192:
				has192 = true
			case crypto.KeyBits256:
				has256 = true
			}
		}

		if !has128 || !has192 || !has256 {
			t.Errorf("GetSupportedKeyBits() missing expected values: 128=%t, 192=%t, 256=%t", has128, has192, has256)
		}
	})
}

func TestValidateEncryptedFile(t *testing.T) {
	service := NewEncryptionService()
	ctx := context.Background()

	t.Run("valid encrypted file", func(t *testing.T) {
		var src bytes.Buffer
		src.WriteString("test data")
		var encrypted bytes.Buffer

		serviceInner := NewEncryptionService()
		params := EncryptParams{
			Source:      &src,
			Destination: &encrypted,
			Password:    "testpass",
			KeyBits:     crypto.KeyBits256,
		}

		_, err := serviceInner.Encrypt(ctx, params)
		if err != nil {
			t.Fatalf("Failed to create test encrypted file: %v", err)
		}

		reader := bytes.NewReader(encrypted.Bytes())
		isValid, keyBits, err := service.ValidateEncryptedFile(ctx, reader)
		if err != nil {
			t.Fatalf("ValidateEncryptedFile() error = %v", err)
		}

		if !isValid {
			t.Error("ValidateEncryptedFile() returned false for valid file")
		}

		if keyBits != crypto.KeyBits256 {
			t.Errorf("ValidateEncryptedFile() returned key bits %d, want %d", int(keyBits), int(crypto.KeyBits256))
		}
	})

	t.Run("invalid file (not an AEShield file)", func(t *testing.T) {
		plainData := strings.NewReader("not encrypted data")
		isValid, keyBits, err := service.ValidateEncryptedFile(ctx, plainData)
		if err != nil {
			t.Fatalf("ValidateEncryptedFile() error = %v", err)
		}

		if isValid {
			t.Error("ValidateEncryptedFile() returned true for non-encrypted file")
		}

		if keyBits != 0 {
			t.Errorf("ValidateEncryptedFile() returned key bits %d for invalid file, want 0", int(keyBits))
		}
	})

	t.Run("short file", func(t *testing.T) {
		shortData := strings.NewReader("short")
		isValid, keyBits, err := service.ValidateEncryptedFile(ctx, shortData)
		if err != nil {
			t.Fatalf("ValidateEncryptedFile() error = %v", err)
		}

		if isValid {
			t.Error("ValidateEncryptedFile() returned true for short file")
		}

		if keyBits != 0 {
			t.Errorf("ValidateEncryptedFile() returned key bits %d for short file, want 0", int(keyBits))
		}
	})
}
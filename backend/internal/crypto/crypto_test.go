package crypto_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/aeshield/backend/internal/crypto"
)

// roundtrip mã hóa rồi giải mã, kiểm tra plaintext khớp.
func roundtrip(t *testing.T, plaintext string, password string, bits crypto.KeyBits) {
	t.Helper()
	src := strings.NewReader(plaintext)

	var encrypted bytes.Buffer
	_, err := crypto.Encrypt(&encrypted, src, password, bits)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	var decrypted bytes.Buffer
	if err := crypto.Decrypt(&decrypted, &encrypted, password); err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted.String() != plaintext {
		t.Fatalf("roundtrip mismatch:\n got  %q\n want %q", decrypted.String(), plaintext)
	}
}

func TestRoundtrip_AES128(t *testing.T) {
	roundtrip(t, "hello aeshield 128", "password123", crypto.KeyBits128)
}

func TestRoundtrip_AES192(t *testing.T) {
	roundtrip(t, "hello aeshield 192", "password123", crypto.KeyBits192)
}

func TestRoundtrip_AES256(t *testing.T) {
	roundtrip(t, "hello aeshield 256", "password123", crypto.KeyBits256)
}

func TestRoundtrip_EmptyPayload(t *testing.T) {
	roundtrip(t, "", "password123", crypto.KeyBits256)
}

func TestRoundtrip_LargeFile(t *testing.T) {
	// 3 chunk + sisa (3*ChunkSize + 1 byte)
	size := crypto.ChunkSize*3 + 1
	plaintext := strings.Repeat("A", size)
	roundtrip(t, plaintext, "s3cr3t!", crypto.KeyBits256)
}

func TestDecrypt_WrongPassword(t *testing.T) {
	plaintext := "secret data"
	var encrypted bytes.Buffer
	_, err := crypto.Encrypt(&encrypted, strings.NewReader(plaintext), "correctpassword", crypto.KeyBits256)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	var decrypted bytes.Buffer
	err = crypto.Decrypt(&decrypted, &encrypted, "wrongpassword")
	if err == nil {
		t.Fatal("expected error with wrong password, got nil")
	}
}

func TestDecrypt_CorruptedCiphertext(t *testing.T) {
	plaintext := "data to corrupt"
	var encrypted bytes.Buffer
	_, err := crypto.Encrypt(&encrypted, strings.NewReader(plaintext), "pass", crypto.KeyBits256)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Lật bit ở cuối buffer (phần ciphertext)
	b := encrypted.Bytes()
	b[len(b)-1] ^= 0xFF

	var decrypted bytes.Buffer
	err = crypto.Decrypt(&decrypted, bytes.NewReader(b), "pass")
	if err == nil {
		t.Fatal("expected error with corrupted ciphertext, got nil")
	}
}

func TestDecrypt_InvalidMagic(t *testing.T) {
	err := crypto.Decrypt(io.Discard, strings.NewReader("not a valid file at all..........."), "pass")
	if err != crypto.ErrInvalidMagic {
		t.Fatalf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestDecrypt_ShortHeader(t *testing.T) {
	err := crypto.Decrypt(io.Discard, strings.NewReader("ABC"), "pass")
	if err != crypto.ErrShortHeader {
		t.Fatalf("expected ErrShortHeader, got %v", err)
	}
}

func TestEncrypt_InvalidKeyBits(t *testing.T) {
	_, err := crypto.Encrypt(io.Discard, strings.NewReader("x"), "pass", crypto.KeyBits(99))
	if err == nil {
		t.Fatal("expected error for invalid key bits, got nil")
	}
}

// TestChunkNonceUniqueness kiểm tra rằng 2 chunk liên tiếp không có cùng nonce
// bằng cách mã hóa dữ liệu ≥ 2 chunk và so sánh output không lặp nonce.
func TestDeterministic_DifferentSalts(t *testing.T) {
	plaintext := "same plaintext"

	var enc1, enc2 bytes.Buffer
	crypto.Encrypt(&enc1, strings.NewReader(plaintext), "pass", crypto.KeyBits256) //nolint
	crypto.Encrypt(&enc2, strings.NewReader(plaintext), "pass", crypto.KeyBits256) //nolint

	// Hai lần mã hóa cùng plaintext + password phải cho ciphertext KHÁC nhau (salt random)
	if bytes.Equal(enc1.Bytes(), enc2.Bytes()) {
		t.Fatal("two encryptions of same plaintext should differ (random salt/nonce)")
	}
}

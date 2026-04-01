package crypto

import (
	"bytes"
	"encoding/hex"
	"io"
	"strings"
	"testing"

)

// roundtrip mã hóa rồi giải mã, kiểm tra plaintext khớp.
func roundtrip(t *testing.T, plaintext string, password string, bits KeyBits) {
	t.Helper()
	src := strings.NewReader(plaintext)

	var encrypted bytes.Buffer
	_, err := Encrypt(&encrypted, src, password, bits)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	var decrypted bytes.Buffer
	if err := Decrypt(&decrypted, &encrypted, password); err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if decrypted.String() != plaintext {
		t.Fatalf("roundtrip mismatch:\n got  %q\n want %q", decrypted.String(), plaintext)
	}
}

func TestRoundtrip_AES128(t *testing.T) {
	roundtrip(t, "hello aeshield 128", "password123", KeyBits128)
}

func TestRoundtrip_AES192(t *testing.T) {
	roundtrip(t, "hello aeshield 192", "password123", KeyBits192)
}

func TestRoundtrip_AES256(t *testing.T) {
	roundtrip(t, "hello aeshield 256", "password123", KeyBits256)
}

func TestRoundtrip_EmptyPayload(t *testing.T) {
	roundtrip(t, "", "password123", KeyBits256)
}

func TestRoundtrip_LargeFile(t *testing.T) {
	// 3 chunk + sisa (3*ChunkSize + 1 byte)
	size := ChunkSize*3 + 1
	plaintext := strings.Repeat("A", size)
	roundtrip(t, plaintext, "s3cr3t!", KeyBits256)
}

func TestDecrypt_WrongPassword(t *testing.T) {
	plaintext := "secret data"
	var encrypted bytes.Buffer
	_, err := Encrypt(&encrypted, strings.NewReader(plaintext), "correctpassword", KeyBits256)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	var decrypted bytes.Buffer
	err = Decrypt(&decrypted, &encrypted, "wrongpassword")
	if err == nil {
		t.Fatal("expected error with wrong password, got nil")
	}
}

func TestDecrypt_CorruptedCiphertext(t *testing.T) {
	plaintext := "data to corrupt"
	var encrypted bytes.Buffer
	_, err := Encrypt(&encrypted, strings.NewReader(plaintext), "pass", KeyBits256)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Lật bit ở cuối buffer (phần ciphertext)
	b := encrypted.Bytes()
	b[len(b)-1] ^= 0xFF

	var decrypted bytes.Buffer
	err = Decrypt(&decrypted, bytes.NewReader(b), "pass")
	if err == nil {
		t.Fatal("expected error with corrupted ciphertext, got nil")
	}
}

func TestDecrypt_InvalidMagic(t *testing.T) {
	err := Decrypt(io.Discard, strings.NewReader("not a valid file at all..........."), "pass")
	if err != ErrInvalidMagic {
		t.Fatalf("expected ErrInvalidMagic, got %v", err)
	}
}

func TestDecrypt_ShortHeader(t *testing.T) {
	err := Decrypt(io.Discard, strings.NewReader("ABC"), "pass")
	if err != ErrShortHeader {
		t.Fatalf("expected ErrShortHeader, got %v", err)
	}
}

func TestEncrypt_InvalidKeyBits(t *testing.T) {
	_, err := Encrypt(io.Discard, strings.NewReader("x"), "pass", KeyBits(99))
	if err == nil {
		t.Fatal("expected error for invalid key bits, got nil")
	}
}

// TestChunkNonceUniqueness kiểm tra rằng 2 chunk liên tiếp không có cùng nonce
// bằng cách mã hóa dữ liệu ≥ 2 chunk và so sánh output không lặp nonce.
func TestDeterministic_DifferentSalts(t *testing.T) {
	plaintext := "same plaintext"

	var enc1, enc2 bytes.Buffer
	Encrypt(&enc1, strings.NewReader(plaintext), "pass", KeyBits256) //nolint
	Encrypt(&enc2, strings.NewReader(plaintext), "pass", KeyBits256) //nolint

	// Hai lần mã hóa cùng plaintext + password phải cho ciphertext KHÁC nhau (salt random)
	if bytes.Equal(enc1.Bytes(), enc2.Bytes()) {
		t.Fatal("two encryptions of same plaintext should differ (random salt/nonce)")
	}
}

func TestEncrypt_OutputSizes(t *testing.T) {
	t.Run("empty payload writes header only", func(t *testing.T) {
		var encrypted bytes.Buffer
		written, err := Encrypt(&encrypted, strings.NewReader(""), "pass", KeyBits256)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
		if written != 33 || encrypted.Len() != 33 {
			t.Fatalf("expected 33-byte output, got written=%d len=%d", written, encrypted.Len())
		}
		if string(encrypted.Bytes()[:4]) != HeaderMagic {
			t.Fatalf("expected header magic %q", HeaderMagic)
		}
	})

	t.Run("one byte payload has header length and tag overhead", func(t *testing.T) {
		var encrypted bytes.Buffer
		written, err := Encrypt(&encrypted, strings.NewReader("A"), "pass", KeyBits256)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
		const want = 33 + 4 + 1 + 16
		if written != want || encrypted.Len() != want {
			t.Fatalf("expected %d-byte output, got written=%d len=%d", want, written, encrypted.Len())
		}
	})

	t.Run("single full chunk keeps expected framing", func(t *testing.T) {
		plaintext := strings.Repeat("A", ChunkSize)
		var encrypted bytes.Buffer
		written, err := Encrypt(&encrypted, strings.NewReader(plaintext), "pass", KeyBits256)
		if err != nil {
			t.Fatalf("Encrypt() error = %v", err)
		}
		want := int64(33 + 4 + ChunkSize + 16)
		if written != want || int64(encrypted.Len()) != want {
			t.Fatalf("expected %d-byte output, got written=%d len=%d", want, written, encrypted.Len())
		}
	})
}

func TestAESBlock_KnownVectors(t *testing.T) {
	tests := []struct {
		name string
		key  string
		src  string
		want string
	}{
		{
			name: "AES-128",
			key:  "000102030405060708090a0b0c0d0e0f",
			src:  "00112233445566778899aabbccddeeff",
			want: "69c4e0d86a7b0430d8cdb78070b4c55a",
		},
		{
			name: "AES-192",
			key:  "000102030405060708090a0b0c0d0e0f1011121314151617",
			src:  "00112233445566778899aabbccddeeff",
			want: "dda97ca4864cdfe06eaf70a0ec0d7191",
		},
		{
			name: "AES-256",
			key:  "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f",
			src:  "00112233445566778899aabbccddeeff",
			want: "8ea2b7ca516745bfeafc49904b496089",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := mustHex(t, tt.key)
			src := mustHex(t, tt.src)
			want := mustHex(t, tt.want)

			block, err := newAESBlock(key)
			if err != nil {
				t.Fatalf("new block error = %v", err)
			}

			got := make([]byte, 16)
			block.Encrypt(got, src)
			if !bytes.Equal(got, want) {
				t.Fatalf("ciphertext mismatch: got %x want %x", got, want)
			}
		})
	}
}

func TestGCM_KnownVector(t *testing.T) {
	key := mustHex(t, "00000000000000000000000000000000")
	nonce := mustHex(t, "000000000000000000000000")
	plaintext := mustHex(t, "00000000000000000000000000000000")
	want := mustHex(t, "0388dace60b6a392f328c2b971b2fe78ab6e47d42cec13bdf53a67b21257bddf")

	block, err := newAESBlock(key)
	if err != nil {
		t.Fatalf("new block error = %v", err)
	}
	gcm, err := newGCM(block)
	if err != nil {
		t.Fatalf("new gcm error = %v", err)
	}

	got := gcm.Seal(nil, nonce, plaintext, nil)
	if !bytes.Equal(got, want) {
		t.Fatalf("gcm mismatch: got %x want %x", got, want)
	}

	opened, err := gcm.Open(nil, nonce, got, nil)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if !bytes.Equal(opened, plaintext) {
		t.Fatalf("plaintext mismatch: got %x want %x", opened, plaintext)
	}
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("invalid hex %q: %v", s, err)
	}
	return b
}

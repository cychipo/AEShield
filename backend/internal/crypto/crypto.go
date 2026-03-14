// Package crypto cung cấp mã hóa/giải mã file streaming dùng AES-GCM + Argon2id.
//
// Định dạng ciphertext (prefix + data):
//
//	[4 bytes: magic "AES\x00"] [1 byte: key-bits / 64] [16 bytes: argon2 salt]
//	[12 bytes: GCM nonce] [N bytes: AES-GCM ciphertext+tag]
//
// Streaming được thực hiện theo từng chunk CHUNK_SIZE (64 KiB) để không tràn RAM.
// Mỗi chunk được mã hóa độc lập với nonce riêng (nonce ban đầu + index chunk).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// ChunkSize là kích thước mỗi khối plaintext khi mã hóa streaming (64 KiB)
	ChunkSize = 64 * 1024

	// HeaderMagic giúp nhận dạng file đã mã hóa bởi AEShield
	HeaderMagic = "AES\x00"

	saltSize  = 16
	nonceSize = 12
)

// Argon2id parameters (OWASP recommended minimum)
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
)

var (
	ErrInvalidMagic   = errors.New("crypto: invalid file magic, not an AEShield encrypted file")
	ErrInvalidKeyBits = errors.New("crypto: unsupported key size")
	ErrShortHeader    = errors.New("crypto: file header too short")
	ErrDecryptChunk   = errors.New("crypto: failed to decrypt chunk")
)

// KeyBits là số bit của khóa AES.
type KeyBits int

const (
	KeyBits128 KeyBits = 128
	KeyBits192 KeyBits = 192
	KeyBits256 KeyBits = 256
)

// DeriveKey dùng Argon2id để chuyển đổi password + salt thành AES key.
func DeriveKey(password string, salt []byte, bits KeyBits) ([]byte, error) {
	keyLen, err := keyLen(bits)
	if err != nil {
		return nil, err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, uint32(keyLen))
	return key, nil
}

// Encrypt đọc plaintext từ src, mã hóa streaming AES-GCM, ghi vào dst.
// Trả về số byte đã ghi vào dst.
func Encrypt(dst io.Writer, src io.Reader, password string, bits KeyBits) (int64, error) {
	if _, err := keyLen(bits); err != nil {
		return 0, err
	}

	// Sinh salt ngẫu nhiên
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return 0, fmt.Errorf("crypto: generate salt: %w", err)
	}

	// Sinh nonce gốc ngẫu nhiên
	baseNonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, baseNonce); err != nil {
		return 0, fmt.Errorf("crypto: generate nonce: %w", err)
	}

	key, err := DeriveKey(password, salt, bits)
	if err != nil {
		return 0, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, fmt.Errorf("crypto: new GCM: %w", err)
	}

	// Ghi header: magic(4) + keyBitsByte(1) + salt(16) + baseNonce(12)
	header := make([]byte, 0, 4+1+saltSize+nonceSize)
	header = append(header, []byte(HeaderMagic)...)
	header = append(header, byte(int(bits)/64))
	header = append(header, salt...)
	header = append(header, baseNonce...)

	written, err := dst.Write(header)
	if err != nil {
		return int64(written), fmt.Errorf("crypto: write header: %w", err)
	}
	total := int64(written)

	// Mã hóa từng chunk
	buf := make([]byte, ChunkSize)
	var chunkIdx uint64 = 0
	for {
		n, readErr := io.ReadFull(src, buf)
		if n == 0 && readErr == io.EOF {
			break
		}
		if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
			return total, fmt.Errorf("crypto: read plaintext: %w", readErr)
		}

		nonce := deriveChunkNonce(baseNonce, chunkIdx)
		ciphertext := gcm.Seal(nil, nonce, buf[:n], nil)

		// Ghi [4 bytes length][ciphertext]
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(ciphertext)))
		if _, err := dst.Write(lenBuf); err != nil {
			return total, fmt.Errorf("crypto: write chunk length: %w", err)
		}
		total += 4

		w, err := dst.Write(ciphertext)
		total += int64(w)
		if err != nil {
			return total, fmt.Errorf("crypto: write chunk: %w", err)
		}

		chunkIdx++
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
	}

	return total, nil
}

// Decrypt membaca ciphertext dari src, mendekripsi, menulis plaintext ke dst.
// Password harus cocok dengan yang dipakai saat Encrypt.
func Decrypt(dst io.Writer, src io.Reader, password string) error {
	// Baca header
	headerBuf := make([]byte, 4+1+saltSize+nonceSize)
	if _, err := io.ReadFull(src, headerBuf); err != nil {
		return ErrShortHeader
	}

	if string(headerBuf[:4]) != HeaderMagic {
		return ErrInvalidMagic
	}

	bits := KeyBits(int(headerBuf[4]) * 64)
	if _, err := keyLen(bits); err != nil {
		return ErrInvalidKeyBits
	}

	salt := headerBuf[5 : 5+saltSize]
	baseNonce := headerBuf[5+saltSize : 5+saltSize+nonceSize]

	key, err := DeriveKey(password, salt, bits)
	if err != nil {
		return err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("crypto: new GCM: %w", err)
	}

	// Giải mã từng chunk
	lenBuf := make([]byte, 4)
	var chunkIdx uint64 = 0
	for {
		_, err := io.ReadFull(src, lenBuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return fmt.Errorf("crypto: read chunk length: %w", err)
		}

		chunkLen := binary.BigEndian.Uint32(lenBuf)
		ciphertext := make([]byte, chunkLen)
		if _, err := io.ReadFull(src, ciphertext); err != nil {
			return fmt.Errorf("crypto: read chunk data: %w", err)
		}

		nonce := deriveChunkNonce(baseNonce, chunkIdx)
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return ErrDecryptChunk
		}

		if _, err := dst.Write(plaintext); err != nil {
			return fmt.Errorf("crypto: write plaintext: %w", err)
		}

		chunkIdx++
	}

	return nil
}

// deriveChunkNonce XOR nonce gốc với chỉ số chunk để mỗi chunk có nonce độc lập.
// Điều này đảm bảo không bao giờ tái sử dụng (key, nonce) pair.
func deriveChunkNonce(baseNonce []byte, chunkIdx uint64) []byte {
	nonce := make([]byte, nonceSize)
	copy(nonce, baseNonce)
	// XOR 8 byte cuối với chunkIdx
	idxBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idxBytes, chunkIdx)
	for i := 0; i < 8; i++ {
		nonce[nonceSize-8+i] ^= idxBytes[i]
	}
	return nonce
}

func keyLen(bits KeyBits) (int, error) {
	switch bits {
	case KeyBits128:
		return 16, nil
	case KeyBits192:
		return 24, nil
	case KeyBits256:
		return 32, nil
	default:
		return 0, ErrInvalidKeyBits
	}
}

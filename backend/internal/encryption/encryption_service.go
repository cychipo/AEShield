// Package encryption cung cấp service lớp cho các thao tác mã hóa
package encryption

import (
	"context"
	"fmt"
	"io"

	"github.com/aeshield/backend/internal/crypto"
)

// EncryptionService cung cấp chức năng mã hóa có thể tái sử dụng
type EncryptionService struct {
}

// NewEncryptionService tạo mới một instance của EncryptionService
func NewEncryptionService() *EncryptionService {
	return &EncryptionService{}
}

// EncryptParams chứa các tham số cho các thao tác mã hóa
type EncryptParams struct {
	Source      io.Reader // Reader đầu vào chứa dữ liệu plaintext
	Destination io.Writer // Writer đầu ra để ghi dữ liệu đã mã hóa
	Password    string    // Mật khẩu do người dùng cung cấp để mã hóa
	KeyBits     crypto.KeyBits // Kích thước khóa AES (128, 192, hoặc 256 bits)
}

// Encrypt thực hiện mã hóa AES-GCM với phương pháp dẫn xuất khóa Argon2id
// theo đặc tả AEShield.
func (s *EncryptionService) Encrypt(ctx context.Context, params EncryptParams) (int64, error) {
	if params.Source == nil {
		return 0, fmt.Errorf("source reader is required")
	}
	if params.Destination == nil {
		return 0, fmt.Errorf("destination writer is required")
	}
	if params.Password == "" {
		return 0, fmt.Errorf("password is required for encryption")
	}

	if _, err := validateKeyBits(params.KeyBits); err != nil {
		return 0, fmt.Errorf("invalid key bits: %w", err)
	}

	return crypto.Encrypt(params.Destination, params.Source, params.Password, params.KeyBits)
}

// DecryptParams chứa các tham số cho các thao tác giải mã
type DecryptParams struct {
	Source      io.Reader // Reader đầu vào chứa dữ liệu đã mã hóa
	Destination io.Writer // Writer đầu ra để ghi dữ liệu đã giải mã
	Password    string    // Mật khẩu do người dùng cung cấp để giải mã (phải trùng với mật khẩu mã hóa)
}

// Decrypt thực hiện giải mã AES-GCM theo đặc tả AEShield.
func (s *EncryptionService) Decrypt(ctx context.Context, params DecryptParams) error {
	if params.Source == nil {
		return fmt.Errorf("source reader is required")
	}
	if params.Destination == nil {
		return fmt.Errorf("destination writer is required")
	}
	if params.Password == "" {
		return fmt.Errorf("password is required for decryption")
	}

	return crypto.Decrypt(params.Destination, params.Source, params.Password)
}

// DeriveKey sử dụng Argon2id để tạo khóa AES từ mật khẩu và salt.
func (s *EncryptionService) DeriveKey(ctx context.Context, password string, salt []byte, bits crypto.KeyBits) ([]byte, error) {
	if _, err := validateKeyBits(bits); err != nil {
		return nil, fmt.Errorf("invalid key bits: %w", err)
	}

	return crypto.DeriveKey(password, salt, bits)
}

// validateKeyBits kiểm tra giá trị key bits có hợp lệ hay không
func validateKeyBits(bits crypto.KeyBits) (int, error) {
	switch bits {
	case crypto.KeyBits128:
		return 16, nil // 128 bits = 16 bytes
	case crypto.KeyBits192:
		return 24, nil // 192 bits = 24 bytes
	case crypto.KeyBits256:
		return 32, nil // 256 bits = 32 bytes
	default:
		return 0, fmt.Errorf("unsupported key size: %d bits", int(bits))
	}
}

// GetDefaultKeyBits trả về kích thước khóa mặc định (AES-256) theo đặc tả
func (s *EncryptionService) GetDefaultKeyBits() crypto.KeyBits {
	return crypto.KeyBits256
}

// GetSupportedKeyBits trả về tất cả các giá trị key bits được hỗ trợ
func (s *EncryptionService) GetSupportedKeyBits() []crypto.KeyBits {
	return []crypto.KeyBits{
		crypto.KeyBits128,
		crypto.KeyBits192,
		crypto.KeyBits256,
	}
}

// ValidateEncryptedFile kiểm tra xem nguồn có chứa header file mã hóa hợp lệ hay không
func (s *EncryptionService) ValidateEncryptedFile(ctx context.Context, source io.Reader) (bool, crypto.KeyBits, error) {
	headerBuf := make([]byte, 4+1+16+12) // magic(4) + keyBits(1) + salt(16) + nonce(12)

	n, err := io.ReadFull(source, headerBuf)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("failed to read file header: %w", err)
	}

	if n < len(headerBuf) {
		return false, 0, nil
	}

	if string(headerBuf[:4]) != crypto.HeaderMagic {
		return false, 0, nil
	}

	bits := crypto.KeyBits(int(headerBuf[4]) * 64)

	if _, err := validateKeyBits(bits); err != nil {
		return false, 0, nil
	}

	return true, bits, nil
}
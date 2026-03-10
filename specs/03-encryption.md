# Đặc tả Mã hóa

## Thuật toán

- **AES-GCM** (Galois/Counter Mode) - Mã hóa xác thực

## Sinh khóa

- **Argon2id** (biến thể khuyến nghị cho sinh khóa từ mật khẩu)
- Mật khẩu người dùng + Salt → Khóa mã hóa
- Độ dài khóa: 128-bit, 192-bit, 256-bit (người dùng chọn)

## Tham số KDF

```go
type KDFParams struct {
    Memory      uint32 // 64MB (65536 KB)
    Iterations  uint32 // 3
    Parallelism uint8  // 4
    SaltLength  uint32 // 32 bytes
    KeyLength   uint32 // 16/24/32 bytes (128/192/256 bits)
}
```

## Mã hóa Streaming

- Sử dụng `io.Copy` với `cipher.StreamWriter`
- Xử lý file theo từng chunk để tránh tràn bộ nhớ
- Kích thước chunk: 64KB - 1MB (cấu hình được)

## Luồng trong Desktop App (Wails)

### Mã hóa File (Go Backend)

1. Người dùng chọn file qua native dialog
2. Người dùng nhập mật khẩu
3. Go backend:
   - Tạo salt ngẫu nhiên (32 bytes)
   - Sinh khóa dùng Argon2id(password, salt)
   - Tạo IV ngẫu nhiên (12 bytes cho GCM)
   - Mã hóa file dùng AES-GCM(key, IV)
4. Stream dữ liệu đã mã hóa lên Cloudflare R2

### Giải mã File (Go Backend)

1. Người dùng yêu cầu tải file
2. Backend tải encrypted data từ R2
3. Go backend:
   - Đọc salt và IV từ file/metadata
   - Sinh khóa dùng Argon2id(password, salt)
   - Giải mã dùng AES-GCM(key, IV)
4. Lưu file đã giải mã qua save dialog

## So sánh Web vs Desktop

| Chức năng | Web | Desktop (Wails) |
|-----------|-----|-----------------|
| Mã hóa | JavaScript (Frontend) | Go (Backend) |
| Giải mã | JavaScript (Frontend) | Go (Backend) |
| Chọn file | HTML input | Native dialog |
| Lưu file | Browser download | Native save dialog |
| Store token | localStorage | Encrypted file |

## Code Example (Go)

```go
func EncryptFile(inputPath, password, encryptionType string) (string, error) {
    keyLen := map[string]uint32{
        "AES-128": 16,
        "AES-192": 24,
        "AES-256": 32,
    }[encryptionType]

    salt := make([]byte, 32)
    io.ReadFull(rand.Reader, salt)

    key, err := argon2id.DeriveKey(password, salt, keyLen)
    if err != nil {
        return "", err
    }

    block, err := aes.NewCipher(key)
    if err != nil {
        return "", err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    nonce := make([]byte, gcm.NonceSize())
    io.ReadFull(rand.Reader, nonce)

    input, err := os.Open(inputPath)
    if err != nil {
        return "", err
    }
    defer input.Close()

    output, err := os.Create(inputPath + ".enc")
    if err != nil {
        return "", err
    }
    defer output.Close()

    writer, err := gcm.NewWriter(nonce, output)
    if err != nil {
        return "", err
    }

    _, err = io.Copy(writer, input)
    if err != nil {
        return "", err
    }
    writer.Close()

    return output.Name(), nil
}
```

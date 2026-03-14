# Đặc tả Mã hóa

## Thuật toán

- **AES-GCM** (Galois/Counter Mode) — mã hóa xác thực (Authenticated Encryption)

## Sinh Khóa (Key Derivation)

- **Argon2id** (OWASP recommended)
- Input: `password` (người dùng cung cấp khi upload) + `salt` ngẫu nhiên
- Output: AES key 128 / 192 / 256 bit (người dùng chọn)
- Mật khẩu **không được lưu** ở bất kỳ đâu — chỉ dùng trong memory khi xử lý request

## Tham số Argon2id

```go
const (
    argonTime    = 1
    argonMemory  = 64 * 1024 // 64 MiB
    argonThreads = 4
)
```

## Định dạng File Mã hóa

Mỗi file được lưu trên R2 có cấu trúc binary:

```
[4 bytes: magic "AES\x00"]
[1 byte:  key bits / 64  ]  → 2=128bit, 3=192bit, 4=256bit
[16 bytes: Argon2id salt ]
[12 bytes: base nonce    ]
[chunk 0]: [4 bytes: len][ciphertext + GCM tag]
[chunk 1]: [4 bytes: len][ciphertext + GCM tag]
...
```

## Mã hóa Streaming (Server-side)

Mã hóa xảy ra **hoàn toàn trên server**, không có dữ liệu rõ nào được lưu vào disk hay buffer toàn bộ vào RAM:

```
multipart upload (HTTP)
        │
        ▼
   io.Pipe writer ◄── crypto.Encrypt() ── chunked AES-GCM
        │
        ▼
   R2 UploadFile (pipe reader)
```

1. Frontend gửi file + password qua `multipart/form-data`
2. Backend mở `io.Pipe` — một goroutine chạy `crypto.Encrypt(pipeWriter, fileBody, password, bits)`
3. R2 client đọc từ `pipeReader` và stream thẳng lên Cloudflare R2
4. Không có byte plaintext nào chạm disk server

## Giải mã

Hiện tại backend trả **presigned URL** để frontend tải file mã hóa về. Việc giải mã do client thực hiện (dùng Web Crypto API với cùng mật khẩu). Hoặc có thể mở rộng thêm endpoint decrypt-proxy trong tương lai.

## Chunk Nonce

Mỗi chunk dùng nonce riêng để không bao giờ tái sử dụng `(key, nonce)` pair:

```go
nonce[i] = baseNonce[i] XOR bigEndian(chunkIndex)[i]  // 8 byte cuối
```

## Kích thước Chunk

- `ChunkSize = 64 * 1024` (64 KiB) — cấu hình tại `internal/crypto/crypto.go`
- Overhead mỗi chunk: 4 bytes length prefix + 16 bytes GCM tag

## Độ dài Khóa

| Tùy chọn | Key bytes | Salt | Bảo mật |
|----------|-----------|------|---------|
| AES-128  | 16        | 16B  | Tốt |
| AES-192  | 24        | 16B  | Rất tốt |
| AES-256  | 32        | 16B  | Tối đa (mặc định) |

## Package

Toàn bộ logic nằm tại `backend/internal/crypto/`:

| File | Mô tả |
|------|-------|
| `crypto.go` | `Encrypt()`, `Decrypt()`, `DeriveKey()` |
| `crypto_test.go` | Unit tests (roundtrip, wrong password, corruption, edge cases) |

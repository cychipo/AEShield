# 📑 Technical Specification: AEShield

## 1. Tổng quan hệ thống

**AEShield** là một nền tảng quản lý và chia sẻ file an toàn. Ứng dụng cho phép mã hóa file cục bộ (Client-side logic/Streaming) trước khi tải lên Cloudflare R2, với các tùy chọn độ dài khóa AES linh hoạt và quản lý quyền truy cập nghiêm ngặt.

### Luồng dữ liệu chính (Data Flow)

---

## 2. Cấu trúc dự án (Monorepo)

Dự án được tổ chức để tách biệt rõ ràng giữa Backend (Go) và Frontend (React).

```text
aeshield/
├── backend/            # Golang (Gin/Fiber)
│   ├── cmd/            # Entry point
│   ├── internal/       # Logic nghiệp vụ (Auth, Encrypt, Storage)
│   ├── models/         # MongoDB Schemas
│   ├── go.mod
│   └── Makefile        # Build/Test commands
├── frontend/           # ReactJS + Vite
│   ├── src/
│   │   ├── components/ # PascalCase components
│   │   ├── hooks/      # Custom React hooks
│   │   └── lib/        # API & Crypto wrappers
│   ├── package.json
│   └── yarn.lock
└── AGENTS.md           # Quy tắc coding (đã cập nhật)

```

---

## 3. Quy trình vận hành & Lệnh (Build/Lint/Test)

Thay vì dùng `npm`, chúng ta sử dụng **Yarn** cho FE và **Standard Go Tools** cho BE:

### Backend (Golang)

- **Chạy Dev:** `go run cmd/main.go`
- **Lint:** `golangci-lint run`
- **Test:** `go test ./...`
- **Build:** `go build -o server cmd/main.go`

### Frontend (React)

- **Cài đặt:** `yarn install`
- **Chạy Dev:** `yarn dev`
- **Lint & Check:** `yarn lint` và `yarn typecheck`
- **Build:** `yarn build`

---

## 4. Đặc tả kỹ thuật chi tiết

### A. Authentication (OAuth2)

Sử dụng Google và GitHub làm nhà cung cấp danh tính.

- **Flow:** Frontend nhận mã Authorization Code -> Gửi về Backend -> Backend đổi mã lấy Access Token -> Tạo JWT (JSON Web Token) riêng cho AEShield.

### B. Cơ chế mã hóa (The Crypto Core)

- **Thuật toán:** AES-GCM (Authenticated Encryption).
- **Key Derivation:** Không lưu mật khẩu. Sử dụng **Argon2id** để chuyển đổi mật khẩu người dùng + Salt thành khóa 128/192/256 bits tùy chọn.
- **Streaming:** Sử dụng `io.Copy` kết hợp với `cipher.StreamWriter` để mã hóa file theo từng khối dữ liệu (chunks), đảm bảo không tràn RAM khi xử lý file lớn.

### C. Quản lý quyền truy cập (Access Control)

Mọi file trong MongoDB sẽ có trường `access_control`:

1. **Public:** Sinh ra một mã Hash (CID) công khai. Ai có link cũng có thể tải.
2. **Private:** Kiểm tra `owner_id` trong JWT phải trùng với `owner_id` của file.
3. **Whitelist:** Một mảng `allowed_users` chứa Email/ID. Backend sẽ kiểm tra danh tính người dùng trước khi tạo _Presigned URL_ từ Cloudflare R2.

### D. Cấu trúc dữ liệu (MongoDB Schema)

```typescript
interface FileMetadata {
  _id: ObjectId;
  owner_id: string;
  filename: string;
  size: number;
  encryption_type: "AES-128" | "AES-192" | "AES-256";
  storage_path: string; // Đường dẫn trên Cloudflare R2
  access_mode: "public" | "private" | "whitelist";
  whitelist: string[]; // Danh sách Email/ID
  created_at: Date;
}
```

---

## 5. Danh sách API (Endpoints)

| Method   | Endpoint                | Description                                  |
| -------- | ----------------------- | -------------------------------------------- |
| `POST`   | `/auth/google`          | Đăng nhập bằng Google                        |
| `POST`   | `/files/encrypt-upload` | Stream mã hóa và đẩy lên R2                  |
| `GET`    | `/files/download/:id`   | Lấy Presigned URL (Check quyền)              |
| `PATCH`  | `/files/share`          | Cập nhật Whitelist hoặc chuyển chế độ Public |
| `DELETE` | `/files/:id`            | Xóa file trên R2 và Metadata                 |

---

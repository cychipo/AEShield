# Đặc tả API Endpoints

## Base URL

```
Development: http://localhost:3000/api/v1
Production:  https://api.aeshield.com/api/v1
```

---

## Xác thực

### GET /auth/urls
Trả về URL OAuth2 để frontend redirect người dùng.

**Response:**
```json
{
  "google": "https://accounts.google.com/o/oauth2/auth?...",
  "github": "https://github.com/login/oauth/authorize?..."
}
```

### GET /auth/google
Redirect trình duyệt đến trang đăng nhập Google.

### GET /auth/google/callback
Đổi Authorization Code lấy JWT. Được gọi từ HTML trung gian (không phải redirect trực tiếp).

**Query params:** `?code=<authorization_code>`

**Response:**
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "name": "John Doe",
    "avatar": "https://example.com/avatar.jpg"
  }
}
```

### GET /auth/github
Redirect trình duyệt đến trang đăng nhập GitHub.

### GET /auth/github/callback
Giống Google callback.

### GET /auth/me _(JWT required)_
Lấy thông tin user hiện tại.

**Headers:** `Authorization: Bearer <token>`

**Response:** `{ user }` object

---

## File

Tất cả endpoints dưới đây yêu cầu `Authorization: Bearer <token>`.

### POST /files/upload
Stream mã hóa file và lưu lên Cloudflare R2.

**Content-Type:** `multipart/form-data`

| Field | Type | Bắt buộc | Mô tả |
|-------|------|----------|-------|
| `file` | file | Có | File cần upload |
| `password` | string | Có | Mật khẩu dùng để mã hóa |
| `encryption_type` | string | Không | `AES-128` / `AES-192` / `AES-256` (mặc định: `AES-256`) |
| `access_mode` | string | Không | `public` / `private` / `whitelist` (mặc định: `private`) |

**Response `201`:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "owner_id": "user_123",
  "filename": "document.pdf",
  "size": 1048576,
  "encryption_type": "AES-256",
  "storage_path": "user_123/550e8400-e29b-41d4-a716-446655440000.pdf",
  "access_mode": "private",
  "whitelist": [],
  "public_cid": "",
  "created_at": "2026-03-15T10:00:00Z",
  "updated_at": "2026-03-15T10:00:00Z"
}
```

---

### GET /files
Lấy danh sách file của user hiện tại.

**Response `200`:** Mảng `FileMetadata[]`

---

### GET /files/:id/download
Kiểm tra quyền truy cập và trả về presigned URL để tải file.

**Response `200`:**
```json
{
  "url": "https://<account>.r2.cloudflarestorage.com/...?X-Amz-Signature=..."
}
```

**Lỗi:**
- `403` — không có quyền
- `404` — file không tồn tại

---

### PATCH /files/share
Cập nhật chế độ chia sẻ hoặc whitelist của file. Chỉ owner mới được phép.

**Request:**
```json
{
  "file_id": "507f1f77bcf86cd799439011",
  "access_mode": "whitelist",
  "whitelist": ["alice@example.com", "bob@example.com"]
}
```

**Response `200`:** `FileMetadata` đã cập nhật

**Lỗi:**
- `400` — `file_id` thiếu hoặc `access_mode` không hợp lệ
- `403` — không phải owner
- `404` — file không tồn tại

---

### DELETE /files/:id
Xóa file trên R2 và metadata trong MongoDB. Chỉ owner.

**Response `200`:**
```json
{ "message": "deleted" }
```

---

## Response Lỗi Chuẩn

```json
{ "error": "mô tả lỗi" }
```

| Status | Ý nghĩa |
|--------|---------|
| `400` | Input không hợp lệ |
| `401` | JWT thiếu hoặc không hợp lệ |
| `403` | Không có quyền truy cập |
| `404` | Tài nguyên không tìm thấy |
| `500` | Lỗi server nội bộ |

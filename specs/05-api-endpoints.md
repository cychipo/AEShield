# Đặc tả API Endpoints

## Base URL

```
Development: http://localhost:8080/api/v1
Production:  https://api.aeshield.com/api/v1

Local (Desktop App): gọi trực tiếp qua Go backend hoặc HTTP requests
```

## Các Endpoint Xác thực

### POST /auth/google
Đổi mã OAuth2 Google lấy JWT.

**Request:**
```json
{
  "code": "string"
}
```

**Response:**
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "user_123",
    "email": "user@example.com",
    "provider": "google"
  }
}
```

### POST /auth/github
Đổi mã OAuth2 GitHub lấy JWT.

**Request:**
```json
{
  "code": "string"
}
```

**Response:** Giống như endpoint Google.

---

## Các Endpoint File

### POST /files/encrypt-upload
Stream file đã mã hóa lên Cloudflare R2.

**Headers:**
```
Authorization: Bearer <jwt>
Content-Type: application/octet-stream
X-Filename: document.pdf
X-Encryption-Type: AES-256
```

**Request Body:** File stream đã mã hóa

**Response:**
```json
{
  "file_id": "507f1f77bcf86cd799439011",
  "filename": "document.pdf",
  "size": 1048576,
  "encryption_type": "AES-256",
  "access_mode": "private",
  "created_at": "2026-03-10T12:00:00Z"
}
```

### GET /files/download/:id
Lấy presigned URL để tải file (kiểm tra quyền).

**Headers:**
```
Authorization: Bearer <jwt>
```

**Response:**
```json
{
  "download_url": "https://r2-bucket.r2.cloudflarestorage.com/...",
  "expires_in": 3600
}
```

### PATCH /files/share
Cập nhật chế độ truy cập hoặc whitelist của file.

**Headers:**
```
Authorization: Bearer <jwt>
```

**Request:**
```json
{
  "file_id": "507f1f77bcf86cd799439011",
  "access_mode": "whitelist",
  "whitelist": ["user1@example.com", "user2@example.com"]
}
```

**Response:**
```json
{
  "success": true,
  "file_id": "507f1f77bcf86cd799439011"
}
```

### DELETE /files/:id
Xóa file khỏi R2 và metadata khỏi MongoDB.

**Headers:**
```
Authorization: Bearer <jwt>
```

**Response:**
```json
{
  "success": true
}
```

---

## Response Lỗi

```json
{
  "error": "error_message",
  "code": "ERROR_CODE"
}
```

Các mã trạng thái phổ biến:
- `200 OK` - Thành công
- `400 Bad Request` - Input không hợp lệ
- `401 Unauthorized` - JWT thiếu/không hợp lệ
- `403 Forbidden` - Truy cập bị từ chối
- `404 Not Found` - File không tìm thấy
- `500 Internal Server Error` - Lỗi server

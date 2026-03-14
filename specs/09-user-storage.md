# Đặc tả Quản lý File Người dùng

## Tổng quan

Mỗi user có một không gian lưu trữ riêng biệt, giới hạn **10 GB**. File được mã hóa server-side trước khi lưu lên Cloudflare R2. User có toàn quyền quản lý file của mình: upload, download, chia sẻ, xóa.

---

## Giới hạn Lưu trữ

| Thuộc tính | Giá trị |
|------------|---------|
| Dung lượng tối đa mỗi user | **10 GB** (10,737,418,240 bytes) |
| Kích thước tối đa 1 file | **1 GB** (1,073,741,824 bytes) |
| Số lượng file | Không giới hạn (chỉ bị giới hạn bởi tổng dung lượng) |
| Tier | Free (duy nhất, không có paid plan) |

> **Lưu ý kích thước:** Trường `size` trong DB lưu kích thước **plaintext gốc**. Kích thước thực tế trên R2 lớn hơn do overhead mã hóa (header 33 bytes + 20 bytes/chunk × số chunks).

---

## Schema MongoDB

### Collection: `user_storage`

Lưu thống kê dung lượng của từng user. Tách riêng khỏi `users` để dễ update atomic và query nhanh.

```go
type UserStorage struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"    json:"id"`
    UserID      string             `bson:"user_id"          json:"user_id"`       // FK → users._id
    UsedBytes   int64              `bson:"used_bytes"       json:"used_bytes"`    // Tổng bytes đã dùng (plaintext size)
    FileCount   int64              `bson:"file_count"       json:"file_count"`    // Số file hiện có
    QuotaBytes  int64              `bson:"quota_bytes"      json:"quota_bytes"`   // Giới hạn: 10 * 1024^3
    UpdatedAt   time.Time          `bson:"updated_at"       json:"updated_at"`
}
```

**Indexes:**
- `{ user_id: 1 }` (unique)

**Document mẫu:**
```json
{
  "_id": "507f1f77bcf86cd799439020",
  "user_id": "507f1f77bcf86cd799439011",
  "used_bytes": 2147483648,
  "file_count": 12,
  "quota_bytes": 10737418240,
  "updated_at": "2026-03-15T10:00:00Z"
}
```

---

### Collection: `files` (đã có — bổ sung chi tiết)

```go
type FileMetadata struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"        json:"id"`
    OwnerID        string             `bson:"owner_id"             json:"owner_id"`
    Filename       string             `bson:"filename"             json:"filename"`
    OriginalName   string             `bson:"original_name"        json:"original_name"`  // Tên gốc trước khi đổi tên
    Size           int64              `bson:"size"                 json:"size"`            // Plaintext bytes
    MimeType       string             `bson:"mime_type"            json:"mime_type"`       // Loại file gốc (vd: "application/pdf")
    EncryptionType string             `bson:"encryption_type"      json:"encryption_type"` // "AES-128"|"AES-192"|"AES-256"
    StoragePath    string             `bson:"storage_path"         json:"storage_path"`   // R2 object key
    AccessMode     string             `bson:"access_mode"          json:"access_mode"`    // "public"|"private"|"whitelist"
    Whitelist      []string           `bson:"whitelist"            json:"whitelist"`       // Email/UserID được phép
    PublicCID      string             `bson:"public_cid,omitempty" json:"public_cid,omitempty"`
    DownloadCount  int64              `bson:"download_count"       json:"download_count"` // Số lần tải xuống
    CreatedAt      time.Time          `bson:"created_at"           json:"created_at"`
    UpdatedAt      time.Time          `bson:"updated_at"           json:"updated_at"`
}
```

**Indexes:**
- `{ owner_id: 1 }` — list file của user
- `{ owner_id: 1, created_at: -1 }` — sort mới nhất lên trên
- `{ public_cid: 1 }` (unique, sparse) — tra cứu file public
- `{ access_mode: 1 }`

---

## Luồng Nghiệp vụ

### Upload File

```
1. Validate: size ≤ 1 GB
2. Validate: user.used_bytes + size ≤ quota_bytes (10 GB)
3. Stream mã hóa → R2 (io.Pipe + crypto.Encrypt)
4. Lưu FileMetadata vào MongoDB
5. Atomic update UserStorage:
   used_bytes += size
   file_count += 1
6. Rollback nếu bước 4 hoặc 5 thất bại: xóa file trên R2
```

### Delete File

```
1. Validate: requester == owner
2. Xóa object trên R2
3. Xóa FileMetadata khỏi MongoDB
4. Atomic update UserStorage:
   used_bytes -= file.size
   file_count -= 1
```

### Kiểm tra Quota

```
Còn lại = quota_bytes - used_bytes
Phần trăm = used_bytes / quota_bytes * 100
```

---

## API Endpoints Bổ sung

Thêm vào `05-api-endpoints.md`:

### GET /api/v1/storage/me
Lấy thông tin dung lượng của user hiện tại.

**Headers:** `Authorization: Bearer <token>`

**Response `200`:**
```json
{
  "used_bytes": 2147483648,
  "quota_bytes": 10737418240,
  "used_gb": 2.0,
  "quota_gb": 10.0,
  "percent_used": 20.0,
  "file_count": 12,
  "available_bytes": 8589934592
}
```

---

## Quy tắc Nghiệp vụ

| Rule | Chi tiết |
|------|----------|
| Quota check | Phải check **trước** khi stream lên R2 (tránh upload rồi reject) |
| Size tính theo plaintext | `size` = kích thước file gốc, không phải sau mã hóa |
| Atomic update | Dùng MongoDB `$inc` operator để cập nhật `used_bytes` và `file_count` |
| Không khôi phục quota nếu R2 xóa thất bại | Xử lý bằng background job cleanup sau |
| Filename | Lưu cả `filename` (có thể đổi sau) và `original_name` (tên lúc upload, bất biến) |
| `download_count` | Tăng mỗi khi presigned URL được tạo thành công |
| Xóa user | Phải xóa hết file trên R2 + `files` + `user_storage` trước khi xóa user |

---

## Storage Path Convention trên R2

```
{owner_id}/{uuid}{ext}
```

Ví dụ:
```
507f1f77bcf86cd799439011/550e8400-e29b-41d4-a716-446655440000.pdf
```

- `owner_id` là ObjectID hex string
- `uuid` đảm bảo không trùng tên dù cùng extension
- Extension giữ nguyên để R2 có thể set Content-Type đúng khi cần

---

## Hiển thị UI

| Thông tin | Vị trí hiển thị |
|-----------|----------------|
| Dung lượng đã dùng / tổng | Sidebar hoặc header (vd: "2.0 GB / 10 GB") |
| Thanh progress | Màu `#F6821F` (primary), chuyển sang `#EF4444` khi > 90% |
| Số file | Hiển thị cùng với dung lượng |
| Cảnh báo 90% | Banner warning khi `percent_used ≥ 90` |
| Lỗi đầy quota | Modal lỗi khi upload thất bại do hết dung lượng (HTTP 413) |

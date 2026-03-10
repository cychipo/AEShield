# Đặc tả Schema Database

## Cấu hình MongoDB

- **Database:** aeshield
- **Collections:** users, files

---

## Collection Users

```go
type User struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Provider    string             `bson:"provider"`      // "google" hoặc "github"
    ProviderID  string             `bson:"provider_id"`   // ID từ OAuth provider
    Email       string             `bson:"email"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}
```

**Indexes:**
- `{ provider: 1, provider_id: 1 }` (unique)
- `{ email: 1 }` (unique)

---

## Collection Files

```go
type FileMetadata struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"`
    OwnerID        string             `bson:"owner_id"`        // User ID từ JWT
    Filename       string             `bson:"filename"`
    Size           int64              `bson:"size"`            // Bytes
    EncryptionType string             `bson:"encryption_type"` // "AES-128", "AES-192", "AES-256"
    StoragePath    string             `bson:"storage_path"`   // R2 object key
    AccessMode     string             `bson:"access_mode"`    // "public", "private", "whitelist"
    Whitelist      []string           `bson:"whitelist"`       // Danh sách email/ID
    PublicCID      string             `bson:"public_cid,omitempty"` // Cho truy cập công khai
    CreatedAt      time.Time          `bson:"created_at"`
    UpdatedAt      time.Time          `bson:"updated_at"`
}
```

**Indexes:**
- `{ owner_id: 1 }`
- `{ public_cid: 1 }` (unique, sparse)
- `{ access_mode: 1 }`

---

## Metadata Mã hóa (lưu riêng hoặc trong header file)

Khi tải lên, Frontend cũng gửi:
- Salt (32 bytes) - lưu như phần của file đã mã hóa hoặc trong metadata
- IV (12 bytes) - lưu như phần của file đã mã hóa

---

## Document Mẫu (Files)

```json
{
  "_id": "507f1f77bcf86cd799439011",
  "owner_id": "user_123",
  "filename": "document.pdf",
  "size": 1048576,
  "encryption_type": "AES-256",
  "storage_path": "users/user_123/2026/03/10/abc123.enc",
  "access_mode": "private",
  "whitelist": [],
  "public_cid": "",
  "created_at": "2026-03-10T12:00:00Z",
  "updated_at": "2026-03-10T12:00:00Z"
}
```

# Đặc tả Schema Database

## Cấu hình MongoDB

- **Database:** `aeshield`
- **Collections:** `users`, `files`

---

## Collection: users

```go
type LinkedProvider struct {
    Provider   string `bson:"provider"    json:"provider"`    // "google" | "github"
    ProviderID string `bson:"provider_id" json:"provider_id"` // ID từ OAuth provider
}

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Email     string             `bson:"email"         json:"email"`
    Name      string             `bson:"name"          json:"name"`
    Avatar    string             `bson:"avatar"        json:"avatar"`
    Providers []LinkedProvider   `bson:"providers"     json:"providers"`
    CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"    json:"updated_at"`
}
```

**Indexes:**
- `{ email: 1 }` (unique)
- `{ providers: { $elemMatch: { provider, provider_id } } }` — dùng `$elemMatch` để tra cứu theo provider

**Ghi chú:**
- Một user có thể có nhiều providers (Google + GitHub merge theo email)
- Index `provider_1_provider_id_1` cũ đã bị drop khi startup để tránh conflict

---

## Collection: files

```go
type FileMetadata struct {
    ID             primitive.ObjectID `bson:"_id,omitempty"          json:"id"`
    OwnerID        string             `bson:"owner_id"               json:"owner_id"`
    Filename       string             `bson:"filename"               json:"filename"`
    Size           int64              `bson:"size"                   json:"size"`
    EncryptionType string             `bson:"encryption_type"        json:"encryption_type"` // "AES-128"|"AES-192"|"AES-256"
    StoragePath    string             `bson:"storage_path"           json:"storage_path"`    // R2 object key
    AccessMode     string             `bson:"access_mode"            json:"access_mode"`     // "public"|"private"|"whitelist"
    Whitelist      []string           `bson:"whitelist"              json:"whitelist"`
    PublicCID      string             `bson:"public_cid,omitempty"   json:"public_cid,omitempty"`
    CreatedAt      time.Time          `bson:"created_at"             json:"created_at"`
    UpdatedAt      time.Time          `bson:"updated_at"             json:"updated_at"`
}
```

**Indexes:**
- `{ owner_id: 1 }`
- `{ public_cid: 1 }` (unique, sparse)
- `{ access_mode: 1 }`

**Ghi chú về mã hóa:**
- Salt và nonce **không được lưu trong MongoDB** — chúng được nhúng vào header của file mã hóa trên R2
- Backend đọc lại header khi decrypt (xem spec 03-encryption.md)
- `size` là kích thước plaintext gốc (byte), kích thước file trên R2 sẽ lớn hơn do overhead mã hóa

---

## Document Mẫu

### User
```json
{
  "_id": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "name": "John Doe",
  "avatar": "https://example.com/avatar.jpg",
  "providers": [
    { "provider": "google", "provider_id": "108234567890" },
    { "provider": "github", "provider_id": "12345678" }
  ],
  "created_at": "2026-03-10T08:00:00Z",
  "updated_at": "2026-03-15T10:00:00Z"
}
```

### FileMetadata
```json
{
  "_id": "507f1f77bcf86cd799439012",
  "owner_id": "507f1f77bcf86cd799439011",
  "filename": "document.pdf",
  "size": 1048576,
  "encryption_type": "AES-256",
  "storage_path": "507f1f77bcf86cd799439011/550e8400-e29b-41d4-a716-446655440000.pdf",
  "access_mode": "private",
  "whitelist": [],
  "public_cid": "",
  "created_at": "2026-03-15T10:00:00Z",
  "updated_at": "2026-03-15T10:00:00Z"
}
```

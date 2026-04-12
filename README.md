# AEShield

Secure file management and sharing platform with client-side AES encryption before uploading to Cloudflare R2.

## Features

- **Client-side Encryption**: Files are encrypted locally using AES-GCM (128/192/256-bit keys) before upload
- **Secure Storage**: Cloudflare R2 for object storage
- **Flexible Access Control**: Public, Private, and Whitelist modes
- **Cross-platform**: macOS, Windows, Linux support via Wails

## Tech Stack

- **Frontend**: React + Vite
- **Backend**: Go (Wails)
- **Encryption**: AES-GCM with Argon2id key derivation
- **Storage**: Cloudflare R2

## Getting Started

### Prerequisites

- Go 1.21+
- Node.js 18+
- Yarn

### Installation

```bash
# Install dependencies
make install
# or
cd frontend && yarn install
```

### Development

```bash
make dev
# or
wails dev
```

### Build

```bash
# Build for current platform
wails build

# Build for specific platform
make build:mac    # macOS (ARM + Intel)
make build:win    # Windows
make build:linux  # Linux
```

## Project Structure

```
aeshield/
├── main.go           # Entry point
├── app.go            # App struct
├── models/           # MongoDB schemas
├── frontend/         # React frontend
│   └── src/
│       ├── components/
│       ├── hooks/
│       └── lib/
├── Makefile
└── wails.json
```

## API Endpoints (Backend)

| Method   | Endpoint                | Description                    |
| -------- | ----------------------- | ------------------------------ |
| `POST`   | `/auth/google`          | Google OAuth login             |
| `POST`   | `/files/encrypt-upload`| Encrypt & upload to R2        |
| `GET`    | `/files/download/:id`   | Download file (check access)   |
| `PATCH`  | `/files/share`          | Update access mode             |
| `DELETE` | `/files/:id`            | Delete file                    |

## Docker Deploy on Ubuntu VPS

1. Vào thư mục `deploy/` rồi sao chép `.env.example` thành `.env`.
2. Cập nhật các giá trị trong `deploy/.env`, đặc biệt là:
   - `FRONTEND_URL`
   - `GOOGLE_REDIRECT_URL`
   - `GITHUB_REDIRECT_URL`
   - `JWT_SECRET`
   - `MONGO_BOOTSTRAP_ROOT_USERNAME`
   - `MONGO_BOOTSTRAP_ROOT_PASSWORD`
   - `MONGO_ADMIN_USERNAME`
   - `MONGO_ADMIN_PASSWORD`
   - các biến Cloudflare R2 và OAuth
3. Mở cổng `5191`, `8010` và `27020` trên VPS firewall; trỏ domain hoặc IP public vào cổng `5191` cho frontend nếu dùng same-origin proxy.
4. Khởi động toàn bộ stack từ trong `deploy/`:

```bash
cd deploy
docker compose --env-file .env up --build -d
```

Hoặc chạy từ thư mục gốc:

```bash
make deploy
```

5. Xác minh:
   - Frontend: `http://<vps-host>:5191`
   - API qua cùng origin frontend: `http://<vps-host>:5191/api/v1/auth/urls`

Useful commands:

```bash
make deploy-build
make deploy-down
```

The deployment stack includes:
- MongoDB container
- A reconciliation step for the deployment-managed MongoDB admin account
- Backend container listening internally on port `8010`
- Frontend/nginx container exposing public port `5191` and proxying `/api` sang backend

## License

MIT

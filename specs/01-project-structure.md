# Đặc tả Cấu trúc Dự án

## Tổng quan

AEShield là nền tảng web quản lý và chia sẻ file an toàn. File được mã hóa **server-side** bằng AES-GCM streaming trước khi lưu lên Cloudflare R2. Người dùng đăng nhập qua OAuth2 (Google / GitHub).

## Cấu trúc Thư mục (Monorepo)

```
aeshield/
├── backend/                   # Go (Fiber v2)
│   ├── cmd/
│   │   └── main.go            # Entry point, route registration
│   ├── internal/
│   │   ├── auth/              # OAuth2 + JWT (service, handlers, middleware)
│   │   ├── config/            # Cấu hình từ .env
│   │   ├── crypto/            # AES-GCM streaming encryption (Argon2id key derive)
│   │   ├── database/          # MongoDB connection + user repository
│   │   ├── files/             # Upload / Download / Delete / Share (service + handler)
│   │   └── storage/           # R2 client + file metadata repository
│   ├── models/                # MongoDB schemas (User, FileMetadata, Claims)
│   ├── docs/                  # Swagger / OpenAPI generated files
│   ├── e2e/                   # End-to-end tests
│   ├── go.mod
│   ├── Makefile
│   └── .env.example
├── frontend/                  # React + Vite + TypeScript
│   ├── src/
│   │   ├── components/        # PascalCase UI components
│   │   ├── hooks/             # Custom React hooks
│   │   ├── lib/               # API client & helpers
│   │   └── pages/             # Route-level page components
│   ├── package.json
│   ├── vite.config.ts
│   └── yarn.lock
├── specs/                     # Tài liệu đặc tả
│   └── *.md
├── AGENTS.md                  # Quy tắc coding
└── README.md
```

## Công nghệ Sử dụng

| Layer | Công nghệ |
|-------|-----------|
| Backend framework | Go + Fiber v2 |
| Authentication | OAuth2 (Google, GitHub) + JWT |
| Encryption | AES-GCM + Argon2id (server-side streaming) |
| Database | MongoDB |
| Object storage | Cloudflare R2 (AWS S3-compatible API) |
| Frontend | React + Vite + TypeScript |
| Package manager (FE) | Yarn |

## Lệnh Vận hành

### Backend (Go)

```bash
# Chạy dev
go run cmd/main.go

# Lint
golangci-lint run

# Test
go test ./...

# Build
go build -o server cmd/main.go
```

### Frontend (React)

```bash
# Cài đặt
yarn install

# Chạy dev
yarn dev

# Lint & typecheck
yarn lint
yarn typecheck

# Build production
yarn build
```

## Cổng Mặc định

| Service | Port |
|---------|------|
| Backend API | `3000` |
| Frontend dev server | `5173` (Vite) |

Frontend production được serve bởi backend từ thư mục `frontend/dist/`.

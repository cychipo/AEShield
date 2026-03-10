# Đặc tả Cấu trúc Dự án

## Tổng quan

AEShield là nền tảng quản lý và chia sẻ file an toàn với mã hóa phía client trước khi tải lên Cloudflare R2.

## Cấu trúc Thư mục

```
aeshield/
├── backend/                    # Go + Wails
│   ├── cmd/                    # Điểm khởi đầu
│   │   └── main.go
│   ├── internal/               # Logic nghiệp vụ
│   │   ├── auth/               # Xử lý xác thực
│   │   ├── encrypt/            # Logic mã hóa
│   │   ├── storage/           # Tích hợp Cloudflare R2
│   │   └── middleware/        # JWT, CORS, v.v.
│   ├── models/                 # Schema MongoDB
│   │   └── file.go
│   ├── frontend/               # Frontend (Wails embed)
│   │   ├── src/                # React
│   │   ├── package.json
│   │   └── vite.config.ts
│   ├── go.mod
│   └── Makefile
├── specs/
│   └── *.md
└── AGENTS.md
```

## Công nghệ Sử dụng

- **Framework:** Wails (Go + WebView)
- **Frontend:** React
- **Backend:** Go (Gin/Fiber middleware)
- **Database:** MongoDB
- **Storage:** Cloudflare R2
- **Quản lý package:** Yarn (FE), Go modules (BE)

## Build Targets

| Nền tảng | Output                                         |
| -------- | ---------------------------------------------- |
| macOS    | `.app` (darwin-arm64, darwin-amd64)            |
| Windows  | `.exe` (windows-amd64)                         |
| Linux    | `.AppImage`, `.deb` (linux-arm64, linux-amd64) |

## Lệnh Build

```bash
# Development
wails dev

# Build production
wails build -platform=darwin-arm64   # macOS ARM
wails build -platform=darwin-amd64   # macOS Intel
wails build -platform=windows-amd64  # Windows

# Build tất cả platforms
wails build
```

## Wails Configuration (wails.json)

```json
{
  "name": "AEShield",
  "outputfilename": "AEShield",
  "frontend:install": "yarn install",
  "frontend:build": "yarn build",
  "frontend:dev:watcher": "yarn dev",
  "author": {
    "name": "AEShield Team"
  },
  "version": "1.0.0",
  "wailsjsdir": "./frontend/src"
}
```

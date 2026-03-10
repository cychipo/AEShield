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

## License

MIT

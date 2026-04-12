# CLAUDE.md

Tệp này cung cấp hướng dẫn cho Claude Code (`claude.ai/code`) khi làm việc với mã nguồn trong repository này.

## Cấu trúc repository

- Thư mục gốc chỉ là lớp bọc workspace mỏng. Công việc backend hằng ngày diễn ra trong `backend/`; ứng dụng frontend nằm trong `frontend/`.
- Có hai Go module: module gốc dành cho Wails shell (`go.mod`) và module HTTP backend thực tế trong `backend/go.mod`. Với công việc liên quan đến API/server, ưu tiên chạy lệnh từ `backend/`.
- `AGENTS.md` chứa phần tổng quan dự án đầy đủ nhất trong repo này và chính xác hơn README cấp cao nhất về việc tách backend/frontend cũng như các lệnh test.

## Các lệnh thường dùng

### Thư mục gốc

- `make install` — cài dependencies cho frontend và tải Go modules cho backend
- `make dev` — chạy backend từ `backend/cmd/main.go`
- `make build` — build `frontend/dist` và biên dịch `backend/server`
- `make run` — chạy file nhị phân backend đã biên dịch
- `make clean` — xóa `frontend/dist` và `backend/server`

### Frontend (`frontend/`)

- `yarn install` — cài dependencies
- `yarn dev` — khởi động Vite dev server
- `yarn build` — tạo bản build production
- `yarn preview` — xem trước bản build frontend

Ghi chú:
- Hiện tại chưa có script test, lint hoặc typecheck riêng trong `frontend/package.json`.
- Frontend trỏ đến `import.meta.env.VITE_API_URL`, nếu không có thì mặc định là `http://localhost:6868/api/v1`.

### Backend (`backend/`)

- `make dev` — chạy backend với Air live reload; cần cài `air` trước (`go install github.com/air-verse/air@latest`)
- `go run cmd/main.go` — chạy backend trực tiếp không qua Air; dùng cách này nếu chưa cài `air`
- `make build` — build file `server`
- `make lint` — chạy `golangci-lint`
- `make test` — chạy toàn bộ backend tests ở chế độ verbose
- `make test-unit` — chạy tests trong `internal/...`
- `make test-e2e` — chạy tests trong `e2e/...`
- `go test ./internal/auth/... -run TestGetAuthURLs -v` — chạy một backend test cụ thể
- `go test ./... -cover` — chạy backend tests kèm coverage

Ghi chú về E2E:
- Các test trong `backend/e2e` yêu cầu server đang chạy.
- URL server mặc định là `http://localhost:6868`; có thể ghi đè bằng `TEST_SERVER_URL=...`.

## Môi trường và runtime

- Cấu hình backend được nạp từ `backend/.env` thông qua `godotenv`; sao chép từ `backend/.env.example`.
- Các tích hợp backend bắt buộc gồm MongoDB, OAuth credentials (Google và GitHub), JWT secret, và Cloudflare R2.
- Cổng mặc định của backend là `6868`.
- Backend phục vụ `frontend/dist` như static files. Nếu `FRONTEND_DIST` không được đặt, nó sẽ trỏ tới `<repo>/frontend/dist`.

## Kiến trúc tổng quan

### Frontend

- SPA dùng React + Vite, định tuyến nằm trong `frontend/src/App.jsx`.
- Cấu trúc ứng dụng hiện tại khá tối giản: `/` hiển thị trang đăng nhập và `/dashboard` được bảo vệ bằng cách kiểm tra token trong `localStorage`.
- Luồng đăng nhập là OAuth khởi tạo từ frontend: `frontend/src/pages/Login.jsx` gọi `/api/v1/auth/{provider}`, sau đó chuyển hướng trình duyệt tới URL nhà cung cấp trả về từ backend.
- Sau khi OAuth hoàn tất, các trang callback của backend sẽ ghi `aeshield_token` và `aeshield_user` vào `localStorage`, rồi chuyển hướng đến `/dashboard`.
- `frontend/src/pages/Dashboard.jsx` gọi `/api/v1/auth/me` với bearer token và xem phản hồi đó là nguồn sự thật của session.

### Backend

- `backend/cmd/main.go` khởi tạo toàn bộ ứng dụng HTTP: nạp config, kết nối MongoDB, khởi tạo repositories/services, tạo app Fiber, đăng ký auth routes và file routes, sau đó phục vụ SPA đã build.
- Backend được tổ chức theo tính năng trong `backend/internal/`:
  - `auth/` xử lý đăng nhập OAuth, callback từ provider, tạo JWT, và auth middleware.
  - `files/` xử lý các luồng upload/list/download/delete/share.
  - `crypto/` và `encryption/` chứa helper mã hóa và các test liên quan.
  - `storage/` bao bọc truy cập Cloudflare R2 và lưu metadata file.
  - `database/` quản lý kết nối MongoDB/repository logic.
  - `config/` tập trung việc nạp biến môi trường.
- Logic auth service gộp danh tính theo provider trước, sau đó theo email, nhờ đó một user record có thể tích lũy nhiều provider đã liên kết.
- Các protected routes dùng trực tiếp `auth.JWTMiddleware(cfg.JWTSecret)` trên từng route thay vì dùng route groups.

## Luồng xử lý file và mã hóa

- Upload được xử lý qua `backend/internal/files/service.go`.
- Backend mã hóa luồng file đầu vào trước khi upload bằng `crypto.Encrypt(...)` và `io.Pipe`; ciphertext được stream thẳng vào Cloudflare R2 thay vì buffer toàn bộ file trong bộ nhớ.
- Metadata của file được lưu riêng trong MongoDB, bao gồm owner, loại mã hóa, đường dẫn lưu trữ, chế độ truy cập, whitelist, và CID công khai tùy chọn.
- Kiểm soát truy cập được thực thi trong file service:
  - `public` — bất kỳ requester nào cũng có thể lấy presigned download URL
  - `private` — chỉ chủ sở hữu mới được tải xuống
  - `whitelist` — chủ sở hữu hoặc requester có trong whitelist mới được tải xuống
- File được tải xuống thông qua presigned R2 URLs, không phải bằng cách backend proxy nội dung file qua API.

## Hình dạng API

Các route chính được định nghĩa trong `backend/cmd/main.go`:
- Auth công khai: `/api/v1/auth/urls`, `/api/v1/auth/google`, `/api/v1/auth/google/callback`, `/api/v1/auth/github`, `/api/v1/auth/github/callback`
- Auth được bảo vệ: `/api/v1/auth/me`
- File APIs được bảo vệ: `/api/v1/files/upload`, `/api/v1/files`, `/api/v1/files/:id/download`, `/api/v1/files/:id`, `/api/v1/files/share`
- Docs/static: `/docs` và `/api/v1/swagger.json`

## Thực trạng testing trong repo này

- Backend unit tests tồn tại trong các package tính năng dưới `backend/internal/...`.
- Có ít nhất một bộ backend E2E trong `backend/e2e/auth_e2e_test.go`.
- README ở thư mục gốc đã cũ một phần: nó vẫn mô tả các lệnh theo kiểu Wails và cấu trúc dự án cũ hơn. Khi có mâu thuẫn, hãy ưu tiên `backend/` source files, `backend/TESTING.md`, `backend/Makefile`, và `AGENTS.md` khi có xung đột.

## Active Technologies
- Go backend, JavaScript React frontend, Docker-based deployment artifacts + Fiber backend, MongoDB, React, Vite, Yarn, Go toolchain (010-vps-deployment)
- MongoDB cho dữ liệu ứng dụng; filesystem/static build artifacts cho frontend production (010-vps-deployment)

## Recent Changes
- 010-vps-deployment: Added Go backend, JavaScript React frontend, Docker-based deployment artifacts + Fiber backend, MongoDB, React, Vite, Yarn, Go toolchain

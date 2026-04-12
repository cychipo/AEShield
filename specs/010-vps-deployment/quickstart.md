# Quickstart: Bộ triển khai VPS

## Mục tiêu
Triển khai toàn bộ AEShield lên một Ubuntu VPS bằng một lệnh, với:
- thư mục vận hành riêng `deploy/`
- frontend public ở cổng `5191`
- backend public ở cổng `8010` và đồng thời được frontend truy cập qua proxy cùng origin
- MongoDB chạy cùng stack
- tài khoản admin MongoDB được đối soát từ env mỗi lần deploy

## Điều kiện tiên quyết
- Ubuntu VPS có Docker và Docker Compose plugin đã được cài đặt
- Firewall/network cho phép truy cập cổng `5191`, `8010` và `27020`
- Có đầy đủ biến môi trường production cho backend, MongoDB, OAuth và Cloudflare R2
- Có giá trị username/password cho tài khoản admin MongoDB do deployment quản lý

## Chuẩn bị cấu hình
1. Vào thư mục `deploy/`.
2. Sao chép `.env.example` thành `.env`.
3. Cập nhật các biến quan trọng trong `deploy/.env`:
   - `FRONTEND_PUBLIC_PORT`
   - `FRONTEND_URL`
   - `GOOGLE_REDIRECT_URL`
   - `GITHUB_REDIRECT_URL`
   - `JWT_SECRET`
   - `MONGO_URI`
   - `MONGO_BOOTSTRAP_ROOT_USERNAME`
   - `MONGO_BOOTSTRAP_ROOT_PASSWORD`
   - `MONGO_ADMIN_USERNAME`
   - `MONGO_ADMIN_PASSWORD`
   - các biến OAuth và Cloudflare R2
4. Giữ `VITE_API_URL=/api/v1` để frontend production gọi backend qua same-origin proxy.
5. Có thể giữ `BACKEND_PUBLIC_PORT=8010` nếu muốn truy cập backend trực tiếp ngoài frontend proxy.
6. Bảo đảm không còn dịch vụ khác đang chiếm các cổng public `5191`, `8010`, `27020`.

## Luồng triển khai cơ bản
1. Từ thư mục `deploy/`, chạy:
   ```bash
   docker compose --env-file .env up --build -d
   ```
   Hoặc từ thư mục gốc repository chạy `make deploy`.
2. MongoDB khởi động trước và chờ healthcheck thành công.
3. Tác vụ `mongo-reconcile` chạy sau khi Mongo healthy để create/no-op/update/replace tài khoản admin theo env.
4. Backend chỉ khởi động sau khi `mongo-reconcile` hoàn tất thành công.
5. Frontend/nginx public ở `http://<vps-host>:5191` và proxy các request `/api` sang backend nội bộ.
6. Backend cũng được publish trực tiếp ở `http://<vps-host>:8010` để phục vụ nhu cầu truy cập riêng nếu cần.
7. Dừng stack bằng `docker compose --env-file .env down` trong `deploy/` hoặc `make deploy-down` từ thư mục gốc.

## Kiểm tra sau triển khai
- Mở frontend trên `http://<vps-host>:5191` và xác nhận trang tải thành công
- Gọi `http://<vps-host>:5191/api/v1/auth/urls` và xác nhận API phản hồi qua proxy cùng origin
- Gọi `http://<vps-host>:8010/api/v1/auth/urls` và xác nhận backend public phản hồi trực tiếp
- Kiểm tra tài khoản admin MongoDB theo expected case:
  - deploy đầu tiên: tài khoản được tạo
  - deploy lại với cùng env: tài khoản giữ nguyên
  - đổi password: password mới có hiệu lực
  - đổi username: user cũ bị thay thế bởi user mới
- Kiểm tra log của service `mongo-reconcile` để xác nhận hành động thực tế đã chạy đúng case

## Tình huống lỗi mong đợi
- Thiếu env admin MongoDB: deploy dừng với lỗi rõ ràng
- Cổng `5191`, `8010` hoặc `27020` bị chiếm: deploy dừng với lỗi rõ ràng
- Không thể xóa hoặc cập nhật admin MongoDB cũ: deploy dừng và không báo thành công

## Trạng thái xác minh hiện tại
- `docker compose --env-file .env config` trong `deploy/` đã chạy thành công.
- `yarn build` trong `frontend/` đã chạy thành công với cấu hình same-origin `/api/v1`.
- `go build -o server ./cmd/main.go` trong `backend/` đã chạy thành công.
- `docker compose --env-file .env up --build -d` trong `deploy/` đã khởi động stack thành công sau khi sửa wrapper `mongosh --nodb`.
- `curl -I http://127.0.0.1:5191` đã trả về `HTTP/1.1 200 OK`.
- `curl http://127.0.0.1:5191/api/v1/auth/urls` đã phản hồi thành công qua nginx proxy.
- `go test ./...` trong `backend/` hiện chưa pass toàn bộ do đã có test lỗi sẵn ở `backend/e2e` và `backend/internal/accesscontrol/handler`, không phải do phần deploy vừa thêm.

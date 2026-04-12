# Deployment Contract: Bộ triển khai VPS

## Purpose
Mô tả các hợp đồng vận hành mà implementation phải cung cấp để đáp ứng spec triển khai VPS.

## 1. Operator Command Contract

- Hệ thống phải cung cấp đúng một lệnh chính thức để operator dùng cho cả deploy lần đầu và redeploy.
- Lệnh này phải:
  - khởi động hoặc cập nhật frontend, backend và MongoDB như một stack thống nhất
  - thất bại rõ ràng nếu thiếu env bắt buộc
  - thất bại rõ ràng nếu cổng `5191` hoặc `8010` không khả dụng
  - có thể chạy lặp lại an toàn với cùng cấu hình

## 2. Public Reachability Contract

- Sau deploy thành công:
  - frontend phải truy cập được từ bên ngoài qua cổng `5191`
  - backend phải truy cập được từ bên ngoài qua cổng `8010`
- Việc publish cổng phải ổn định giữa deploy đầu tiên và các lần redeploy.

## 3. Mongo Admin Reconciliation Contract

Mỗi lần deploy phải thực hiện đối soát tài khoản admin MongoDB do deployment quản lý theo đúng quy tắc sau:

1. Nếu chưa có tài khoản admin do deployment quản lý, tạo tài khoản mới từ env.
2. Nếu username và password hiện có khớp env, không tạo lại.
3. Nếu username khớp nhưng password khác, cập nhật password để khớp env.
4. Nếu username khác, xóa tài khoản cũ do deployment quản lý và thay bằng tài khoản mới từ env.
5. Nếu đối soát thất bại, deployment phải thất bại và không báo thành công.

## 4. Managed Identity Boundary Contract

- Chỉ tài khoản được đánh dấu là “do deployment quản lý” mới được thay đổi hoặc xóa tự động.
- Các tài khoản MongoDB do người vận hành tạo thủ công nằm ngoài phạm vi tác động tự động.

## 5. Runtime Configuration Contract

- Frontend build production phải dùng cấu hình API phù hợp với backend public endpoint `:8010`.
- Backend runtime phải nhận được cấu hình MongoDB, JWT, `FRONTEND_DIST`, `MONGO_ADMIN_*`, `MONGO_BOOTSTRAP_ROOT_*` và các biến cần thiết khác từ môi trường deploy.
- Không được phụ thuộc vào giá trị localhost mặc định cũ cho môi trường VPS production.

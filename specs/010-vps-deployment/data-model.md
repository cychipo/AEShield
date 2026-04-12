# Data Model: Bộ triển khai VPS

## 1. Deployment Environment

### Purpose
Tập hợp cấu hình đầu vào cho một lần triển khai trên Ubuntu VPS.

### Fields
- `frontend_public_port`: cổng public cho frontend, cố định là `5191`
- `backend_public_port`: cổng public cho backend, cố định là `8010`
- `mongo_connection_target`: thông tin kết nối MongoDB nội bộ cho backend
- `mongo_admin_username`: username admin MongoDB do deployment quản lý
- `mongo_admin_password`: password admin MongoDB do deployment quản lý
- `application_runtime_variables`: các biến môi trường backend/frontend cần cho runtime

### Validation Rules
- Phải có đủ các biến môi trường bắt buộc trước khi chạy deploy
- Không chấp nhận username/password admin MongoDB rỗng
- Không chấp nhận cấu hình thiếu thông tin để frontend gọi đúng backend public endpoint

## 2. Deployment Stack

### Purpose
Mô tả một lần khởi chạy đầy đủ của frontend, backend, MongoDB và lớp public exposure.

### Fields
- `frontend_service`: dịch vụ cung cấp giao diện người dùng
- `backend_service`: dịch vụ API và xử lý static assets
- `database_service`: dịch vụ MongoDB
- `public_entrypoints`: danh sách cổng public bắt buộc
- `deployment_command`: lệnh duy nhất dùng để khởi động hoặc cập nhật stack
- `deployment_status`: trạng thái của lần triển khai

### Relationships
- Deployment Stack sử dụng một Deployment Environment
- Deployment Stack sở hữu đúng một Deployment-Managed MongoDB Admin Account
- Deployment Stack công bố hai Port Exposure Rules

### State Transitions
- `pending` → `starting`
- `starting` → `running`
- `starting` → `failed`
- `running` → `updating`
- `updating` → `running`
- `updating` → `failed`

## 3. Deployment-Managed MongoDB Admin Account

### Purpose
Định danh admin MongoDB được tạo và đối soát bởi quy trình deploy.

### Fields
- `username`
- `password`
- `managed_marker`: dấu hiệu nhận diện đây là tài khoản do deployment quản lý
- `last_reconciled_at`
- `reconciliation_result`

### Validation Rules
- Chỉ một tài khoản admin MongoDB do deployment quản lý được coi là active tại một thời điểm
- Nếu username giữ nguyên nhưng password đổi, chỉ cập nhật password
- Nếu username đổi, tài khoản cũ do deployment quản lý phải bị vô hiệu/xóa trước khi hoàn tất triển khai

### State Transitions
- `absent` → `created`
- `created` → `unchanged`
- `created` → `password_updated`
- `created` → `replaced`
- `created` → `removal_failed`

## 4. Admin Reconciliation Run

### Purpose
Một lần thực thi logic đối soát tài khoản admin MongoDB trong quá trình deploy.

### Fields
- `desired_username`
- `desired_password`
- `existing_managed_username`
- `action_taken`
- `result_status`
- `failure_reason`

### Validation Rules
- Mỗi lần deploy phải có tối đa một reconciliation run chính thức
- Nếu reconcile thất bại, Deployment Stack không được báo thành công

## 5. Port Exposure Rule

### Purpose
Quy định khả năng truy cập public của từng entrypoint.

### Fields
- `service_name`
- `public_port`
- `reachability_requirement`
- `failure_behavior`

### Validation Rules
- Frontend phải map ra `5191`
- Backend phải map ra `8010`
- Nếu bind cổng thất bại, deployment phải trả về lỗi rõ ràng và không báo thành công

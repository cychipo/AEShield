# Tasks: Bộ triển khai VPS

**Input**: Design documents from `/specs/010-vps-deployment/`
**Prerequisites**: [plan.md](./plan.md), [spec.md](./spec.md), [research.md](./research.md), [data-model.md](./data-model.md), [contracts/deployment-contract.md](./contracts/deployment-contract.md), [quickstart.md](./quickstart.md)

**Tests**: Spec không yêu cầu TDD hay tạo test suite riêng; các task dưới đây tập trung vào triển khai và các bước xác minh theo quickstart.

**Organization**: Tasks được nhóm theo user story để mỗi story có thể triển khai và kiểm thử độc lập.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Có thể chạy song song (khác file, không phụ thuộc task chưa hoàn tất)
- **[Story]**: User story mà task phục vụ (`[US1]`, `[US2]`, `[US3]`)
- Mỗi task đều ghi rõ file path cần sửa/tạo

## Path Conventions

- Deployment assets vận hành: `deploy/`
- Backend runtime/config: `backend/`
- Frontend build/runtime config: `frontend/`
- Mongo reconciliation scripts: `scripts/`

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Tạo khung deployment assets và cấu hình môi trường dùng chung

- [X] T001 Tạo file mô tả biến môi trường deploy tại `deploy/.env.example`
- [X] T002 [P] Tạo Dockerfile backend production tại `backend/Dockerfile`
- [X] T003 [P] Tạo Dockerfile frontend/reverse-proxy production tại `frontend/Dockerfile`
- [X] T004 [P] Tạo cấu hình reverse proxy public FE `5191` và proxy `/api` sang backend tại `deploy/nginx/default.conf`
- [X] T005 Tạo file orchestration một lệnh triển khai tại `deploy/docker-compose.yml`
- [X] T006 Tạo entrypoint lệnh deploy tại `Makefile`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Hạ tầng dùng chung bắt buộc phải có trước khi hoàn thiện từng user story

**⚠️ CRITICAL**: Không bắt đầu user story nếu phase này chưa xong

- [X] T007 Mở rộng cấu hình backend để nhận env deploy và Mongo admin credentials tại `backend/internal/config/config.go`
- [X] T008 [P] Cập nhật ví dụ env backend production tại `backend/.env.example`
- [X] T009 [P] Chuẩn hóa cấu hình build frontend API endpoint cho production tại `frontend/package.json`
- [X] T010 Tạo script đối soát Mongo admin theo env tại `scripts/mongo/reconcile-admin.mjs`
- [X] T011 Tạo shell wrapper chờ Mongo healthy và chạy reconciliation tại `scripts/mongo/run-reconcile.sh`
- [X] T012 Tích hợp tác vụ reconcile vào orchestration deploy tại `deploy/docker-compose.yml`

**Checkpoint**: Foundation sẵn sàng để hoàn thiện từng user story

---

## Phase 3: User Story 1 - Triển khai toàn bộ hệ thống trên Ubuntu VPS mới (Priority: P1) 🎯 MVP

**Goal**: Operator có thể chạy một lệnh để đưa FE, BE, MongoDB lên cùng một stack từ thư mục `deploy/`, với frontend public ở `:5191` và backend được truy cập qua proxy cùng origin.

**Independent Test**: Từ máy Ubuntu VPS sạch đã cài Docker, copy `deploy/.env.example` thành `deploy/.env`, chạy đúng một lệnh trong `deploy/`, rồi truy cập được frontend qua `:5191` và API backend qua `:5191/api/v1/...`.

### Implementation for User Story 1

- [X] T013 [US1] Cập nhật backend để dùng port/runtime config phù hợp container deploy tại `backend/cmd/main.go`
- [X] T014 [P] [US1] Thay fallback API URL production trong `frontend/src/pages/Login.jsx`
- [X] T015 [P] [US1] Thay fallback API URL production trong `frontend/src/pages/Dashboard.jsx`
- [X] T016 [P] [US1] Thay fallback API URL production trong `frontend/src/pages/Files.jsx`
- [X] T017 [P] [US1] Thay fallback API URL production trong `frontend/src/pages/Settings.jsx`
- [X] T018 [P] [US1] Thay fallback API URL production trong `frontend/src/context/NotificationsContext.jsx`
- [X] T019 [US1] Hoàn thiện mapping service, volume và port publish cho frontend/backend tại `deploy/docker-compose.yml`
- [X] T020 [US1] Bổ sung hướng dẫn chạy một lệnh từ `deploy/` và xác minh frontend cùng proxied API tại `specs/010-vps-deployment/quickstart.md`

**Checkpoint**: User Story 1 hoàn chỉnh khi stack có thể deploy mới từ `deploy/`, frontend public ở `5191`, và API phản hồi qua same-origin proxy.

---

## Phase 4: User Story 2 - Cấu hình tài khoản admin MongoDB qua biến môi trường triển khai (Priority: P2)

**Goal**: Operator có thể cung cấp username/password admin MongoDB qua env và deploy sẽ tạo hoặc giữ nguyên đúng tài khoản mong muốn.

**Independent Test**: Deploy với `MONGO_ADMIN_USERNAME` và `MONGO_ADMIN_PASSWORD`, sau đó xác nhận account được tạo ở lần đầu và giữ nguyên khi redeploy với cùng giá trị.

### Implementation for User Story 2

- [X] T021 [US2] Thêm logic validate env Mongo admin bắt buộc tại `scripts/mongo/reconcile-admin.mjs`
- [X] T022 [US2] Cài đặt luồng create/no-op cho tài khoản admin do deployment quản lý tại `scripts/mongo/reconcile-admin.mjs`
- [X] T023 [US2] Đánh dấu managed identity boundary cho tài khoản admin do deployment tạo tại `scripts/mongo/reconcile-admin.mjs`
- [X] T024 [US2] Bổ sung biến Mongo admin và ví dụ sử dụng trong `deploy/.env.example`
- [X] T025 [US2] Cập nhật hợp đồng vận hành cho create/no-op theo env tại `specs/010-vps-deployment/contracts/deployment-contract.md`

**Checkpoint**: User Story 2 hoàn chỉnh khi deploy đầu tiên tạo admin account và redeploy cùng env không tạo duplicate.

---

## Phase 5: User Story 3 - Thay thế định danh admin MongoDB cũ khi triển khai lại (Priority: P3)

**Goal**: Khi đổi password hoặc đổi username trong env, deploy sẽ update hoặc thay thế đúng tài khoản admin do deployment quản lý.

**Independent Test**: Redeploy với cùng username nhưng password mới để xác nhận password được cập nhật; redeploy tiếp với username mới để xác nhận user cũ bị thay thế và user mới đăng nhập được.

### Implementation for User Story 3

- [X] T026 [US3] Cài đặt luồng update password khi username Mongo admin giữ nguyên tại `scripts/mongo/reconcile-admin.mjs`
- [X] T027 [US3] Cài đặt luồng thay thế user khi username Mongo admin thay đổi tại `scripts/mongo/reconcile-admin.mjs`
- [X] T028 [US3] Bổ sung xử lý fail-fast khi không thể update/xóa user cũ tại `scripts/mongo/reconcile-admin.mjs`
- [X] T029 [US3] Tích hợp kết quả reconcile thất bại để chặn trạng thái deploy thành công tại `scripts/mongo/run-reconcile.sh`
- [X] T030 [US3] Cập nhật hướng dẫn xác minh rotation password/username tại `specs/010-vps-deployment/quickstart.md`

**Checkpoint**: User Story 3 hoàn chỉnh khi cả đổi password và đổi username đều hoạt động idempotent và không để lại user cũ do deployment quản lý.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Hoàn thiện tài liệu, tính rõ ràng vận hành và xác minh toàn stack

- [X] T031 [P] Đồng bộ lại mô tả feature branch/các artefact nếu cần trong `specs/010-vps-deployment/plan.md`
- [X] T032 [P] Đồng bộ lại yêu cầu/giả định nếu implementation làm rõ thêm hành vi biên trong `specs/010-vps-deployment/spec.md`
- [X] T033 Cập nhật lệnh deploy chính thức, env cần thiết và lưu ý vận hành tại `README.md`
- [X] T034 Chạy xác minh quickstart end-to-end và ghi lại kết quả thực tế trong `specs/010-vps-deployment/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1: Setup**: Không phụ thuộc task nào khác
- **Phase 2: Foundational**: Phụ thuộc Phase 1
- **Phase 3: US1**: Phụ thuộc Phase 2
- **Phase 4: US2**: Phụ thuộc Phase 2; có thể bắt đầu sau foundation hoàn tất
- **Phase 5: US3**: Phụ thuộc Phase 4 vì cần mở rộng trên cùng logic reconciliation
- **Phase 6: Polish**: Phụ thuộc các user story đã chọn hoàn tất

### User Story Dependencies

- **US1**: Không phụ thuộc US2/US3, là MVP
- **US2**: Không phụ thuộc US1 về mặt nghiệp vụ, nhưng dùng chung deployment foundation
- **US3**: Phụ thuộc US2 vì mở rộng logic create/no-op sang update password và replace username

### Within Each User Story

- Cấu hình/chuẩn hóa runtime trước
- Tích hợp orchestration sau khi các file runtime sẵn sàng
- Tài liệu quickstart/contract cập nhật sau khi flow chính của story hoàn tất

### Parallel Opportunities

- T002, T003, T004 có thể làm song song sau T001
- T008 và T009 có thể làm song song sau T007
- T014-T018 có thể làm song song trong US1
- T031 và T032 có thể làm song song trong phase Polish

---

## Parallel Example: User Story 1

```bash
Task: "T014 [US1] Thay fallback API URL production trong frontend/src/pages/Login.jsx"
Task: "T015 [US1] Thay fallback API URL production trong frontend/src/pages/Dashboard.jsx"
Task: "T016 [US1] Thay fallback API URL production trong frontend/src/pages/Files.jsx"
Task: "T017 [US1] Thay fallback API URL production trong frontend/src/pages/Settings.jsx"
Task: "T018 [US1] Thay fallback API URL production trong frontend/src/context/NotificationsContext.jsx"
```

## Parallel Example: Setup Phase

```bash
Task: "T002 Tạo Dockerfile backend production tại backend/Dockerfile"
Task: "T003 Tạo Dockerfile frontend/reverse-proxy production tại frontend/Dockerfile"
Task: "T004 Tạo cấu hình reverse proxy public FE 5191 và BE 8010 tại deploy/nginx/default.conf"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Hoàn tất Phase 1: Setup
2. Hoàn tất Phase 2: Foundational
3. Hoàn tất Phase 3: User Story 1
4. Xác minh deploy mới trên VPS hoặc môi trường tương đương theo quickstart
5. Chỉ sau đó mới mở rộng sang Mongo admin reconciliation đầy đủ

### Incremental Delivery

1. Setup + Foundational → stack deploy cơ bản sẵn sàng
2. Thêm US1 → xác minh public FE/BE ports
3. Thêm US2 → xác minh create/no-op cho Mongo admin
4. Thêm US3 → xác minh password rotation và username replacement
5. Polish → đồng bộ tài liệu và kiểm chứng end-to-end

### Parallel Team Strategy

1. Một người xử lý root deployment assets (T001-T006)
2. Một người xử lý backend/frontend runtime config (T007-T009, T013-T019)
3. Một người xử lý Mongo reconciliation scripts (T010-T012, T021-T029)

---

## Notes

- Tổng số task: 34
- Tất cả task đều theo đúng format checklist `- [ ] Txxx ...`
- Các task có `[P]` đều là task khác file hoặc có thể làm độc lập
- Mỗi task user story đều có nhãn `[US1]`, `[US2]`, `[US3]`
- MVP được khuyến nghị là hoàn tất qua hết US1 trước
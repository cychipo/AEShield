# Research: Bộ triển khai VPS

## Decision 1: Triển khai bằng một stack container duy nhất từ thư mục gốc

- **Decision**: Dùng một stack triển khai từ thư mục gốc repository để build frontend, build backend và chạy MongoDB cùng nhau bằng một lệnh vận hành.
- **Rationale**: Repo đã tách `frontend/` và `backend/`, nhưng backend hiện là nơi phục vụ API và cũng có khả năng phục vụ static frontend sau khi build qua `FRONTEND_DIST` trong [backend/cmd/main.go:24-38](../../../backend/cmd/main.go#L24-L38) và [backend/cmd/main.go:117-125](../../../backend/cmd/main.go#L117-L125). Cách này phù hợp nhất với yêu cầu “một lệnh deploy” trên một Ubuntu VPS đơn lẻ.
- **Alternatives considered**:
  - Chạy frontend dev server và backend riêng lẻ trên VPS: không phù hợp production, phụ thuộc nhiều tiến trình hơn.
  - Tách frontend và backend thành hai dịch vụ public độc lập ngay từ đầu: vẫn khả thi nhưng làm tăng độ phức tạp vận hành so với mô hình backend phục vụ frontend đã có sẵn trong repo.

## Decision 2: Dùng reverse proxy làm điểm public duy nhất cho 2 cổng ngoài

- **Decision**: Đặt một reverse proxy đứng trước stack để public cổng `5191` cho frontend và `8010` cho backend, trong khi backend app nội bộ vẫn phục vụ frontend static và API.
- **Rationale**: Frontend hiện đang build thành static assets, còn backend phục vụ SPA và API. Tuy nhiên spec yêu cầu đồng thời public FE ở `5191` và BE ở `8010`, trong khi code hiện chỉ có một HTTP server theo `PORT` trong [backend/internal/config/config.go:32-50](../../../backend/internal/config/config.go#L32-L50). Reverse proxy là cách phù hợp nhất để thỏa yêu cầu port public mà không phải đổi kiến trúc lõi của ứng dụng.
- **Alternatives considered**:
  - Chỉ public backend một cổng và để frontend đi cùng cùng origin: đơn giản hơn nhưng không đáp ứng yêu cầu port FE `5191` riêng.
  - Chạy thêm một web server chỉ để phục vụ frontend build: đáp ứng được nhưng tạo thêm bề mặt triển khai không cần thiết nếu reverse proxy đã tồn tại.

## Decision 3: Giữ frontend build-time API URL trỏ về backend public port

- **Decision**: Khi build frontend cho môi trường VPS, đặt cấu hình API của frontend trỏ về backend public port `8010`.
- **Rationale**: Frontend hiện fallback về `http://localhost:6888/api/v1` trong nhiều màn hình, gồm [frontend/src/pages/Login.jsx](../../../frontend/src/pages/Login.jsx), [frontend/src/pages/Dashboard.jsx](../../../frontend/src/pages/Dashboard.jsx), [frontend/src/pages/Files.jsx](../../../frontend/src/pages/Files.jsx), [frontend/src/pages/Settings.jsx](../../../frontend/src/pages/Settings.jsx), và [frontend/src/context/NotificationsContext.jsx](../../../frontend/src/context/NotificationsContext.jsx). Với deploy VPS, các giá trị này cần được điều khiển bằng environment khi build để không trỏ sai cổng mặc định cũ.
- **Alternatives considered**:
  - Giữ nguyên fallback localhost cũ: không phù hợp production.
  - Dùng path tương đối hoàn toàn cho mọi API: tốt nếu frontend và backend cùng origin, nhưng không khớp với yêu cầu public FE/BE trên 2 cổng riêng.

## Decision 4: Đối soát Mongo admin bằng một tác vụ reconciliation chạy mỗi lần deploy

- **Decision**: Thực hiện đối soát tài khoản admin MongoDB bằng một tác vụ riêng chạy trong stack mỗi lần deploy, thay vì phụ thuộc vào cơ chế init một lần của image MongoDB.
- **Rationale**: Mongo image chỉ xử lý `MONGO_INITDB_ROOT_USERNAME` và `MONGO_INITDB_ROOT_PASSWORD` ở lần khởi tạo đầu tiên khi data directory rỗng; nó không tự reconcile lại khi redeploy. Điều này không đáp ứng các yêu cầu FR-006 đến FR-009 trong spec. Một tác vụ reconciliation chạy mỗi lần deploy mới có thể bảo đảm idempotent cho các trường hợp giữ nguyên, đổi mật khẩu hoặc đổi username.
- **Alternatives considered**:
  - Chỉ dùng `/docker-entrypoint-initdb.d/`: không xử lý được redeploy.
  - Nhúng toàn bộ logic đối soát vào backend startup: làm backend phụ thuộc chặt vào quyền quản trị Mongo và tăng rủi ro khởi động ứng dụng bị chặn bởi tác vụ hạ tầng.

## Decision 5: Đánh dấu tài khoản admin do deployment quản lý để tránh đụng tài khoản thủ công

- **Decision**: Ghi dấu rõ ràng tài khoản admin do deployment quản lý để chỉ đối soát/xóa đúng tài khoản do stack này tạo ra.
- **Rationale**: Spec yêu cầu xóa tài khoản cũ khi đổi username, nhưng assumptions cũng nêu rõ các tài khoản MongoDB không do deployment quản lý nằm ngoài phạm vi. Vì vậy logic reconciliation phải có cơ chế nhận diện tài khoản do deployment quản lý để không xóa nhầm user thủ công.
- **Alternatives considered**:
  - Xem mọi admin user là do deployment quản lý: nguy hiểm, dễ xóa nhầm.
  - Không lưu metadata nhận diện: khó xác định user cũ nào cần thay thế.

## Decision 6: Fail-fast nếu thiếu env bắt buộc hoặc nếu cổng public không bind được

- **Decision**: Quy trình deploy phải dừng rõ ràng nếu thiếu env quản trị Mongo hoặc nếu cổng `5191` / `8010` không khả dụng.
- **Rationale**: Spec yêu cầu tránh trạng thái “thành công một phần nhưng không rõ ràng” và yêu cầu thông báo lỗi rõ ràng cho operator. Điều này cần được phản ánh cả ở orchestration lẫn tài liệu quickstart.
- **Alternatives considered**:
  - Tự fallback sang giá trị mặc định: dễ gây deploy sai môi trường.
  - Cho phép stack lên một phần rồi để operator tự xử lý: không đạt yêu cầu spec.

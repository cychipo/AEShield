# Feature Specification: Job Queue for File Encryption and Decryption

**Feature Branch**: `011-job-queue`  
**Created**: 2026-04-12  
**Status**: Draft  
**Input**: User description: "vì tôi thấy thời gian mã hóa sẽ phụ thuộc vào độ lớn của file mà nếu chờ đến khi api upload thành công thì nó sẽ có trường hợp lỗi timeout, nên tôi muốn xử lý job queue chỗ này, tôi muốn khi upload file sẽ tạo một job id và check job status cho tới khi hoàn thành, ở ui sẽ hiện ui progress thay vì text loading bình thường, ngoài ra ở bước giải mã tôi cũng muốn nó là dạng job queue với y hệt như trên, UI cũng là progress"

## User Scenarios & Testing

### User Story 1 - Upload File with Progress (Priority: P1)

Người dùng tải file lên hệ thống và xem tiến trình mã hóa theo thời gian thực.

**Why this priority**: Đây là luồng chính của ứng dụng, người dùng cần biết file đang được xử lý đến đâu để tránh confusion khi upload lâu.

**Independent Test**: Có thể test bằng cách upload file bất kỳ và quan sát progress bar từ 0-100%.

**Acceptance Scenarios**:

1. **Given** người dùng chọn file và nhập mật khẩu, **When** bấm "Tải lên", **Then** hiện progress bar bắt đầu từ 0% và tăng dần theo thời gian
2. **Given** file đang upload, **When** mã hóa hoàn tất 100%, **Then** chuyển sang trạng thái hoàn thành và hiển thị thông báo thành công
3. **Given** quá trình upload bị lỗi (timeout, network error), **When** lỗi xảy ra, **Then** hiển thị thông báo lỗi rõ ràng và cho phép thử lại
4. **Given** người dùng đóng trình duyệt trong khi upload, **When** mở lại ứng dụng, **Then** job vẫn tiếp tục chạy ở backend và có thể kiểm tra status

---

### User Story 2 - Decrypt File with Progress (Priority: P1)

Người dùng giải mã file và xem tiến trình giải mã theo thời gian thực.

**Why this priority**: Giải mã file lớn cũng tốn thời gian, người dùng cần feedback rõ ràng.

**Independent Test**: Có thể test bằng cách giải mã file đã mã hóa và quan sát progress bar.

**Acceptance Scenarios**:

1. **Given** người dùng nhập mật khẩu để giải mã, **When** bấm "Giải mã", **Then** hiện progress bar từ 0% đến khi hoàn thành
2. **Given** giải mã thành công, **Then** hiển thị preview/download file đã giải mã
3. **Given** sai mật khẩu, **Then** hiển thị thông báo lỗi ngay lập tức mà không cần đợi hết quá trình giải mã
4. **Given** file bị hỏng hoặc không đúng format, **Then** hiển thị thông báo lỗi rõ ràng

---

### User Story 3 - Check Job Status (Priority: P2)

Người dùng có thể kiểm tra trạng thái của các job đang chạy hoặc đã hoàn thành.

**Why this priority**: Cho phép người dùng theo dõi tiến trình và xem lịch sử xử lý.

**Independent Test**: Có thể test bằng cách upload file và kiểm tra job status qua API.

**Acceptance Scenarios**:

1. **Given** người dùng có job đang chạy, **When** gọi API kiểm tra status, **Then** trả về trạng thái hiện tại (pending/processing/completed/failed) và percentage
2. **Given** job hoàn thành, **When** kiểm tra status, **Then** trả về kết quả hoặc link download
3. **Given** job thất bại, **When** kiểm tra status, **Then** trả về thông báo lỗi chi tiết

---

### User Story 4 - Job Queue Management (Priority: P3)

Hệ thống quản lý nhiều job đồng thời và xử lý queue hiệu quả.

**Why this priority**: Đảm bảo hệ thống không bị quá tải khi nhiều người dùng upload cùng lúc.

**Independent Test**: Test bằng cách gửi nhiều request upload song song.

**Acceptance Scenarios**:

1. **Given** nhiều user upload file cùng lúc, **When** hệ thống nhận request, **Then** mỗi job được gán unique ID và xử lý theo thứ tự
2. **Given** job bị timeout hoặc bị hủy, **When** kiểm tra status, **Then** trả về trạng thái cancelled/failed và cleanup tài nguyên

---

### Edge Cases

- Người dùng refresh trang trong khi job đang chạy → job vẫn tiếp tục ở backend
- Mất kết nối mạng trong khi upload → job vẫn tiếp tục, có thể resume hoặc retry
- File quá lớn (>1GB) → vẫn xử lý được với chunked processing
- Server bị restart trong khi job đang chạy → job được retry tự động
- Người dùng upload trùng file → xử lý như job riêng biệt

## Requirements

### Functional Requirements

- **FR-001**: System MUST tạo unique job ID khi nhận request upload file
- **FR-002**: System MUST cho phép kiểm tra trạng thái job qua job ID
- **FR-003**: System MUST trả về percentage hoàn thành cho mỗi job
- **FR-004**: Users MUST có thể xem progress bar real-time trên UI
- **FR-005**: System MUST hỗ trợ job queue với nhiều job đồng thời
- **FR-006**: System MUST xử lý timeout gracefully và cho phép retry
- **FR-007**: System MUST lưu trữ kết quả job để có thể retrieve sau
- **FR-008**: System MUST cleanup job đã hoàn thành sau một thời gian (garbage collection)
- **FR-009**: Users MUST có thể hủy job đang chạy
- **FR-010**: System MUST xử lý giải mã theo job queue với progress tracking
- **FR-011**: System MUST thông báo lỗi ngay lập tức khi sai mật khẩu (không cần đợi hết quá trình)

### Key Entities

- **Job**: Đại diện cho một tác vụ xử lý file (upload/encrypt hoặc decrypt), chứa job ID, status, percentage, kết quả/lỗi
- **JobStatus**: Enum các trạng thái job (pending, processing, completed, failed, cancelled)
- **JobResult**: Kết quả của job thành công (file metadata, download URL)
- **JobError**: Thông tin lỗi khi job thất bại (error code, message, timestamp)

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can view real-time progress with percentage updates at least every 1 second during processing
- **SC-002**: Upload jobs larger than 100MB complete successfully without timeout errors
- **SC-003**: 95% of jobs complete within reasonable time based on file size (linear with file size, not exponential)
- **SC-004**: Users can continue interacting with application while large files are processing in background
- **SC-005**: System can handle at least 10 concurrent jobs without performance degradation
- **SC-006**: Users receive clear error messages within 5 seconds when decryption fails due to wrong password
- **SC-007**: Job status remains queryable for at least 24 hours after completion

## Assumptions

- Người dùng có kết nối internet ổn định để upload file
- Server có đủ tài nguyên để xử lý job queue (CPU, memory, storage)
- Redis hoặc database hiện có có thể dùng để lưu trữ job status
- Người dùng sử dụng trình duyệt hiện đại hỗ trợ WebSocket hoặc polling cho real-time updates
- Ứng dụng đã có cơ chế authentication để xác định user sở hữu job
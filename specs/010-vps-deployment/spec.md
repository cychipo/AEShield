# Feature Specification: Bộ triển khai VPS

**Feature Branch**: `[010-vps-deployment]`  
**Created**: 2026-04-12  
**Status**: Draft  
**Input**: User description: "hiện tại tôi muốn có thể deploy lên vps ubuntu, tôi muốn có docker file, chỉ cần chạy một lệnh là public cả port fe và be, be là port 8010, fe là port 5191, sẽ có container cho db mongo luôn, có thể cònig tài khoản auth admin cho db đó qua env khi deploy, nếu mà key env đó đúng auth ( tồn tjai ) thì không tạo lại nữa, nếu mà khác mật khẩu thì update mật khẩu, còn khác tài khoanr thì xóa tk cũ và thay thế bằng tài khoản mới, hãy tạo spec cho tôi nhé"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Triển khai toàn bộ hệ thống trên Ubuntu VPS mới (Priority: P1)

Với vai trò là người vận hành, tôi muốn khởi động frontend, backend và cơ sở dữ liệu cùng nhau chỉ bằng một lệnh triển khai để một Ubuntu VPS mới có thể công khai ứng dụng ngay mà không cần thiết lập thủ công từng dịch vụ.

**Why this priority**: Đây là kết quả triển khai cốt lõi. Nếu không có luồng khởi động bằng một lệnh, gói triển khai sẽ không giải quyết được nhu cầu vận hành chính.

**Independent Test**: Có thể kiểm thử độc lập bằng cách chuẩn bị một Ubuntu VPS sạch, vào thư mục `deploy/`, cung cấp các giá trị môi trường triển khai, chạy lệnh khởi động đã được tài liệu hóa một lần, rồi xác nhận frontend truy cập được qua cổng 5191, API backend phản hồi qua cùng origin frontend tại `/api/v1/...`, và dịch vụ cơ sở dữ liệu đang chạy.

**Acceptance Scenarios**:

1. **Given** một Ubuntu VPS sạch với runtime cần thiết đã được cài đặt và các giá trị môi trường triển khai đã được chuẩn bị, **When** người vận hành chạy lệnh triển khai, **Then** frontend, backend và cơ sở dữ liệu khởi động thành công như một hệ thống thống nhất.
2. **Given** hệ thống đã khởi động thành công, **When** người vận hành kiểm tra các dịch vụ có thể truy cập từ bên ngoài, **Then** frontend có thể truy cập công khai qua cổng 5191 và các API backend phản hồi qua cùng origin frontend tại đường dẫn `/api/v1/...`.
3. **Given** hệ thống đã được triển khai, **When** người vận hành chạy lại cùng một lệnh triển khai với các giá trị môi trường không đổi, **Then** việc triển khai hoàn tất mà không cần dọn dẹp thủ công hoặc phát sinh bước thiết lập trùng lặp.

---

### User Story 2 - Cấu hình tài khoản admin MongoDB qua biến môi trường triển khai (Priority: P2)

Với vai trò là người vận hành, tôi muốn thông tin xác thực quản trị MongoDB được điều khiển bằng biến môi trường triển khai để có thể khởi tạo hoặc xoay vòng quyền truy cập cơ sở dữ liệu mà không cần vào container thủ công.

**Why this priority**: Quyền truy cập cơ sở dữ liệu phải có thể tái lập và được điều khiển bằng môi trường để việc triển khai VPS luôn dễ quản lý và an toàn.

**Independent Test**: Có thể kiểm thử độc lập bằng cách triển khai với username và password admin đã chọn, sau đó xác minh rằng các thông tin xác thực đó đăng nhập được và không cần thêm bước thiết lập người dùng cơ sở dữ liệu thủ công nào khác.

**Acceptance Scenarios**:

1. **Given** chưa tồn tại tài khoản quản trị MongoDB khớp với cấu hình yêu cầu, **When** người vận hành triển khai với các biến môi trường thông tin xác thực admin, **Then** hệ thống sẽ khởi tạo tài khoản quản trị đó và làm cho tài khoản này có thể dùng để xác thực cơ sở dữ liệu.
2. **Given** một tài khoản quản trị MongoDB đã tồn tại và khớp với username và password được cung cấp, **When** người vận hành triển khai lại với cùng các giá trị đó, **Then** tài khoản hiện có được giữ nguyên mà không tạo tài khoản trùng lặp.
3. **Given** username quản trị được cung cấp đã tồn tại nhưng password được cung cấp khác với password hiện tại, **When** người vận hành triển khai, **Then** password của tài khoản hiện có được cập nhật để khớp với các giá trị môi trường.

---

### User Story 3 - Thay thế định danh admin MongoDB cũ khi triển khai lại (Priority: P3)

Với vai trò là người vận hành, tôi muốn khi username admin MongoDB thay đổi thì định danh quản trị cũ phải được thay thế để mô hình truy cập cơ sở dữ liệu luôn đồng bộ với cấu hình triển khai hiện tại.

**Why this priority**: Việc xoay vòng thông tin xác thực sẽ chưa hoàn chỉnh nếu việc đổi tên tài khoản quản trị vẫn để lại định danh cũ với quyền cao trong hệ thống.

**Independent Test**: Có thể kiểm thử độc lập bằng cách triển khai lần đầu với một username admin, triển khai lại với một username admin khác, rồi xác nhận rằng tài khoản admin trước đó đã bị xóa và tài khoản admin mới trở thành định danh quản trị đang hoạt động.

**Acceptance Scenarios**:

1. **Given** cơ sở dữ liệu hiện có một tài khoản quản trị được tạo từ lần triển khai trước, **When** người vận hành triển khai lại với một username quản trị khác, **Then** tài khoản quản trị cũ do quá trình triển khai quản lý sẽ bị xóa và được thay thế bằng tài khoản mới.
2. **Given** username quản trị thay đổi trong quá trình triển khai lại, **When** việc triển khai hoàn tất, **Then** username cũ không còn đăng nhập được và username mới có thể đăng nhập thành công.

---

### Edge Cases

- Điều gì xảy ra khi cơ sở dữ liệu khởi động chậm hơn các dịch vụ ứng dụng nhưng lệnh triển khai chỉ được chạy một lần?
- Hệ thống xử lý thế nào khi các biến môi trường thông tin xác thực admin bị thiếu, rỗng hoặc chỉ được cung cấp một phần?
- Việc triển khai lại sẽ hoạt động ra sao nếu không thể xóa sạch tài khoản admin cũ do deployment quản lý?
- Điều gì xảy ra nếu các cổng công khai 5191 hoặc 8010 đã bị dịch vụ khác sử dụng trên VPS?
- Hệ thống triển khai sẽ hoạt động thế nào khi các dịch vụ ứng dụng khởi động lại sau khi tài khoản cơ sở dữ liệu đã được khởi tạo trước đó?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Hệ thống MUST cung cấp một lệnh triển khai duy nhất để khởi động frontend, backend và cơ sở dữ liệu như một hệ thống triển khai thống nhất trên môi trường Ubuntu VPS.
- **FR-002**: Hệ thống MUST công khai frontend qua cổng 5191 sau khi triển khai thành công.
- **FR-003**: Hệ thống MUST cho phép frontend production gọi backend qua same-origin proxy tại đường dẫn `/api/v1/...` mà không phụ thuộc backend public URL riêng.
- **FR-004**: Gói triển khai MUST bao gồm một dịch vụ MongoDB khởi động cùng hệ thống ứng dụng và sẵn sàng cho backend sử dụng.
- **FR-005**: Hệ thống MUST cho phép cung cấp username và password quản trị MongoDB thông qua các biến môi trường triển khai, cùng với thông tin bootstrap root cần thiết để tác vụ đối soát có thể chạy an toàn.
- **FR-006**: Nếu tài khoản quản trị MongoDB do deployment quản lý đã tồn tại và thông tin xác thực của nó khớp với các giá trị môi trường được cung cấp, hệ thống MUST giữ nguyên tài khoản đó.
- **FR-007**: Nếu tài khoản quản trị MongoDB do deployment quản lý đã tồn tại với cùng username nhưng password khác, hệ thống MUST cập nhật password để tài khoản khớp với các giá trị môi trường được cung cấp.
- **FR-008**: Nếu username quản trị MongoDB được cung cấp khác với username quản trị hiện có do deployment quản lý, hệ thống MUST xóa tài khoản quản trị cũ do deployment quản lý và thay thế bằng một tài khoản quản trị mới khớp với các giá trị môi trường được cung cấp.
- **FR-009**: Hệ thống MUST bảo đảm rằng việc triển khai lặp lại với cùng các giá trị môi trường là an toàn và không tạo ra các tài khoản quản trị MongoDB trùng lặp do deployment quản lý.
- **FR-010**: Hệ thống MUST thất bại với thông báo lỗi rõ ràng cho người vận hành khi các biến môi trường bắt buộc cho quyền quản trị MongoDB bị thiếu hoặc không hợp lệ.
- **FR-011**: Hệ thống MUST bảo toàn khả năng xác thực với MongoDB bằng tài khoản quản trị do deployment quản lý sau lần triển khai đầu tiên và sau mỗi lần thay đổi thông tin xác thực.
- **FR-012**: Quy trình triển khai MUST xác định rõ cách xử lý khi các cổng công khai được yêu cầu không khả dụng, bao gồm việc ngăn trạng thái thành công một phần nhưng không rõ ràng.

### Key Entities *(include if feature involves data)*

- **Deployment Stack**: Gói hệ thống đầy đủ có thể triển khai lên VPS, bao gồm frontend, backend, cơ sở dữ liệu, thư mục vận hành `deploy/`, các giá trị môi trường, cổng công khai và hành vi khởi động.
- **Deployment Command**: Hành động duy nhất do người vận hành kích hoạt để khởi tạo hoặc cập nhật hệ thống triển khai.
- **Deployment-Managed MongoDB Admin Account**: Định danh quản trị MongoDB có username và password được điều khiển bởi các biến môi trường triển khai và được đối soát ở mỗi lần triển khai.
- **Credential Configuration**: Tập hợp các giá trị môi trường triển khai dùng để xác định quyền truy cập quản trị MongoDB cho hệ thống.
- **Port Exposure Rule**: Quy ước về khả năng truy cập công khai mong đợi đối với frontend trên cổng 5191 và việc backend được truy cập từ client qua same-origin proxy `/api`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Trên một Ubuntu VPS sạch, người vận hành có thể đưa toàn bộ hệ thống lên hoạt động chỉ bằng một lệnh triển khai đã được tài liệu hóa mà không cần khởi động thủ công từng dịch vụ riêng lẻ.
- **SC-002**: Sau khi triển khai hoàn tất, người dùng có thể truy cập frontend qua cổng 5191 và nhận phản hồi API thành công qua same-origin endpoint `/api/v1/...` ngay trong lần xác minh đầu tiên.
- **SC-003**: Trong các lần triển khai lặp lại với cấu hình thông tin xác thực không đổi, 100% lần triển khai lại hoàn tất mà không tạo ra tài khoản quản trị MongoDB trùng lặp do deployment quản lý.
- **SC-004**: Trong các bài kiểm thử xoay vòng thông tin xác thực mà chỉ thay đổi password, password mới trở thành password duy nhất đăng nhập được cho tài khoản quản trị MongoDB do deployment quản lý sau khi triển khai lại.
- **SC-005**: Trong các bài kiểm thử xoay vòng thông tin xác thực mà username admin thay đổi, username cũ do deployment quản lý không còn đăng nhập được sau khi triển khai lại và username mới đăng nhập thành công.
- **SC-006**: Người vận hành nhận được kết quả thất bại rõ ràng mỗi khi các biến môi trường bắt buộc cho MongoDB bị thiếu hoặc không hợp lệ, và không có trạng thái admin được cấu hình dở dang nhưng vẫn báo thành công.

## Assumptions

- Môi trường triển khai mục tiêu là một Ubuntu VPS đơn lẻ chạy frontend, backend và MongoDB cùng nhau, với thư mục vận hành riêng `deploy/`.
- Người vận hành chịu trách nhiệm cung cấp các giá trị môi trường triển khai hợp lệ trước khi khởi động hệ thống.
- Quyền truy cập công khai vào cổng 5191 được cho phép bởi cấu hình mạng và firewall của VPS tại thời điểm triển khai; backend được truy cập từ phía client qua frontend origin.
- Chỉ có một tài khoản quản trị MongoDB do deployment quản lý cần được đối soát trong phạm vi tính năng này.
- Các tài khoản MongoDB hiện có nhưng không do deployment quản lý nằm ngoài phạm vi của tính năng này, trừ khi chúng được xác định rõ là tài khoản quản trị cũ do deployment quản lý.
- Tính năng này chỉ bao phủ hành vi triển khai và triển khai lại trên môi trường VPS, không bao gồm điều phối nhiều node, clustering hoặc dịch vụ cơ sở dữ liệu được quản lý sẵn.

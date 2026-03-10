# Đặc tả Xác thực

## Nhà cung cấp OAuth2

- **Google**
- **GitHub**

## Luồng Xác thực

1. **Frontend** bắt đầu luồng OAuth2 bằng cách chuyển hướng người dùng đến URL ủy quyền của nhà cung cấp
2. Người dùng cấp quyền
3. Nhà cung cấp chuyển hướng về Frontend với **Authorization Code**
4. Frontend gửi mã về Backend
5. Backend đổi mã lấy **Access Token** từ nhà cung cấp
6. Backend tạo **JWT (JSON Web Token)** cho phiên AEShield
7. Backend trả JWT về Frontend

## Cấu trúc JWT

```go
type Claims struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    Provider  string `json:"provider"` // "google" hoặc "github"
    ExpiresAt int64  `json:"exp"`
}
```

## Các Endpoint (base path: /api/v1)

| Method | Endpoint     | Mô tả                 |
| ------ | ------------ | --------------------- |
| POST   | /auth/google | Đổi mã Google lấy JWT |
| POST   | /auth/github | Đổi mã GitHub lấy JWT |

## Lưu trữ Token

- JWT được lưu ở Frontend (localStorage hoặc httpOnly cookie)
- Gửi trong header `Authorization: Bearer <token>` cho các yêu cầu được bảo vệ

## File Mockup cho spec này:

```
/mockups/auth-login.html
/mockups/terms-of-service.html
```

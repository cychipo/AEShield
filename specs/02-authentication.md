# Đặc tả Xác thực

## Nhà cung cấp OAuth2

- **Google**
- **GitHub**

Không có đăng ký thủ công. Toàn bộ user được tạo/cập nhật tự động qua OAuth.

## Luồng Xác thực (Web)

```
1. Frontend gọi GET /api/v1/auth/urls → nhận Google URL & GitHub URL
2. Người dùng click → trình duyệt chuyển hướng đến trang OAuth của provider
3. Provider redirect về /auth/google/callback hoặc /auth/github/callback
   (đây là HTML page trung gian, không phải API)
4. HTML page gọi fetch('/api/v1/auth/google/callback?code=...')
5. Backend đổi code lấy Access Token từ provider
6. Backend tìm hoặc tạo User (merge theo email nếu đã tồn tại)
7. Backend tạo JWT và trả về { token, user }
8. Frontend lưu token vào localStorage, redirect → /dashboard
```

**Lý do dùng HTML trung gian:** Authorization Code chỉ dùng được 1 lần. Nếu redirect thẳng về frontend SPA thì React Router sẽ re-render trước khi fetch API kịp chạy, dẫn đến code bị consumed mà không xử lý được.

## Merge Tài khoản

Nếu Google và GitHub login cùng email → **merge thành 1 User duy nhất**, thêm provider mới vào mảng `providers`.

## Cấu trúc JWT

```go
type Claims struct {
    UserID string `json:"user_id"`
    Email  string `json:"email"`
    Avatar string `json:"avatar"`
    Name   string `json:"name"`
    jwt.RegisteredClaims
}
```

## Các Endpoint

| Method | Endpoint | Mô tả |
|--------|----------|-------|
| `GET` | `/api/v1/auth/urls` | Trả về Google URL & GitHub URL để redirect |
| `GET` | `/api/v1/auth/google` | Redirect trình duyệt đến Google OAuth |
| `GET` | `/api/v1/auth/google/callback` | Đổi code lấy JWT (API, gọi từ HTML trung gian) |
| `GET` | `/api/v1/auth/github` | Redirect trình duyệt đến GitHub OAuth |
| `GET` | `/api/v1/auth/github/callback` | Đổi code lấy JWT (API, gọi từ HTML trung gian) |
| `GET` | `/api/v1/auth/me` | Lấy thông tin user hiện tại (JWT required) |
| `GET` | `/auth/google/callback` | HTML trung gian (nhận redirect từ Google) |
| `GET` | `/auth/github/callback` | HTML trung gian (nhận redirect từ GitHub) |

## Lưu trữ Token

- JWT lưu ở `localStorage` với key `aeshield_token`
- User info lưu ở `localStorage` với key `aeshield_user`
- Gửi trong header `Authorization: Bearer <token>` cho mọi request có bảo vệ

## Biến Môi trường

```env
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:3000/auth/google/callback

GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=
GITHUB_REDIRECT_URL=http://localhost:3000/auth/github/callback

JWT_SECRET=
JWT_EXPIRY=168h
```

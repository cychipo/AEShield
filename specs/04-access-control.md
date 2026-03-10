# Đặc tả Quản lý Quyền truy cập

## Chế độ truy cập

### 1. Công khai (Public)
- Tạo hash công khai (CID)
- Ai có link đều có thể tải
- Không cần xác thực

### 2. Riêng tư (Private)
- Chỉ chủ sở hữu được truy cập
- OwnerID trong JWT phải trùng với ownerID của file
- Trả về 403 Forbidden cho người không phải chủ sở hữu

### 3. Danh sách trắng (Whitelist)
- Giới hạn cho người dùng cụ thể
- Trường `allowed_users` chứa danh sách email hoặc user ID
- Yêu cầu xác thực
- Backend xác minh danh tính người dùng trước khi tạo presigned URL

## Luồng kiểm tra quyền

```
1. Người dùng yêu cầu tải file
2. Backend tra cứu metadata file trong MongoDB
3. Kiểm tra access_mode:
   - Public: Tạo presigned URL (không kiểm tra auth)
   - Private: Xác minh JWT.user_id == file.owner_id
   - Whitelist: Xác minh JWT.email hoặc JWT.user_id trong whitelist
4. Nếu được phép: Tạo và trả về presigned URL
5. Nếu không được phép: Trả về 403 Forbidden
```

## Định dạng Link Công khai

```
https://r2-bucket.r2.cloudflarestorage.com/{storage_path}?{presigned_params}
```

hoặc dùng domain tùy chỉnh:

```
https://files.aeshield.com/{cid}
```

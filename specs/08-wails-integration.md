# Đặc tả Tích hợp Wails (Desktop App)

## Tổng quan

Sử dụng Wails để build ứng dụng desktop cross-platform với Go backend và web frontend.

## Window Management

### Main Window

- **Kích thước mặc định:** 1200x800px
- **Kích thước tối thiểu:** 800x600px
- **Resizable:** Có
- **Title:** "AEShield"
- **Background color:** `#F9F7F2`

### Window Events

| Event | Handler |
|-------|---------|
| OnStart | Khởi tạo app, load config |
| OnShutdown | Cleanup resources |
| OnQuit | Lưu trạng thái trước khi thoát |

---

## System Tray (Menu Bar)

### Tray Icon

- **macOS:** Menu bar icon (16x16, 32x32 @2x)
- **Windows:** System tray icon

### Tray Menu

```
AEShield
├── Hiển thị cửa sổ      # Show Window
├── ─────────────────
├── Đăng nhập           # Login (nếu chưa login)
├── Quản lý File        # File Manager
├── ─────────────────
├── Cài đặt             # Settings
├── Thoát               # Quit
```

### Tray Events

- **Left click:** Hiển thị/ẩn cửa sổ
- **Right click:** Hiển thị menu

---

## Native Dialogs

### File Picker

```go
// Mở file picker để chọn file cần mã hóa
dialog.ShowOpenFile Dialog {
    Title: "Chọn file",
    Filters: []Filter{
        {Name: "All Files", Extensions: ["*"]},
    },
    Multiple: false,
}
```

### Save Dialog

```go
// Mở file picker để lưu file đã giải mã
dialog.ShowSaveFile Dialog {
    Title: "Lưu file",
    DefaultFilename: "decrypted_file",
}
```

### Message Box

```go
dialog.ShowMessageBox Dialog {
    Type:    "info", // "info", "warning", "error"
    Title:   "AEShield",
    Message: "File đã được mã hóa thành công!",
}
```

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| Cmd/Ctrl + O | Mở file |
| Cmd/Ctrl + S | Lưu/Tải lên |
| Cmd/Ctrl + , | Mở cài đặt |
| Cmd/Ctrl + Q | Thoát |
| Cmd/Ctrl + W | Đóng cửa sổ |
| Cmd/Ctrl + M | Minimize to tray |

---

## Clipboard

```go
// Copy presigned URL
clipboard.WriteText(url)

// Paste password hoặc text
password := clipboard.ReadText()
```

---

## Notifications

### System Notifications

```go
// Thông báo khi upload hoàn tất
runtime.Notification struct {
    Title:   "AEShield",
    Body:    "File đã được tải lên thành công",
}
```

---

## Storage (Local)

### App Data Location

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/AEShield` |
| Windows | `%APPDATA%/AEShield` |
| Linux | `~/.config/aeshield` |

### Local Storage

```go
// Lưu JWT token (encrypted)
store.Set("auth_token", encryptedToken)

// Lưu cài đặt người dùng
store.SetObject("settings", settings)

// Lấy cài đặt
settings := store.GetObject("settings", DefaultSettings{})
```

---

## Go ↔ Frontend Communication

### Gọi Go từ Frontend

```javascript
// Frontend (Vue/React/Svelte)
import { UploadFile, DownloadFile, GetFiles } from "@wails/go/app";

// Upload file
const result = await UploadFile(filePath, password, encryptionType);

// Download file
await DownloadFile(fileId, password, savePath);
```

### Events từ Go sang Frontend

```go
// Go
runtime.EventsEmit(ctx, "upload-progress", 50)

// Frontend
window.runtime.EventsOn("upload-progress", (progress) => {
    console.log(progress);
});
```

---

## Build Configuration

### Version Info (Windows)

```json
{
  "Name": "AEShield",
  "Version": "1.0.0",
  "Description": "Secure File Management Platform",
  "Author": "AEShield Team",
  "Icon": "build/icon.ico"
}
```

### Version Info (macOS)

```json
{
  "Name": "AEShield",
  "Version": "1.0.0",
  "Identifier": "com.aeshield.app",
  "Icon": "build/icon.icns"
}
```

---

## Auto-Update

```go
// Kiểm tra update khi khởi động
if runtime.NewerVersionAvailable() {
    // Hiển thị thông báo update
    dialog.ShowMessageBox("Có phiên bản mới!", "Bạn có muốn cập nhật không?")
}
```

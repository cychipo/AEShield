# Đặc tả Màu sắc & Giao diện

## Bảng Màu (Color Palette)

### Màu nền và bề mặt

| Tên | Giá trị | Sử dụng |
|-----|---------|----------|
| Background (Nền chính) | `#F9F7F2` | Nền toàn trang, màu giấy nhạt |
| Surface (Vùng làm việc) | `#FFFFFF` | Card, modal, vùng nhập liệu |
| Deep Tone (Màu bổ trợ) | `#3E3935` | Text chính, border, icons |

### Màu hành động

| Tên | Giá trị | Sử dụng |
|-----|---------|----------|
| Primary (Màu nhấn) | `#F6821F` | Buttons, links, hover states |
| Primary Hover | `#E57518` | Trạng thái hover của primary |
| Primary Light | `#FFF3E6` | Background nhẹ cho badges, tags |

### Màu trạng thái

| Tên | Giá trị | Sử dụng |
|-----|---------|----------|
| Success (Bảo mật/Thành công) | `#10B981` | Icon bảo mật, trạng thái thành công |
| Error (Lỗi) | `#EF4444` | Thông báo lỗi, icon lỗi |
| Warning (Cảnh báo) | `#F59E0B` | Cảnh báo, trạng thái chờ |
| Info (Thông tin) | `#3B82F6` | Thông báo, tips |

---

## Typography

### Font Family

- **Font chính:** Inter, system-ui, sans-serif
- **Font code:** JetBrains Mono, monospace (cho hiển thị hash, CID)

### Cỡ chữ

| Cỡ | Sử dụng |
|----|----------|
| 12px | Labels, captions |
| 14px | Body text |
| 16px | Subheadings |
| 20px | Page titles |
| 24px | Hero headings |
| 32px | Main headings |

---

## Spacing System

| Tên | Giá trị |
|-----|---------|
| xs | 4px |
| sm | 8px |
| md | 16px |
| lg | 24px |
| xl | 32px |
| 2xl | 48px |

---

## Border Radius

| Tên | Giá trị | Sử dụng |
|-----|---------|----------|
| sm | 4px | Buttons, inputs |
| md | 8px | Cards |
| lg | 12px | Modals |
| full | 9999px | Avatars, badges |

---

## Shadows

```css
--shadow-sm: 0 1px 2px rgba(62, 57, 53, 0.05);
--shadow-md: 0 4px 6px rgba(62, 57, 53, 0.07);
--shadow-lg: 0 10px 15px rgba(62, 57, 53, 0.1);
```

---

## Components

### Buttons

**Primary Button:**
- Background: `#F6821F`
- Text: `#FFFFFF`
- Hover: `#E57518`
- Border-radius: 4px
- Padding: 12px 24px

**Secondary Button:**
- Background: `#FFFFFF`
- Border: 1px solid `#3E3935`
- Text: `#3E3935`
- Hover: Background `#F9F7F2`

**Ghost Button:**
- Background: transparent
- Text: `#3E3935`
- Hover: Background `#F9F7F2`

### Input Fields

- Background: `#FFFFFF`
- Border: 1px solid `#E5E5E5`
- Border focus: 1px solid `#F6821F`
- Border-radius: 4px
- Padding: 12px 16px
- Placeholder color: `#9CA3AF`

### Cards

- Background: `#FFFFFF`
- Border-radius: 8px
- Shadow: `--shadow-md`
- Padding: 24px

### File Item

- Background: `#FFFFFF`
- Border: 1px solid `#E5E5E5`
- Border-radius: 8px
- Padding: 16px
- Hover: Border `#F6821F`

### Badges/Tags

**Security Badge:**
- Background: `#ECFDF5` (light green)
- Text: `#10B981`
- Icon: Shield check

**Encryption Badge:**
- Background: `#FFF3E6` (light orange)
- Text: `#F6821F`
- Icon: Lock

---

## Trạng thái Bảo mật (Security States)

### File đã mã hóa
- Icon: Shield check với màu `#10B981`
- Badge: "Đã bảo vệ" màu xanh bảo

### File công khai
- Icon: Globe
- Badge: "Công khai" màu `#3B82F6`

### File riêng tư
- Icon: Lock
- Badge: "Riêng tư" màu `#3E3935`

---

## Responsive Breakpoints

| Breakpoint | Width |
|------------|-------|
| Mobile | < 640px |
| Tablet | 640px - 1024px |
| Desktop | > 1024px |

---

## Animations

```css
--transition-fast: 150ms ease;
--transition-normal: 250ms ease;
--transition-slow: 350ms ease;
```

- Hover transitions: 150ms
- Page transitions: 250ms
- Modal open/close: 350ms

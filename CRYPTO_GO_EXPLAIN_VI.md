# Giải thích chi tiết file `backend/internal/crypto/crypto.go`

Tài liệu này giải thích chi tiết từng phần trong file [backend/internal/crypto/crypto.go](backend/internal/crypto/crypto.go), bao gồm mục đích tổng thể, luồng mã hóa/giải mã, ý nghĩa từng hàm, và diễn giải gần như từng dòng code.

---

## 1. Mục đích của file này

File này triển khai cơ chế:

- **mã hóa file theo kiểu streaming** để tránh phải nạp toàn bộ file vào RAM
- dùng **AES-GCM** để đảm bảo:
  - **bí mật dữ liệu**
  - **xác thực tính toàn vẹn** của từng khối dữ liệu
- dùng **Argon2id** để sinh khóa AES từ **password người dùng**

Nói ngắn gọn:
- đầu vào: `plaintext + password`
- đầu ra: `ciphertext` có định dạng riêng của hệ thống AEShield

Ngược lại khi giải mã:
- đầu vào: `ciphertext + password đúng`
- đầu ra: `plaintext` gốc

---

## 2. Phần comment đầu file

```go
// Package crypto cung cấp mã hóa/giải mã file streaming dùng AES-GCM + Argon2id.
```

- Dòng này mô tả package `crypto` có nhiệm vụ mã hóa/giải mã file.
- Từ **streaming** nghĩa là xử lý dữ liệu theo từng phần nhỏ, không cần đọc hết cả file vào bộ nhớ.
- AES-GCM là thuật toán mã hóa đối xứng có kèm xác thực dữ liệu.
- Argon2id là thuật toán dẫn xuất khóa từ password.

```go
//
// Định dạng ciphertext (prefix + data):
//
//	[4 bytes: magic "AES\x00"] [1 byte: key-bits / 64] [16 bytes: argon2 salt]
//	[12 bytes: GCM nonce] [N bytes: AES-GCM ciphertext+tag]
```

Comment này mô tả **header format** của file mã hóa.

Cụ thể:

1. **4 bytes magic**: `AES\x00`
   - Dùng để nhận diện đây có phải file mã hóa của hệ thống hay không.
   - Khi giải mã, nếu 4 byte đầu không khớp, hệ thống báo lỗi.

2. **1 byte key-bits / 64**
   - Không lưu trực tiếp `128`, `192`, `256`.
   - Thay vào đó lưu:
     - `2` cho AES-128
     - `3` cho AES-192
     - `4` cho AES-256
   - Khi đọc lại sẽ nhân với `64` để ra số bit thật.

3. **16 bytes salt**
   - Salt dùng cho Argon2id để dẫn xuất khóa từ password.
   - Salt ngẫu nhiên giúp cùng một password nhưng mỗi lần mã hóa vẫn cho ra key khác nhau.

4. **12 bytes GCM nonce**
   - Đây là nonce gốc (`baseNonce`).
   - Mỗi chunk sẽ sinh nonce riêng dựa trên nonce gốc + chỉ số chunk.

5. **ciphertext**
   - Dữ liệu sau khi mã hóa bằng AES-GCM.
   - Trong thực tế file này không chỉ có “một cục ciphertext duy nhất”, mà gồm nhiều chunk.
   - Mỗi chunk được ghi dưới dạng: `[4 bytes length][ciphertext_chunk]`

```go
// Streaming được thực hiện theo từng chunk CHUNK_SIZE (64 KiB) để không tràn RAM.
// Mỗi chunk được mã hóa độc lập với nonce riêng (nonce ban đầu + index chunk).
```

- File lớn sẽ được chia thành từng chunk 64 KiB.
- Mỗi chunk được mã hóa độc lập.
- Mỗi chunk có nonce riêng, tránh tái sử dụng cùng `(key, nonce)` — đây là yêu cầu rất quan trọng khi dùng GCM.

---

## 3. Khai báo package và import

```go
package crypto
```

- Tệp này thuộc package `crypto`.

```go
import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)
```

Giải thích từng import:

- `crypto/aes`
  - Cung cấp block cipher AES.
- `crypto/cipher`
  - Cung cấp các mode như GCM.
- `crypto/rand`
  - Sinh số ngẫu nhiên an toàn cho salt và nonce.
- `encoding/binary`
  - Ghi/đọc số nguyên theo dạng bytes, ở đây dùng BigEndian.
- `errors`
  - Tạo biến lỗi tĩnh.
- `fmt`
  - Format lỗi có ngữ cảnh.
- `io`
  - Đọc/ghi stream.
- `golang.org/x/crypto/argon2`
  - Dùng Argon2id để dẫn xuất key từ password.

---

## 4. Khối hằng số đầu tiên

```go
const (
	// ChunkSize là kích thước mỗi khối plaintext khi mã hóa streaming (64 KiB)
	ChunkSize = 64 * 1024

	// HeaderMagic giúp nhận dạng file đã mã hóa bởi AEShield
	HeaderMagic = "AES\x00"

	saltSize  = 16
	nonceSize = 12
)
```

### `ChunkSize = 64 * 1024`
- Mỗi lần đọc tối đa 64 KiB plaintext.
- 64 * 1024 = 65536 bytes.
- Đây là kích thước chunk trước khi mã hóa.

### `HeaderMagic = "AES\x00"`
- 4 byte đầu file mã hóa.
- Giá trị này giúp nhận biết file thuộc định dạng AEShield.
- `\x00` là byte null để đủ 4 byte.

### `saltSize = 16`
- Salt Argon2 dài 16 byte.

### `nonceSize = 12`
- AES-GCM chuẩn dùng nonce 12 byte.

---

## 5. Tham số Argon2id

```go
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
)
```

Các tham số này dùng khi gọi `argon2.IDKey(...)`.

### `argonTime = 1`
- Số vòng lặp tính toán.
- Càng lớn thì càng chậm hơn nhưng chống brute-force tốt hơn.

### `argonMemory = 64 * 1024`
- Dung lượng RAM Argon2 sử dụng, đơn vị là KiB.
- `64 * 1024 KiB = 64 MiB`.

### `argonThreads = 4`
- Số luồng song song Argon2 dùng.

Ý nghĩa tổng thể:
- Đây là các tham số kiểm soát độ khó khi biến password thành key.
- Mục tiêu là khiến việc brute-force password trở nên đắt hơn.

---

## 6. Các biến lỗi

```go
var (
	ErrInvalidMagic   = errors.New("crypto: invalid file magic, not an AEShield encrypted file")
	ErrInvalidKeyBits = errors.New("crypto: unsupported key size")
	ErrShortHeader    = errors.New("crypto: file header too short")
	ErrDecryptChunk   = errors.New("crypto: failed to decrypt chunk")
)
```

Các lỗi này được định nghĩa sẵn để tái sử dụng.

### `ErrInvalidMagic`
- Dùng khi 4 byte đầu file không phải `AES\x00`.
- Nghĩa là file không đúng định dạng hệ thống.

### `ErrInvalidKeyBits`
- Dùng khi giá trị key size không thuộc 128/192/256.

### `ErrShortHeader`
- Dùng khi file quá ngắn, không đủ header.

### `ErrDecryptChunk`
- Dùng khi giải mã một chunk thất bại.
- Thường do password sai, file hỏng, hoặc dữ liệu bị sửa đổi.

---

## 7. Kiểu `KeyBits`

```go
// KeyBits là số bit của khóa AES.
type KeyBits int
```

- Đây là kiểu dữ liệu riêng để biểu diễn độ dài khóa AES.
- Dùng kiểu riêng giúp code rõ nghĩa hơn thay vì dùng `int` trần.

```go
const (
	KeyBits128 KeyBits = 128
	KeyBits192 KeyBits = 192
	KeyBits256 KeyBits = 256
)
```

Ba giá trị hợp lệ:
- AES-128
- AES-192
- AES-256

---

## 8. Hàm `DeriveKey`

```go
// DeriveKey dùng Argon2id để chuyển đổi password + salt thành AES key.
func DeriveKey(password string, salt []byte, bits KeyBits) ([]byte, error) {
	keyLen, err := keyLen(bits)
	if err != nil {
		return nil, err
	}
	key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, uint32(keyLen))
	return key, nil
}
```

### Mục đích
Biến:
- `password`
- `salt`
- loại khóa `bits`

thành một **AES key** có độ dài đúng.

### Giải thích từng dòng

```go
keyLen, err := keyLen(bits)
```
- Gọi hàm `keyLen(...)` để đổi số bit thành số byte:
  - 128 -> 16
  - 192 -> 24
  - 256 -> 32

```go
if err != nil {
	return nil, err
}
```
- Nếu `bits` không hợp lệ thì trả lỗi luôn.

```go
key := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, uint32(keyLen))
```
- Dùng Argon2id để sinh key.
- `[]byte(password)` chuyển chuỗi password thành bytes.
- `salt` là giá trị ngẫu nhiên lưu trong header.
- `argonTime`, `argonMemory`, `argonThreads` là tham số bảo mật.
- `uint32(keyLen)` là số byte cần sinh ra.

```go
return key, nil
```
- Trả key vừa sinh.

### Ý nghĩa bảo mật
- Password không dùng trực tiếp làm key AES.
- Nó phải qua Argon2id để tạo key mạnh hơn và chống brute-force tốt hơn.

---

## 9. Hàm `Encrypt`

```go
// Encrypt đọc plaintext từ src, mã hóa streaming AES-GCM, ghi vào dst.
// Trả về số byte đã ghi vào dst.
func Encrypt(dst io.Writer, src io.Reader, password string, bits KeyBits) (int64, error) {
```

### Mục đích
- Đọc dữ liệu gốc từ `src`
- Mã hóa bằng AES-GCM
- Ghi dữ liệu mã hóa vào `dst`
- Trả lại tổng số byte đã ghi

### Tham số
- `dst io.Writer`: nơi ghi ciphertext ra
- `src io.Reader`: nơi đọc plaintext vào
- `password string`: mật khẩu người dùng
- `bits KeyBits`: loại AES muốn dùng (128/192/256)

---

### 9.1. Kiểm tra độ dài khóa

```go
	if _, err := keyLen(bits); err != nil {
		return 0, err
	}
```

- Xác nhận `bits` hợp lệ.
- Nếu không hợp lệ thì dừng sớm.

---

### 9.2. Sinh salt ngẫu nhiên

```go
	// Sinh salt ngẫu nhiên
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return 0, fmt.Errorf("crypto: generate salt: %w", err)
	}
```

- Tạo slice `salt` dài 16 byte.
- Dùng `rand.Reader` để lấy byte ngẫu nhiên an toàn.
- `io.ReadFull` đảm bảo đọc đủ 16 byte.
- Nếu lỗi, bọc lỗi với context rõ hơn.

---

### 9.3. Sinh nonce gốc ngẫu nhiên

```go
	// Sinh nonce gốc ngẫu nhiên
	baseNonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, baseNonce); err != nil {
		return 0, fmt.Errorf("crypto: generate nonce: %w", err)
	}
```

- Tạo nonce gốc 12 byte.
- Nonce này chưa dùng trực tiếp cho tất cả chunk.
- Nó là cơ sở để sinh nonce riêng cho từng chunk.

---

### 9.4. Dẫn xuất khóa AES từ password

```go
	key, err := DeriveKey(password, salt, bits)
	if err != nil {
		return 0, err
	}
```

- Dùng password + salt + key size để tạo key thật.
- Nếu Argon2 lỗi thì dừng.

---

### 9.5. Tạo AES block cipher

```go
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, fmt.Errorf("crypto: new cipher: %w", err)
	}
```

- Tạo AES block cipher từ key.
- Nếu key không đúng độ dài thì AES sẽ lỗi.

---

### 9.6. Bọc AES thành GCM

```go
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, fmt.Errorf("crypto: new GCM: %w", err)
	}
```

- Tạo AEAD theo mode GCM từ block cipher.
- Sau bước này có thể gọi:
  - `gcm.Seal(...)` để mã hóa
  - `gcm.Open(...)` để giải mã

---

### 9.7. Tạo header

```go
	// Ghi header: magic(4) + keyBitsByte(1) + salt(16) + baseNonce(12)
	header := make([]byte, 0, 4+1+saltSize+nonceSize)
	header = append(header, []byte(HeaderMagic)...)
	header = append(header, byte(int(bits)/64))
	header = append(header, salt...)
	header = append(header, baseNonce...)
```

Giải thích từng dòng:

```go
header := make([]byte, 0, 4+1+saltSize+nonceSize)
```
- Tạo slice rỗng nhưng cấp sẵn capacity.
- Tổng capacity = 4 + 1 + 16 + 12 = 33 bytes.

```go
header = append(header, []byte(HeaderMagic)...)
```
- Thêm 4 byte magic `AES\x00`.

```go
header = append(header, byte(int(bits)/64))
```
- Chuyển `128/192/256` thành `2/3/4`.
- Chỉ lưu 1 byte để tiết kiệm.

```go
header = append(header, salt...)
```
- Gắn salt vào header.

```go
header = append(header, baseNonce...)
```
- Gắn nonce gốc vào header.

---

### 9.8. Ghi header ra output

```go
	written, err := dst.Write(header)
	if err != nil {
		return int64(written), fmt.Errorf("crypto: write header: %w", err)
	}
	total := int64(written)
```

- Ghi 33 byte header ra `dst`.
- `written` là số byte thực tế đã ghi.
- Nếu lỗi khi ghi, trả về số byte đã ghi trước khi lỗi xảy ra.
- `total` dùng để theo dõi tổng số byte ciphertext đã ghi.

---

### 9.9. Chuẩn bị buffer đọc plaintext

```go
	// Mã hóa từng chunk
	buf := make([]byte, ChunkSize)
	var chunkIdx uint64 = 0
```

- `buf` là vùng nhớ tạm để đọc từng chunk plaintext.
- `chunkIdx` đánh số chunk: 0, 1, 2, ...

---

### 9.10. Vòng lặp đọc và mã hóa từng chunk

```go
	for {
		n, readErr := io.ReadFull(src, buf)
```

- Cố gắng đọc đầy `ChunkSize` byte từ `src` vào `buf`.
- `n` là số byte thực sự đọc được.
- `readErr` cho biết trạng thái đọc.

`io.ReadFull` có hành vi quan trọng:
- Nếu đủ chunk: `n = ChunkSize`, `readErr = nil`
- Nếu chunk cuối ngắn hơn: `n < ChunkSize`, `readErr = io.ErrUnexpectedEOF`
- Nếu không còn gì để đọc ngay từ đầu: `n = 0`, `readErr = io.EOF`

---

### 9.11. Kiểm tra kết thúc dữ liệu

```go
		if n == 0 && readErr == io.EOF {
			break
		}
```

- Không còn dữ liệu nào nữa thì thoát vòng lặp.

---

### 9.12. Xử lý lỗi đọc bất thường

```go
		if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
			return total, fmt.Errorf("crypto: read plaintext: %w", readErr)
		}
```

- Chỉ chấp nhận 2 lỗi “bình thường” ở cuối stream:
  - `io.ErrUnexpectedEOF`
  - `io.EOF`
- Các lỗi khác là lỗi thực sự khi đọc nguồn dữ liệu.

---

### 9.13. Sinh nonce cho chunk hiện tại

```go
		nonce := deriveChunkNonce(baseNonce, chunkIdx)
```

- Tạo nonce riêng cho chunk này.
- Nonce được tạo từ `baseNonce` + `chunkIdx`.

---

### 9.14. Mã hóa chunk bằng GCM

```go
		ciphertext := gcm.Seal(nil, nonce, buf[:n], nil)
```

- `buf[:n]` là plaintext thực sự của chunk.
- `gcm.Seal(...)` trả về ciphertext kèm authentication tag.
- `nil` cuối cùng là `additional authenticated data (AAD)`; ở đây không dùng.

---

### 9.15. Ghi độ dài chunk ciphertext

```go
		// Ghi [4 bytes length][ciphertext]
		lenBuf := make([]byte, 4)
		binary.BigEndian.PutUint32(lenBuf, uint32(len(ciphertext)))
```

- Mỗi chunk được ghi kèm độ dài 4 byte trước nó.
- Vì ciphertext của AES-GCM dài hơn plaintext do có tag, nên cần lưu độ dài để đọc lại chính xác lúc giải mã.

```go
		if _, err := dst.Write(lenBuf); err != nil {
			return total, fmt.Errorf("crypto: write chunk length: %w", err)
		}
		total += 4
```

- Ghi 4 byte length.
- Nếu lỗi thì trả về.
- Cộng 4 vào tổng byte đã ghi.

---

### 9.16. Ghi ciphertext chunk

```go
		w, err := dst.Write(ciphertext)
		total += int64(w)
		if err != nil {
			return total, fmt.Errorf("crypto: write chunk: %w", err)
		}
```

- Ghi ciphertext của chunk ra `dst`.
- Cộng số byte đã ghi vào `total`.
- Nếu ghi lỗi thì trả ra luôn.

---

### 9.17. Tăng chỉ số chunk và quyết định dừng

```go
		chunkIdx++
		if readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
			break
		}
	}
```

- Sau khi xử lý xong chunk hiện tại, tăng `chunkIdx`.
- Nếu đây là chunk cuối cùng thì thoát vòng lặp.

Lưu ý:
- `io.ErrUnexpectedEOF` ở đây rất bình thường cho chunk cuối nếu kích thước file không chia hết cho 64 KiB.

---

### 9.18. Kết thúc hàm Encrypt

```go
	return total, nil
}
```

- Trả tổng số byte đã ghi và không có lỗi.

---

## 10. Hàm `Decrypt`

```go
// Decrypt membaca ciphertext dari src, mendekripsi, menulis plaintext ke dst.
// Password harus cocok dengan yang dipakai saat Encrypt.
func Decrypt(dst io.Writer, src io.Reader, password string) error {
```

### Mục đích
- Đọc ciphertext theo định dạng AEShield
- Phân tích header
- Dẫn xuất lại key từ password
- Giải mã từng chunk và ghi plaintext vào `dst`

---

### 10.1. Đọc header

```go
	// Baca header
	headerBuf := make([]byte, 4+1+saltSize+nonceSize)
	if _, err := io.ReadFull(src, headerBuf); err != nil {
		return ErrShortHeader
	}
```

- Cấp phát buffer 33 byte cho header.
- Dùng `io.ReadFull` để bắt buộc đọc đủ 33 byte.
- Nếu không đủ thì file quá ngắn hoặc hỏng -> trả `ErrShortHeader`.

---

### 10.2. Kiểm tra magic

```go
	if string(headerBuf[:4]) != HeaderMagic {
		return ErrInvalidMagic
	}
```

- Lấy 4 byte đầu và so sánh với `AES\x00`.
- Nếu không khớp thì đây không phải file hợp lệ của hệ thống.

---

### 10.3. Lấy thông tin key size

```go
	bits := KeyBits(int(headerBuf[4]) * 64)
	if _, err := keyLen(bits); err != nil {
		return ErrInvalidKeyBits
	}
```

- Byte thứ 5 lưu `2/3/4`.
- Nhân `64` để ra `128/192/256`.
- Kiểm tra lại xem có hợp lệ không.

---

### 10.4. Cắt salt và baseNonce từ header

```go
	salt := headerBuf[5 : 5+saltSize]
	baseNonce := headerBuf[5+saltSize : 5+saltSize+nonceSize]
```

- `salt` nằm ngay sau byte key size.
- `baseNonce` nằm sau salt.

Cụ thể:
- byte 0..3: magic
- byte 4: key size marker
- byte 5..20: salt (16 byte)
- byte 21..32: nonce gốc (12 byte)

---

### 10.5. Dẫn xuất lại key

```go
	key, err := DeriveKey(password, salt, bits)
	if err != nil {
		return err
	}
```

- Dùng chính password người dùng nhập vào.
- Nếu password giống lúc mã hóa, key sẽ giống.
- Nếu password sai, key sinh ra khác và GCM sẽ không mở được chunk.

---

### 10.6. Tạo AES cipher và GCM giống lúc mã hóa

```go
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("crypto: new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("crypto: new GCM: %w", err)
	}
```

- Hai bước này giống với phía `Encrypt`.
- Muốn giải mã được thì thuật toán, key và nonce phải khớp.

---

### 10.7. Chuẩn bị đọc từng chunk

```go
	// Giải mã từng chunk
	lenBuf := make([]byte, 4)
	var chunkIdx uint64 = 0
```

- `lenBuf` dùng đọc 4 byte độ dài chunk ciphertext.
- `chunkIdx` dùng để tính nonce tương ứng cho từng chunk.

---

### 10.8. Vòng lặp đọc và giải mã từng chunk

```go
	for {
		_, err := io.ReadFull(src, lenBuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
		if err != nil {
			return fmt.Errorf("crypto: read chunk length: %w", err)
		}
```

Giải thích:
- Mỗi chunk bắt đầu bằng 4 byte length.
- Nếu không đọc được vì đã hết file ở ranh giới bình thường, vòng lặp dừng.
- Nếu có lỗi khác thì báo lỗi.

---

### 10.9. Đọc ciphertext của chunk hiện tại

```go
		chunkLen := binary.BigEndian.Uint32(lenBuf)
		ciphertext := make([]byte, chunkLen)
		if _, err := io.ReadFull(src, ciphertext); err != nil {
			return fmt.Errorf("crypto: read chunk data: %w", err)
		}
```

- Chuyển 4 byte length thành số nguyên `chunkLen`.
- Cấp phát buffer đúng bằng độ dài chunk ciphertext.
- Đọc đủ toàn bộ chunk ciphertext.

---

### 10.10. Tái tạo nonce của chunk

```go
		nonce := deriveChunkNonce(baseNonce, chunkIdx)
```

- Phải tạo đúng nonce như phía mã hóa.
- Nếu `chunkIdx` hoặc `baseNonce` lệch thì GCM sẽ mở thất bại.

---

### 10.11. Giải mã và xác thực chunk

```go
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return ErrDecryptChunk
		}
```

- `gcm.Open(...)` vừa giải mã vừa kiểm tra tính toàn vẹn.
- Nếu password sai hoặc ciphertext bị sửa, hàm này sẽ lỗi.
- Code không trả lỗi gốc chi tiết mà chuẩn hóa thành `ErrDecryptChunk`.

---

### 10.12. Ghi plaintext ra output

```go
		if _, err := dst.Write(plaintext); err != nil {
			return fmt.Errorf("crypto: write plaintext: %w", err)
		}
```

- Ghi phần plaintext vừa giải mã ra `dst`.

---

### 10.13. Sang chunk tiếp theo

```go
		chunkIdx++
	}

	return nil
}
```

- Tăng chỉ số chunk.
- Khi hết dữ liệu thì thoát vòng lặp và trả `nil`.

---

## 11. Hàm `deriveChunkNonce`

```go
// deriveChunkNonce XOR nonce gốc với chỉ số chunk để mỗi chunk có nonce độc lập.
// Điều này đảm bảo không bao giờ tái sử dụng (key, nonce) pair.
func deriveChunkNonce(baseNonce []byte, chunkIdx uint64) []byte {
```

### Mục đích
Từ một `baseNonce`, tạo ra nonce riêng cho mỗi chunk.

---

### Giải thích từng dòng

```go
	nonce := make([]byte, nonceSize)
```
- Tạo slice mới dài 12 byte.

```go
	copy(nonce, baseNonce)
```
- Copy `baseNonce` sang để không sửa dữ liệu gốc.

```go
	// XOR 8 byte cuối với chunkIdx
	idxBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idxBytes, chunkIdx)
```

- Chuyển `chunkIdx` thành 8 byte theo thứ tự BigEndian.
- Ví dụ:
  - chunk 0 -> `00 00 00 00 00 00 00 00`
  - chunk 1 -> `00 00 00 00 00 00 00 01`

```go
	for i := 0; i < 8; i++ {
		nonce[nonceSize-8+i] ^= idxBytes[i]
	}
```

- Lấy 8 byte cuối của nonce và XOR với `idxBytes`.
- Kết quả là mỗi chunk có nonce khác nhau.

```go
	return nonce
}
```

- Trả nonce mới.

### Vì sao làm vậy?
- GCM cực kỳ nhạy cảm với việc tái sử dụng nonce cùng một key.
- Cách này tạo ra nonce khác nhau theo chunk index mà vẫn có thể tái tạo lại trong lúc giải mã.

---

## 12. Hàm `keyLen`

```go
func keyLen(bits KeyBits) (int, error) {
	switch bits {
	case KeyBits128:
		return 16, nil
	case KeyBits192:
		return 24, nil
	case KeyBits256:
		return 32, nil
	default:
		return 0, ErrInvalidKeyBits
	}
}
```

### Mục đích
Đổi số bit của khóa AES thành số byte thực tế.

### Mapping
- 128 bit = 16 byte
- 192 bit = 24 byte
- 256 bit = 32 byte

### Vì sao cần hàm này?
- AES API trong Go cần key theo byte, không phải theo bit.
- Đồng thời hàm này cũng là nơi validate giá trị `bits`.

---

## 13. Luồng hoạt động hoàn chỉnh

### Khi mã hóa
1. Nhận `password` và `bits`
2. Sinh `salt` ngẫu nhiên
3. Sinh `baseNonce` ngẫu nhiên
4. Dùng Argon2id để tạo key từ `password + salt`
5. Tạo AES-GCM
6. Ghi header vào file output
7. Đọc plaintext từng chunk 64 KiB
8. Với mỗi chunk:
   - sinh nonce riêng từ `baseNonce + chunkIdx`
   - mã hóa bằng GCM
   - ghi `[4 byte length][ciphertext]`

### Khi giải mã
1. Đọc header
2. Kiểm tra magic
3. Lấy `bits`, `salt`, `baseNonce`
4. Dùng password nhập vào để tạo lại key
5. Tạo AES-GCM
6. Đọc từng chunk ciphertext
7. Với mỗi chunk:
   - tính lại nonce theo `chunkIdx`
   - gọi `gcm.Open(...)`
   - nếu hợp lệ thì ghi plaintext ra output

---

## 14. Những điểm quan trọng cần hiểu

### 14.1. Vì sao dùng Argon2id?
Vì password người dùng thường yếu hơn random key. Argon2id giúp:
- biến password thành key đúng chuẩn AES
- tăng chi phí brute-force
- đảm bảo cùng password nhưng khác salt vẫn ra key khác

### 14.2. Vì sao dùng AES-GCM?
Vì GCM không chỉ mã hóa mà còn xác thực dữ liệu.
Nếu dữ liệu bị sửa, `gcm.Open(...)` sẽ fail.

### 14.3. Vì sao chia chunk?
Để xử lý file lớn mà không phải giữ toàn bộ trong RAM.

### 14.4. Vì sao mỗi chunk phải có nonce riêng?
Vì với GCM, dùng lại cùng `(key, nonce)` là rất nguy hiểm.
Hàm `deriveChunkNonce(...)` được tạo ra để tránh điều đó.

### 14.5. Vì sao phải lưu key size trong header?
Vì lúc giải mã hệ thống cần biết phải dẫn xuất key dài 16, 24 hay 32 byte.

---

## 15. Các vị trí đáng chú ý trong file gốc

- Hằng số chunk/header: [backend/internal/crypto/crypto.go:24-33](backend/internal/crypto/crypto.go#L24-L33)
- Tham số Argon2id: [backend/internal/crypto/crypto.go:35-40](backend/internal/crypto/crypto.go#L35-L40)
- Các lỗi chuẩn hóa: [backend/internal/crypto/crypto.go:42-47](backend/internal/crypto/crypto.go#L42-L47)
- Kiểu `KeyBits`: [backend/internal/crypto/crypto.go:49-56](backend/internal/crypto/crypto.go#L49-L56)
- Hàm `DeriveKey`: [backend/internal/crypto/crypto.go:58-66](backend/internal/crypto/crypto.go#L58-L66)
- Hàm `Encrypt`: [backend/internal/crypto/crypto.go:68-151](backend/internal/crypto/crypto.go#L68-L151)
- Hàm `Decrypt`: [backend/internal/crypto/crypto.go:153-221](backend/internal/crypto/crypto.go#L153-L221)
- Hàm `deriveChunkNonce`: [backend/internal/crypto/crypto.go:223-235](backend/internal/crypto/crypto.go#L223-L235)
- Hàm `keyLen`: [backend/internal/crypto/crypto.go:237-248](backend/internal/crypto/crypto.go#L237-L248)

---

## 16. Tóm tắt ngắn

File này làm đúng 4 việc chính:

1. **Dẫn xuất AES key từ password** bằng Argon2id
2. **Mã hóa dữ liệu theo từng chunk** bằng AES-GCM
3. **Lưu header** để có thể giải mã lại sau này
4. **Giải mã và kiểm tra toàn vẹn dữ liệu** từng chunk

Nếu nhìn ở mức hệ thống, đây là lõi mã hóa AES của backend AEShield.

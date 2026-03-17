import argon2 from "argon2-browser/dist/argon2-bundled.min.js";

const HEADER_MAGIC = [0x41, 0x45, 0x53, 0x00]; // "AES\x00"
const HEADER_SIZE = 4 + 1 + 16 + 12;

const KEY_LENGTH_BY_HEADER_BYTE = {
  2: 16,
  3: 24,
  4: 32,
};

const TEXT_MIME_BY_EXT = {
  txt: "text/plain",
  md: "text/markdown",
  json: "application/json",
  csv: "text/csv",
  log: "text/plain",
};

const IMAGE_MIME_BY_EXT = {
  png: "image/png",
  jpg: "image/jpeg",
  jpeg: "image/jpeg",
  gif: "image/gif",
  webp: "image/webp",
  svg: "image/svg+xml",
};

export function parseEncryptedHeader(encryptedBytes) {
  if (!(encryptedBytes instanceof Uint8Array)) {
    throw new Error("Dữ liệu mã hóa không hợp lệ.");
  }

  if (encryptedBytes.length < HEADER_SIZE) {
    throw new Error("Tệp mã hóa không hợp lệ (header quá ngắn).");
  }

  for (let i = 0; i < HEADER_MAGIC.length; i += 1) {
    if (encryptedBytes[i] !== HEADER_MAGIC[i]) {
      throw new Error("Tệp không đúng định dạng mã hóa AEShield.");
    }
  }

  const keyByte = encryptedBytes[4];
  const keyLength = KEY_LENGTH_BY_HEADER_BYTE[keyByte];
  if (!keyLength) {
    throw new Error("Tệp mã hóa có kích thước khóa không được hỗ trợ.");
  }

  const salt = encryptedBytes.slice(5, 21);
  const baseNonce = encryptedBytes.slice(21, 33);

  return {
    keyByte,
    keyLength,
    salt,
    baseNonce,
    dataOffset: HEADER_SIZE,
  };
}

export async function deriveKeyArgon2id(password, salt, keyLength) {
  if (!password || !password.trim()) {
    throw new Error("Vui lòng nhập mật khẩu để giải mã.");
  }

  if (!window?.crypto?.subtle) {
    throw new Error("Trình duyệt không hỗ trợ Web Crypto API.");
  }

  const result = await argon2.hash({
    pass: password,
    salt,
    time: 1,
    mem: 64 * 1024,
    parallelism: 4,
    hashLen: keyLength,
    type: argon2.ArgonType.Argon2id,
  });

  const rawKey =
    result.hash instanceof Uint8Array ? result.hash : new Uint8Array(result.hash);

  return window.crypto.subtle.importKey(
    "raw",
    rawKey,
    { name: "AES-GCM" },
    false,
    ["decrypt"]
  );
}

export function deriveChunkNonce(baseNonce, chunkIndex) {
  if (!(baseNonce instanceof Uint8Array) || baseNonce.length !== 12) {
    throw new Error("Nonce gốc không hợp lệ.");
  }

  const nonce = new Uint8Array(baseNonce);
  const idx = BigInt(chunkIndex);

  for (let i = 0; i < 8; i += 1) {
    const shift = BigInt((7 - i) * 8);
    const idxByte = Number((idx >> shift) & 0xffn);
    nonce[4 + i] ^= idxByte;
  }

  return nonce;
}

export async function decryptEncryptedFile(encryptedInput, password) {
  const encryptedBytes =
    encryptedInput instanceof Uint8Array
      ? encryptedInput
      : new Uint8Array(encryptedInput);

  const { keyLength, salt, baseNonce, dataOffset } =
    parseEncryptedHeader(encryptedBytes);

  const cryptoKey = await deriveKeyArgon2id(password, salt, keyLength);

  const plaintextChunks = [];
  let totalPlaintextLength = 0;
  let offset = dataOffset;
  let chunkIndex = 0;

  while (offset < encryptedBytes.length) {
    if (offset + 4 > encryptedBytes.length) {
      throw new Error("Tệp mã hóa bị lỗi (thiếu độ dài chunk).");
    }

    const chunkLength =
      (encryptedBytes[offset] * 0x1000000) +
      (encryptedBytes[offset + 1] << 16) +
      (encryptedBytes[offset + 2] << 8) +
      encryptedBytes[offset + 3];
    offset += 4;

    if (!Number.isFinite(chunkLength) || chunkLength < 16) {
      throw new Error("Tệp mã hóa bị lỗi (chunk không hợp lệ).");
    }

    if (offset + chunkLength > encryptedBytes.length) {
      throw new Error("Tệp mã hóa bị lỗi (chunk vượt quá kích thước tệp).");
    }

    const cipherChunk = encryptedBytes.slice(offset, offset + chunkLength);
    offset += chunkLength;

    const nonce = deriveChunkNonce(baseNonce, chunkIndex);

    try {
      const plainChunkBuffer = await window.crypto.subtle.decrypt(
        {
          name: "AES-GCM",
          iv: nonce,
        },
        cryptoKey,
        cipherChunk
      );

      const plainChunk = new Uint8Array(plainChunkBuffer);
      plaintextChunks.push(plainChunk);
      totalPlaintextLength += plainChunk.length;
      chunkIndex += 1;
    } catch {
      throw new Error("Giải mã thất bại. Mật khẩu không đúng hoặc dữ liệu đã bị thay đổi.");
    }
  }

  const plaintext = new Uint8Array(totalPlaintextLength);
  let writeOffset = 0;

  for (const chunk of plaintextChunks) {
    plaintext.set(chunk, writeOffset);
    writeOffset += chunk.length;
  }

  return plaintext;
}

export function detectPreviewType(filename) {
  const safeName = typeof filename === "string" ? filename : "";
  const fileNameParts = safeName.toLowerCase().split(".");
  const ext = fileNameParts.length > 1 ? fileNameParts[fileNameParts.length - 1] : "";

  if (TEXT_MIME_BY_EXT[ext]) {
    return {
      type: "text",
      mimeType: TEXT_MIME_BY_EXT[ext],
      supported: true,
    };
  }

  if (IMAGE_MIME_BY_EXT[ext]) {
    return {
      type: "image",
      mimeType: IMAGE_MIME_BY_EXT[ext],
      supported: true,
    };
  }

  if (ext === "pdf") {
    return {
      type: "pdf",
      mimeType: "application/pdf",
      supported: true,
    };
  }

  return {
    type: "unsupported",
    mimeType: "application/octet-stream",
    supported: false,
  };
}

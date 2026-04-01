import argon2 from "argon2-browser/dist/argon2-bundled.min.js";

const HEADER_MAGIC = [0x41, 0x45, 0x53, 0x00]; // "AES\x00"
const HEADER_SIZE = 4 + 1 + 16 + 12;
const AUTH_TAG_SIZE = 16;

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

const AES_S_BOX = new Uint8Array([
  0x63, 0x7c, 0x77, 0x7b, 0xf2, 0x6b, 0x6f, 0xc5, 0x30, 0x01, 0x67, 0x2b, 0xfe,
  0xd7, 0xab, 0x76, 0xca, 0x82, 0xc9, 0x7d, 0xfa, 0x59, 0x47, 0xf0, 0xad, 0xd4,
  0xa2, 0xaf, 0x9c, 0xa4, 0x72, 0xc0, 0xb7, 0xfd, 0x93, 0x26, 0x36, 0x3f, 0xf7,
  0xcc, 0x34, 0xa5, 0xe5, 0xf1, 0x71, 0xd8, 0x31, 0x15, 0x04, 0xc7, 0x23, 0xc3,
  0x18, 0x96, 0x05, 0x9a, 0x07, 0x12, 0x80, 0xe2, 0xeb, 0x27, 0xb2, 0x75, 0x09,
  0x83, 0x2c, 0x1a, 0x1b, 0x6e, 0x5a, 0xa0, 0x52, 0x3b, 0xd6, 0xb3, 0x29, 0xe3,
  0x2f, 0x84, 0x53, 0xd1, 0x00, 0xed, 0x20, 0xfc, 0xb1, 0x5b, 0x6a, 0xcb, 0xbe,
  0x39, 0x4a, 0x4c, 0x58, 0xcf, 0xd0, 0xef, 0xaa, 0xfb, 0x43, 0x4d, 0x33, 0x85,
  0x45, 0xf9, 0x02, 0x7f, 0x50, 0x3c, 0x9f, 0xa8, 0x51, 0xa3, 0x40, 0x8f, 0x92,
  0x9d, 0x38, 0xf5, 0xbc, 0xb6, 0xda, 0x21, 0x10, 0xff, 0xf3, 0xd2, 0xcd, 0x0c,
  0x13, 0xec, 0x5f, 0x97, 0x44, 0x17, 0xc4, 0xa7, 0x7e, 0x3d, 0x64, 0x5d, 0x19,
  0x73, 0x60, 0x81, 0x4f, 0xdc, 0x22, 0x2a, 0x90, 0x88, 0x46, 0xee, 0xb8, 0x14,
  0xde, 0x5e, 0x0b, 0xdb, 0xe0, 0x32, 0x3a, 0x0a, 0x49, 0x06, 0x24, 0x5c, 0xc2,
  0xd3, 0xac, 0x62, 0x91, 0x95, 0xe4, 0x79, 0xe7, 0xc8, 0x37, 0x6d, 0x8d, 0xd5,
  0x4e, 0xa9, 0x6c, 0x56, 0xf4, 0xea, 0x65, 0x7a, 0xae, 0x08, 0xba, 0x78, 0x25,
  0x2e, 0x1c, 0xa6, 0xb4, 0xc6, 0xe8, 0xdd, 0x74, 0x1f, 0x4b, 0xbd, 0x8b, 0x8a,
  0x70, 0x3e, 0xb5, 0x66, 0x48, 0x03, 0xf6, 0x0e, 0x61, 0x35, 0x57, 0xb9, 0x86,
  0xc1, 0x1d, 0x9e, 0xe1, 0xf8, 0x98, 0x11, 0x69, 0xd9, 0x8e, 0x94, 0x9b, 0x1e,
  0x87, 0xe9, 0xce, 0x55, 0x28, 0xdf, 0x8c, 0xa1, 0x89, 0x0d, 0xbf, 0xe6, 0x42,
  0x68, 0x41, 0x99, 0x2d, 0x0f, 0xb0, 0x54, 0xbb, 0x16,
]);

const AES_RCON = new Uint8Array([
  0x00, 0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x1b, 0x36, 0x6c, 0xd8,
  0xab, 0x4d,
]);

function readChunkLength(encryptedBytes, offset) {
  return (
    (encryptedBytes[offset] * 0x1000000) +
    (encryptedBytes[offset + 1] << 16) +
    (encryptedBytes[offset + 2] << 8) +
    encryptedBytes[offset + 3]
  );
}

function bytesToHex(bytes, bytesPerLine = 16) {
  if (!(bytes instanceof Uint8Array) || bytes.length === 0) {
    return "";
  }

  const lines = [];

  for (let i = 0; i < bytes.length; i += bytesPerLine) {
    const lineBytes = bytes.slice(i, i + bytesPerLine);
    const line = Array.from(lineBytes, (value) =>
      value.toString(16).padStart(2, "0")
    ).join(" ");
    lines.push(line);
  }

  return lines.join("\n");
}

function bytesToBase64(bytes) {
  if (!(bytes instanceof Uint8Array) || bytes.length === 0) {
    return "";
  }

  let binary = "";
  for (let i = 0; i < bytes.length; i += 1) {
    binary += String.fromCharCode(bytes[i]);
  }

  return btoa(binary);
}

function gmul2(x) {
  if ((x & 0x80) !== 0) {
    return ((x << 1) ^ 0x1b) & 0xff;
  }
  return (x << 1) & 0xff;
}

function gmul3(x) {
  return gmul2(x) ^ x;
}

function subBytes(state) {
  for (let i = 0; i < state.length; i += 1) {
    state[i] = AES_S_BOX[state[i]];
  }
}

function shiftRows(state) {
  [state[1], state[5], state[9], state[13]] = [
    state[5],
    state[9],
    state[13],
    state[1],
  ];
  [state[2], state[6], state[10], state[14]] = [
    state[10],
    state[14],
    state[2],
    state[6],
  ];
  [state[3], state[7], state[11], state[15]] = [
    state[15],
    state[3],
    state[7],
    state[11],
  ];
}

function mixColumns(state) {
  for (let c = 0; c < 4; c += 1) {
    const off = c * 4;
    const s0 = state[off];
    const s1 = state[off + 1];
    const s2 = state[off + 2];
    const s3 = state[off + 3];
    state[off] = gmul2(s0) ^ gmul3(s1) ^ s2 ^ s3;
    state[off + 1] = s0 ^ gmul2(s1) ^ gmul3(s2) ^ s3;
    state[off + 2] = s0 ^ s1 ^ gmul2(s2) ^ gmul3(s3);
    state[off + 3] = gmul3(s0) ^ s1 ^ s2 ^ gmul2(s3);
  }
}

function addRoundKey(state, roundKey, offset = 0) {
  for (let i = 0; i < 16; i += 1) {
    state[i] ^= roundKey[offset + i];
  }
}

function rotWord(word) {
  return [word[1], word[2], word[3], word[0]];
}

function subWord(word) {
  return [
    AES_S_BOX[word[0]],
    AES_S_BOX[word[1]],
    AES_S_BOX[word[2]],
    AES_S_BOX[word[3]],
  ];
}

function expandAESKey(key) {
  const nk = key.length / 4;
  let rounds;
  switch (nk) {
    case 4:
      rounds = 10;
      break;
    case 6:
      rounds = 12;
      break;
    case 8:
      rounds = 14;
      break;
    default:
      throw new Error(`AES key length không hợp lệ: ${key.length}`);
  }

  const totalWords = 4 * (rounds + 1);
  const words = new Uint8Array(totalWords * 4);
  words.set(key);

  for (let i = nk; i < totalWords; i += 1) {
    let temp = [
      words[(i - 1) * 4],
      words[(i - 1) * 4 + 1],
      words[(i - 1) * 4 + 2],
      words[(i - 1) * 4 + 3],
    ];

    if (i % nk === 0) {
      temp = subWord(rotWord(temp));
      temp[0] ^= AES_RCON[Math.floor(i / nk)];
    } else if (nk > 6 && i % nk === 4) {
      temp = subWord(temp);
    }

    for (let j = 0; j < 4; j += 1) {
      words[i * 4 + j] = words[(i - nk) * 4 + j] ^ temp[j];
    }
  }

  return { roundKeys: words, rounds };
}

function createAESBlock(rawKey) {
  if (![16, 24, 32].includes(rawKey.length)) {
    throw new Error("Kích thước khóa AES không được hỗ trợ.");
  }

  return expandAESKey(rawKey);
}

function encryptAESBlock(block, src) {
  if (!(src instanceof Uint8Array) || src.length < 16) {
    throw new Error("AES block cần đúng 16 bytes dữ liệu đầu vào.");
  }

  const state = new Uint8Array(16);
  state.set(src.slice(0, 16));

  addRoundKey(state, block.roundKeys, 0);
  for (let round = 1; round < block.rounds; round += 1) {
    subBytes(state);
    shiftRows(state);
    mixColumns(state);
    addRoundKey(state, block.roundKeys, round * 16);
  }
  subBytes(state);
  shiftRows(state);
  addRoundKey(state, block.roundKeys, block.rounds * 16);

  return state;
}

function xorInto16(dst, src) {
  for (let i = 0; i < 16; i += 1) {
    dst[i] ^= src[i];
  }
}

function bitAt(block, bitIndex) {
  const byteIndex = Math.floor(bitIndex / 8);
  const shift = 7 - (bitIndex % 8);
  return (block[byteIndex] >> shift) & 1;
}

function shiftRightOne(block) {
  let carry = 0;
  for (let i = 0; i < 16; i += 1) {
    const nextCarry = block[i] & 1;
    block[i] = ((block[i] >> 1) | (carry << 7)) & 0xff;
    carry = nextCarry;
  }
}

function gfMul(x, y) {
  const z = new Uint8Array(16);
  const v = new Uint8Array(y);

  for (let i = 0; i < 128; i += 1) {
    if (bitAt(x, i) === 1) {
      xorInto16(z, v);
    }
    const lsb = v[15] & 1;
    shiftRightOne(v);
    if (lsb === 1) {
      v[0] ^= 0xe1;
    }
  }

  return z;
}

function processGHASHBlocks(y, h, data) {
  let offset = 0;
  while (offset + 16 <= data.length) {
    xorInto16(y, data.slice(offset, offset + 16));
    y.set(gfMul(y, h));
    offset += 16;
  }

  if (offset >= data.length) {
    return;
  }

  const block = new Uint8Array(16);
  block.set(data.slice(offset));
  xorInto16(y, block);
  y.set(gfMul(y, h));
}

function writeUint64BE(target, offset, value) {
  const big = BigInt(value);
  for (let i = 0; i < 8; i += 1) {
    const shift = BigInt((7 - i) * 8);
    target[offset + i] = Number((big >> shift) & 0xffn);
  }
}

function ghash(h, aad, ciphertext) {
  const y = new Uint8Array(16);
  processGHASHBlocks(y, h, aad);
  processGHASHBlocks(y, h, ciphertext);

  const lengthBlock = new Uint8Array(16);
  writeUint64BE(lengthBlock, 0, aad.length * 8);
  writeUint64BE(lengthBlock, 8, ciphertext.length * 8);
  xorInto16(y, lengthBlock);
  return gfMul(y, h);
}

function inc32(counter) {
  let value =
    ((counter[12] << 24) | (counter[13] << 16) | (counter[14] << 8) | counter[15]) >>> 0;
  value = (value + 1) >>> 0;
  counter[12] = (value >>> 24) & 0xff;
  counter[13] = (value >>> 16) & 0xff;
  counter[14] = (value >>> 8) & 0xff;
  counter[15] = value & 0xff;
}

function ctrXOR(block, src, counterBytes) {
  const streamCounter = new Uint8Array(counterBytes);
  const dst = new Uint8Array(src.length);

  for (let offset = 0; offset < src.length; offset += 16) {
    const keystream = encryptAESBlock(block, streamCounter);
    const chunkLen = Math.min(16, src.length - offset);
    for (let i = 0; i < chunkLen; i += 1) {
      dst[offset + i] = src[offset + i] ^ keystream[i];
    }
    inc32(streamCounter);
  }

  return dst;
}

function deriveJ0(h, block, nonce) {
  if (nonce.length === 12) {
    const j0 = new Uint8Array(16);
    j0.set(nonce, 0);
    j0[15] = 1;
    return j0;
  }

  return ghash(h, new Uint8Array(0), nonce);
}

function constantTimeEqual(a, b) {
  if (a.length !== b.length) {
    return false;
  }

  let diff = 0;
  for (let i = 0; i < a.length; i += 1) {
    diff |= a[i] ^ b[i];
  }
  return diff === 0;
}

function decryptChunkWithFallback(rawKey, nonce, ciphertext) {
  if (ciphertext.length < AUTH_TAG_SIZE) {
    throw new Error("Giải mã thất bại. Dữ liệu chunk không hợp lệ.");
  }

  const block = createAESBlock(rawKey);
  const h = encryptAESBlock(block, new Uint8Array(16));
  const data = ciphertext.slice(0, ciphertext.length - AUTH_TAG_SIZE);
  const tag = ciphertext.slice(ciphertext.length - AUTH_TAG_SIZE);

  const j0 = deriveJ0(h, block, nonce);
  const s = ghash(h, new Uint8Array(0), data);
  const tagMask = encryptAESBlock(block, j0);
  for (let i = 0; i < AUTH_TAG_SIZE; i += 1) {
    s[i] ^= tagMask[i];
  }

  if (!constantTimeEqual(tag, s)) {
    throw new Error("Giải mã thất bại. Mật khẩu không đúng hoặc dữ liệu đã bị thay đổi.");
  }

  const counter = new Uint8Array(j0);
  inc32(counter);
  return ctrXOR(block, data, counter);
}

async function decryptChunkWithWebCrypto(rawKey, nonce, ciphertext) {
  if (!window?.crypto?.subtle) {
    throw new Error("Trình duyệt không hỗ trợ Web Crypto API.");
  }

  const cryptoKey = await window.crypto.subtle.importKey(
    "raw",
    rawKey,
    { name: "AES-GCM" },
    false,
    ["decrypt"]
  );

  const plainChunkBuffer = await window.crypto.subtle.decrypt(
    {
      name: "AES-GCM",
      iv: nonce,
    },
    cryptoKey,
    ciphertext
  );

  return new Uint8Array(plainChunkBuffer);
}

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

  const result = await argon2.hash({
    pass: password,
    salt,
    time: 1,
    mem: 64 * 1024,
    parallelism: 4,
    hashLen: keyLength,
    type: argon2.ArgonType.Argon2id,
  });

  return result.hash instanceof Uint8Array
    ? result.hash
    : new Uint8Array(result.hash);
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

  const rawKey = await deriveKeyArgon2id(password, salt, keyLength);

  const plaintextChunks = [];
  let totalPlaintextLength = 0;
  let offset = dataOffset;
  let chunkIndex = 0;

  while (offset < encryptedBytes.length) {
    if (offset + 4 > encryptedBytes.length) {
      throw new Error("Tệp mã hóa bị lỗi (thiếu độ dài chunk).");
    }

    const chunkLength = readChunkLength(encryptedBytes, offset);
    offset += 4;

    if (!Number.isFinite(chunkLength) || chunkLength < AUTH_TAG_SIZE) {
      throw new Error("Tệp mã hóa bị lỗi (chunk không hợp lệ).");
    }

    if (offset + chunkLength > encryptedBytes.length) {
      throw new Error("Tệp mã hóa bị lỗi (chunk vượt quá kích thước tệp).");
    }

    const cipherChunk = encryptedBytes.slice(offset, offset + chunkLength);
    offset += chunkLength;

    const nonce = deriveChunkNonce(baseNonce, chunkIndex);

    try {
      let plainChunk;
      if (keyLength === 24) {
        plainChunk = decryptChunkWithFallback(rawKey, nonce, cipherChunk);
      } else {
        try {
          plainChunk = await decryptChunkWithWebCrypto(rawKey, nonce, cipherChunk);
        } catch {
          plainChunk = decryptChunkWithFallback(rawKey, nonce, cipherChunk);
        }
      }

      plaintextChunks.push(plainChunk);
      totalPlaintextLength += plainChunk.length;
      chunkIndex += 1;
    } catch (error) {
      if (error instanceof Error && error.message) {
        throw error;
      }
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

export function inspectEncryptedFile(encryptedInput, options = {}) {
  const encryptedBytes =
    encryptedInput instanceof Uint8Array
      ? encryptedInput
      : new Uint8Array(encryptedInput);

  const { keyByte, keyLength, salt, baseNonce, dataOffset } =
    parseEncryptedHeader(encryptedBytes);

  const maxPreviewBytes = Number.isFinite(options.maxPreviewBytes)
    ? Math.max(1, Math.floor(options.maxPreviewBytes))
    : 512;

  let offset = dataOffset;
  let chunkCount = 0;
  let totalCipherPayloadBytes = 0;

  while (offset < encryptedBytes.length) {
    if (offset + 4 > encryptedBytes.length) {
      throw new Error("Tệp mã hóa bị lỗi (thiếu độ dài chunk).");
    }

    const chunkLength = readChunkLength(encryptedBytes, offset);
    offset += 4;

    if (!Number.isFinite(chunkLength) || chunkLength < AUTH_TAG_SIZE) {
      throw new Error("Tệp mã hóa bị lỗi (chunk không hợp lệ).");
    }

    if (offset + chunkLength > encryptedBytes.length) {
      throw new Error("Tệp mã hóa bị lỗi (chunk vượt quá kích thước tệp).");
    }

    totalCipherPayloadBytes += chunkLength;
    chunkCount += 1;
    offset += chunkLength;
  }

  const previewBytes = encryptedBytes.slice(
    0,
    Math.min(maxPreviewBytes, encryptedBytes.length)
  );

  return {
    totalEncryptedBytes: encryptedBytes.length,
    headerSize: dataOffset,
    keyByte,
    keyLength,
    saltHex: bytesToHex(salt),
    baseNonceHex: bytesToHex(baseNonce),
    chunkCount,
    totalCipherPayloadBytes,
    previewHex: bytesToHex(previewBytes),
    previewBase64: bytesToBase64(previewBytes),
    truncated:
      previewBytes.length < encryptedBytes.length
        ? `Đang hiển thị ${previewBytes.length}/${encryptedBytes.length} bytes đầu.`
        : "Đang hiển thị toàn bộ dữ liệu mã hóa.",
  };
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

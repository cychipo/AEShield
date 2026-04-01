package crypto

import (
	"crypto/subtle"
	"errors"
)

var errAuthFailed = errors.New("crypto: message authentication failed")

type gcmCipher struct {
	block *aesBlock
	h     [16]byte
}

func newGCM(block *aesBlock) (*gcmCipher, error) {
	if block == nil {
		return nil, errors.New("crypto: nil block")
	}
	var zero [16]byte
	var h [16]byte
	block.Encrypt(h[:], zero[:])
	return &gcmCipher{block: block, h: h}, nil
}

func (g *gcmCipher) Seal(dst, nonce, plaintext, aad []byte) []byte {
	j0 := deriveJ0(g, nonce)
	counter := j0
	inc32(counter[:])

	ciphertext := make([]byte, len(plaintext))
	g.ctrXOR(ciphertext, plaintext, counter[:])

	s := ghash(g.h, aad, ciphertext)
	var tagMask [16]byte
	g.block.Encrypt(tagMask[:], j0[:])
	for i := 0; i < 16; i++ {
		s[i] ^= tagMask[i]
	}

	out := make([]byte, len(dst)+len(ciphertext)+16)
	copy(out, dst)
	copy(out[len(dst):], ciphertext)
	copy(out[len(dst)+len(ciphertext):], s[:])
	return out
}

func (g *gcmCipher) Open(dst, nonce, ciphertext, aad []byte) ([]byte, error) {
	if len(ciphertext) < 16 {
		return nil, errAuthFailed
	}

	dataLen := len(ciphertext) - 16
	data := ciphertext[:dataLen]
	tag := ciphertext[dataLen:]

	j0 := deriveJ0(g, nonce)
	s := ghash(g.h, aad, data)
	var tagMask [16]byte
	g.block.Encrypt(tagMask[:], j0[:])
	for i := 0; i < 16; i++ {
		s[i] ^= tagMask[i]
	}
	if subtle.ConstantTimeCompare(tag, s[:]) != 1 {
		return nil, errAuthFailed
	}

	counter := j0
	inc32(counter[:])
	plaintext := make([]byte, dataLen)
	g.ctrXOR(plaintext, data, counter[:])

	out := make([]byte, len(dst)+len(plaintext))
	copy(out, dst)
	copy(out[len(dst):], plaintext)
	return out, nil
}

func (g *gcmCipher) ctrXOR(dst, src, counter []byte) {
	streamCounter := make([]byte, 16)
	copy(streamCounter, counter)

	var keystream [16]byte
	for offset := 0; offset < len(src); offset += 16 {
		g.block.Encrypt(keystream[:], streamCounter)
		chunkLen := len(src) - offset
		if chunkLen > 16 {
			chunkLen = 16
		}
		for i := 0; i < chunkLen; i++ {
			dst[offset+i] = src[offset+i] ^ keystream[i]
		}
		inc32(streamCounter)
	}
}

func deriveJ0(g *gcmCipher, nonce []byte) [16]byte {
	var j0 [16]byte
	if len(nonce) == 12 {
		copy(j0[:12], nonce)
		j0[15] = 1
		return j0
	}
	return ghash(g.h, nil, nonce)
}

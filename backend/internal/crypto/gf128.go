package crypto

import "encoding/binary"

func ghash(h [16]byte, aad, ciphertext []byte) [16]byte {
	var y [16]byte
	processGHASHBlocks(&y, h, aad)
	processGHASHBlocks(&y, h, ciphertext)

	var lengthBlock [16]byte
	binary.BigEndian.PutUint64(lengthBlock[:8], uint64(len(aad))*8)
	binary.BigEndian.PutUint64(lengthBlock[8:], uint64(len(ciphertext))*8)
	xorBlock(&y, lengthBlock[:])
	y = gfMul(y, h)
	return y
}

func processGHASHBlocks(y *[16]byte, h [16]byte, data []byte) {
	for len(data) >= 16 {
		xorBlock(y, data[:16])
		*y = gfMul(*y, h)
		data = data[16:]
	}
	if len(data) == 0 {
		return
	}

	var block [16]byte
	copy(block[:], data)
	xorBlock(y, block[:])
	*y = gfMul(*y, h)
}

func xorBlock(dst *[16]byte, src []byte) {
	for i := 0; i < 16; i++ {
		dst[i] ^= src[i]
	}
}

func gfMul(x, y [16]byte) [16]byte {
	var z [16]byte
	v := y
	for i := 0; i < 128; i++ {
		if bitAt(x, i) == 1 {
			xorBlock(&z, v[:])
		}
		lsb := v[15] & 1
		shiftRightOne(&v)
		if lsb == 1 {
			v[0] ^= 0xe1
		}
	}
	return z
}

func bitAt(block [16]byte, bitIndex int) byte {
	byteIndex := bitIndex / 8
	shift := 7 - (bitIndex % 8)
	return (block[byteIndex] >> shift) & 1
}

func shiftRightOne(block *[16]byte) {
	var carry byte
	for i := 0; i < 16; i++ {
		nextCarry := block[i] & 1
		block[i] = (block[i] >> 1) | (carry << 7)
		carry = nextCarry
	}
}

func inc32(counter []byte) {
	v := binary.BigEndian.Uint32(counter[12:])
	v++
	binary.BigEndian.PutUint32(counter[12:], v)
}

package murmur

import (
	"encoding/binary"
	"math/bits"
)

const (
	c1 uint32 = 0xcc9e2d51
	c2 uint32 = 0x1b873593
	r1 uint32 = 0xf
	r2 uint32 = 0xd
	m  uint32 = 0x5
	n  uint32 = 0xe6546b64
)

// Hash32 returns a non-cryptographic hash of the provided key
// provided an initial seed.
func Hash32(seed uint32, key []byte) uint32 {
	var (
		hash      = seed
		keyLen    = len(key)
		numChunks = keyLen / 4
		k         uint32
	)

	// Note that if numChunks < 0, the loop isn't executed
	for c := 0; c < numChunks; c++ {
		k = binary.LittleEndian.Uint32(key[c*4:])
		k = k * c1
		k = bits.RotateLeft32(k, int(r1))
		k = k * c2

		hash = hash ^ k
		hash = bits.RotateLeft32(hash, int(r2))
		hash = (hash * m) + n
	}

	// Handle any bytes not handled by the above
	// note that this is either 0, 1, 2, or 3.
	offset := keyLen - (numChunks * 4)
	if offset > 0 {
		rem := to32(key[keyLen-offset:])
		rem = rem * c1
		rem = bits.RotateLeft32(rem, int(r1))
		rem = rem * c2

		hash = hash ^ rem
	}

	hash = hash ^ uint32(keyLen)
	hash = hash ^ (hash >> 16)
	hash = hash * 0x85ebca6b
	hash = hash ^ (hash >> 13)
	hash = hash * 0xc2b2ae35
	hash = hash ^ (hash >> 16)
	return hash
}

// to32 converts a provided byte slice into a uint32.
// no assumptions are made about the size of the slice,
// whereas binary.*.Uint32() assume that the byte slice
// is of size at least 4.
func to32(b []byte) (r uint32) {
	for i, bt := range b {
		r += uint32(bt) << (i * 8)
	}
	return
}

package bloom

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"

	"github.com/butcher-of-blaviken/merkle/murmur"
	"github.com/ethereum/go-ethereum/crypto"
)

type HasherType int

const (
	HasherTypeKeccak256 HasherType = iota
	HasherTypeMurmur32
)

// hasherKeySize is the hasher's key size in bytes.
const hasherKeySize = 4

// hasher represents a keyed hash function.
type hasher func(input []byte) []byte

type hashers []hasher

// hash hashes the given input with the hashers and returns indices into
// the bloom filter for which those bits will be set.
func (h hashers) hash(input []byte, filterSize int) (indices []int) {
	for _, hshr := range h {
		h := new(big.Int).SetBytes(hshr(input))
		index := h.Mod(h, big.NewInt(int64(filterSize)))
		indices = append(indices, int(index.Int64())) // should be a safe downcast
	}
	return
}

func newKeccak256Hasher() hasher {
	b := make([]byte, hasherKeySize)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return func(input []byte) []byte {
		return crypto.Keccak256(append(b, input...))
	}
}

func newMurmur32Hasher() hasher {
	b := make([]byte, hasherKeySize)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return func(input []byte) (r []byte) {
		return binary.LittleEndian.AppendUint32(
			r,
			murmur.Hash32(binary.LittleEndian.Uint32(b), input),
		)
	}
}

func makeKeccak256Hashers(numHashes int) (hs hashers) {
	for i := 0; i < numHashes; i++ {
		hs = append(hs, newKeccak256Hasher())
	}
	return
}

func makeMurmur32Hashers(numHashes int) (hs hashers) {
	for i := 0; i < numHashes; i++ {
		hs = append(hs, newMurmur32Hasher())
	}
	return
}

func makeHashers(numHashes int, tp HasherType) (hs hashers) {
	switch tp {
	case HasherTypeKeccak256:
		return makeKeccak256Hashers(numHashes)
	case HasherTypeMurmur32:
		return makeMurmur32Hashers(numHashes)
	default:
		panic("unknown hasher type")
	}
}

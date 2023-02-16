package bloom

import (
	"crypto/rand"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// FalsePositiveRate is the recommended false positive rate
	// of the bloom filter of 1%.
	FalsePositiveRate float64 = 0.01
)

type Filter interface {
	// Insert inserts the given item into the bloom filter.
	// The item must be hashable.
	Insert(item []byte) error
	// Contains checks whether the provided item is in the bloom filter.
	// The item must be hashable.
	Contains(item []byte) bool
}

// hasherKeySize is the hasher's key size in bytes.
const hasherKeySize = 8

// hasher represents a keyed hash function.
type hasher struct {
	// the actual hash function, e.g kecca256 or sha256
	hash func(input []byte) []byte
	// a random byte prefix applied to each input
	key []byte
}

type hashers []*hasher

// hash hashes the given input with the hashers and returns indices into
// the bloom filter for which those bits will be set.
func (h hashers) hash(input []byte, filterSize int) (indices []int) {
	for _, hshr := range h {
		h := new(big.Int).SetBytes(hshr.hash(input))
		index := h.Mod(h, big.NewInt(int64(filterSize)))
		indices = append(indices, int(index.Int64())) // should be a safe downcast
	}
	return
}

func newHasher() *hasher {
	b := make([]byte, hasherKeySize)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return &hasher{
		hash: func(input []byte) []byte {
			return crypto.Keccak256(append(b, input...))
		},
		key: b,
	}
}

func bloomParams(expectedSize int, falsePositiveRate float64) (filterSize int, numHashes int) {
	m := math.Ceil(
		(float64(-expectedSize) * math.Log2(falsePositiveRate)) / math.Pow(math.Log2(2), 2.0),
	)
	numHashes = int((m / float64(expectedSize)) * math.Log2(2))
	return int(m), numHashes
}

func makeHashers(numHashes int) (hs hashers) {
	for i := 0; i < numHashes; i++ {
		hs = append(hs, newHasher())
	}
	return
}

// New creates a new bloom filter, given the expected size of the set
// backed by the filter, and a false positive rate interpreted as a
// percentage out of 100 (i.e 1 = 1%).
func New(expectedSize int, falsePositiveRate float64) Filter {
	filterSizeBits, numHashes := bloomParams(expectedSize, falsePositiveRate)
	return &filter{
		hs:     makeHashers(numHashes),
		filter: make([]byte, filterSizeBits/8),
	}
}

type filter struct {
	hs     hashers
	filter []byte
}

func (f *filter) Insert(item []byte) error {
	indices := f.hs.hash(item, len(f.filter)*8)
	setBits(f.filter, indices)
	return nil
}

func (f *filter) Contains(item []byte) bool {
	indices := f.hs.hash(item, len(f.filter)*8)
	for _, idx := range indices {
		idxByte := idx / 8
		if f.filter[idxByte] == 0 {
			return false
		}
	}
	return true
}

func setBits(bitfield []byte, indices []int) {
	for _, idx := range indices {
		idxByte := idx / 8
		// e.g 0000 0000 0000 0000 <= byte field
		// bit index: 5
		// 5 // 8 = 0 => 0th byte
		// 8 // 8 = 1 => 1st byte
		// etc.
		// how to get relative index in the byte?
		// byte index * 8 == starting index of the byte
		// e.g 0 * 8 == 0, starting index of 0th byte
		// e.g 1 * 8 == 8, starting index of 1st byte
		// so bit index - starting byte index == relative index in the byte
		relativeIndex := idx - (idxByte * 8)
		toOr := byte(1 << relativeIndex)
		bitfield[idxByte] |= toOr
	}
}

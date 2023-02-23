package bloom

import (
	"math"
)

const (
	// RecommendedFalsePositiveRate is the recommended false positive rate
	// to use with bloom filters - 1%.
	RecommendedFalsePositiveRate float64 = 0.01
)

type Filter interface {
	// Insert inserts the given item into the bloom filter.
	// The item must be hashable.
	Insert(item []byte) error
	// Contains checks whether the provided item is in the bloom filter.
	// The item must be hashable.
	Contains(item []byte) bool
}

func bloomParams(expectedSize int, falsePositiveRate float64) (filterSize int, numHashes int) {
	m := math.Ceil(
		(float64(-expectedSize) * math.Log2(falsePositiveRate)) / math.Pow(math.Log2(2), 2.0),
	)
	numHashes = int((m / float64(expectedSize)) * math.Log2(2))
	return int(m), numHashes
}

// New creates a new bloom filter, given the expected size of the set
// backed by the filter, and a false positive rate interpreted as a
// percentage out of 100 (i.e 1 = 1%).
func New(expectedSize int, falsePositiveRate float64, tp HasherType) Filter {
	filterSizeBits, numHashes := bloomParams(expectedSize, falsePositiveRate)
	return &filter{
		hs:     makeHashers(numHashes, tp),
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

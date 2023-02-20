package sparse

import "errors"

// Array represents a sparse array that is able to
// store the presence of an element (rather than the element
// itself). This is most useful when an element represents
// a letter of some alphabet (i.e the English alphabet,
// which has 26 letters). Since Array is backed by a uint64,
// it can store up to 64 possible characters, meaning
// the base64 alphabet can be represented.
type Array interface {
	Set(i int)
	Unset(i int)
	Get(i int) bool
}

type array struct {
	// 0 index represents "right-most" bit.
	bitfield uint64
}

var (
	ErrOutOfRange       = errors.New("index out of range")
	_             Array = &array{}
)

// New creates a new Array with no elements.
func New() Array {
	return &array{}
}

// Set sets the index i in the sparse array to be present.
func (a *array) Set(i int) {
	a.bitfield |= 1 << i
}

// Unset sets the index i in the sparse array to be not present.
func (a *array) Unset(i int) {
	a.bitfield &= ^(1 << i)
}

// Get returns whether the bit at index i in the sparse array is set or not.
func (a *array) Get(i int) bool {
	return (a.bitfield & (1 << i)) != 0
}

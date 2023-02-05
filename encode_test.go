package merkle

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMPT_CompactEncode(t *testing.T) {
	path := []byte{1, 2, 3, 4, 5}
	encoded := compactEncode(path)
	assert.Equal(t, "010102030405", hex.EncodeToString(encoded))
	path = []byte{0, 1, 2, 3, 4, 5}
	encoded = compactEncode(path)
	assert.Equal(t, "0000000102030405", hex.EncodeToString(encoded))
}

func TestMPT_ExtractCommonPrefix(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5}
	b := []byte{1, 2, 3}
	c := extractCommonPrefix(a, b)
	assert.Equal(t, []byte{1, 2, 3}, c)
	assert.Equal(t, []byte(nil), extractCommonPrefix([]byte{1, 2, 3}, []byte{4, 5, 6, 7}))
	assert.Equal(t, []byte(nil), extractCommonPrefix([]byte{1, 2, 3}, []byte{2, 1, 2, 3}))
}

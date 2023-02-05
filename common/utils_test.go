package common_test

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/butcher-of-blaviken/merkle/common"
)

func TestCompactEncode(t *testing.T) {
	path := []byte{1, 2, 3, 4, 5}
	encoded := common.CompactEncode(path)
	assert.Equal(t, "010102030405", hex.EncodeToString(encoded))
	path = []byte{0, 1, 2, 3, 4, 5}
	encoded = common.CompactEncode(path)
	assert.Equal(t, "0000000102030405", hex.EncodeToString(encoded))
}

func TestExtractCommonPrefix(t *testing.T) {
	a := []byte{1, 2, 3, 4, 5}
	b := []byte{1, 2, 3}
	c := common.ExtractCommonPrefix(a, b)
	assert.Equal(t, []byte{1, 2, 3}, c)
	assert.Equal(t, []byte(nil), common.ExtractCommonPrefix([]byte{1, 2, 3}, []byte{4, 5, 6, 7}))
	assert.Equal(t, []byte(nil), common.ExtractCommonPrefix([]byte{1, 2, 3}, []byte{2, 1, 2, 3}))
}

func TestConcat(t *testing.T) {
	assert.Equal(t, []byte{1, 2, 3, 4, 5, 6}, common.Concat(
		[]byte{1, 2, 3}, []byte{4, 5, 6},
	))
	assert.Equal(t, []byte{1, 2, 3, 4}, common.Concat(
		[]byte{1, 2, 3}, []byte{4},
	))
	assert.Equal(t, []byte{1, 2, 3}, common.Concat(
		[]byte{1, 2, 3}, []byte{},
	))
}

func TestBytesToNibbles(t *testing.T) {
	assert.Equal(
		t,
		[]byte{6, 6, 6, 9, 7, 2, 7, 3, 7, 4, 7, 0, 6, 1, 7, 4, 6, 8},
		common.BytesToNibbles(([]byte("firstpath"))))
	assert.Equal(
		t,
		[]byte{7, 3, 6, 5, 6, 3, 6, 15, 6, 14, 6, 4, 7, 0, 6, 1, 7, 4, 6, 8},
		common.BytesToNibbles([]byte("secondpath")),
	)
	assert.Equal(
		t,
		[]byte{7, 4, 6, 8, 6, 9, 7, 2, 6, 4, 7, 0, 6, 1, 7, 4, 6, 8},
		common.BytesToNibbles([]byte("thirdpath")),
	)
}

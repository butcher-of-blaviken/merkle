package bloom_test

import (
	"testing"

	"github.com/butcher-of-blaviken/merkle/bloom"
	"github.com/stretchr/testify/require"
)

func TestBloomInsert(t *testing.T) {
	b := bloom.New(1000, 0.1)

	require.NoError(t, b.Insert([]byte("hello")))
	require.NoError(t, b.Insert([]byte("world")))
	require.True(t, b.Contains([]byte("hello")))
	require.True(t, b.Contains([]byte("world")))
}

func FuzzBloom(f *testing.F) {
	testCases := [][]byte{
		[]byte("hello"),
		[]byte("world"),
		[]byte("standard issue"),
		[]byte("complications"),
	}
	for _, tc := range testCases {
		f.Add(tc)
	}
	b := bloom.New(1000, 0.01)
	f.Fuzz(func(t *testing.T, item []byte) {
		require.NoError(t, b.Insert(item))
		require.True(t, b.Contains(item))
	})
}

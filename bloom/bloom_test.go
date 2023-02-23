package bloom_test

import (
	"math/rand"
	"testing"

	"github.com/butcher-of-blaviken/merkle/bloom"
	"github.com/stretchr/testify/require"
)

var (
	letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

func TestBloomInsert(t *testing.T) {
	b := bloom.New(1000, 0.1, bloom.HasherTypeKeccak256)

	require.NoError(t, b.Insert([]byte("hello")))
	require.NoError(t, b.Insert([]byte("world")))
	require.True(t, b.Contains([]byte("hello")))
	require.True(t, b.Contains([]byte("world")))
}

func TestBloomInsertMurmur32(t *testing.T) {
	b := bloom.New(1000, 0.1, bloom.HasherTypeKeccak256)

	require.NoError(t, b.Insert([]byte("hello")))
	require.NoError(t, b.Insert([]byte("world")))
	require.True(t, b.Contains([]byte("hello")))
	require.True(t, b.Contains([]byte("world")))
}

func TestBloomInsertContains(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	var testCases [][]byte
	for i := 0; i < 100; i++ {
		testCases = append(testCases, randomString(r, 10))
	}
	b := bloom.New(100, 0.001, bloom.HasherTypeMurmur32)
	for _, tc := range testCases {
		require.NoError(t, b.Insert(tc))
	}
	for _, tc := range testCases {
		require.True(t, b.Contains(tc))
	}
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
	b := bloom.New(1000, 0.01, bloom.HasherTypeKeccak256)
	f.Fuzz(func(t *testing.T, item []byte) {
		require.NoError(t, b.Insert(item))
		require.True(t, b.Contains(item))
	})
}

func randomString(r *rand.Rand, n int) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return []byte(string(b))
}

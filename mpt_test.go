package merkle

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMPT_PutGet(t *testing.T) {
	t.Run("empty trie", func(t *testing.T) {
		trie := NewMerklePatriciaTrie()
		_, err := trie.Get([]byte("not-there"))
		assert.Error(t, err)
	})

	t.Run("simple put/get", func(t *testing.T) {
		trie := NewMerklePatriciaTrie()
		assert.NoError(t, trie.Put([]byte("hello"), []byte("world")))
		val, err := trie.Get([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, []byte("world"), val)
	})

	t.Run("multiple nodes in tree - leaf to extension to branch", func(t *testing.T) {
		trie := NewMerklePatriciaTrie()
		assert.NoError(t, trie.Put([]byte("hello"), []byte("world")))
		assert.NoError(t, trie.Put([]byte("hello-poop"), []byte("world-also")))
		val, err := trie.Get([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, []byte("world"), val)
		val, err = trie.Get([]byte("hello-poop"))
		assert.NoError(t, err)
		assert.Equal(t, []byte("world-also"), val)
		_, err = trie.Get([]byte("hell"))
		assert.Error(t, err)
	})

	t.Run("multiple nodes in tree, different paths", func(t *testing.T) {
		trie := NewMerklePatriciaTrie()
		assert.NoError(t, trie.Put([]byte("firstpath"), []byte("first")))
		assert.NoError(t, trie.Put([]byte("secondpath"), []byte("second")))
		assert.NoError(t, trie.Put([]byte("thirdpath"), []byte("third")))
		for _, s := range []struct {
			key, expected string
		}{{"firstpath", "first"}, {"secondpath", "second"}, {"thirdpath", "third"}} {
			v, err := trie.Get([]byte(s.key))
			assert.NoError(t, err)
			assert.Equal(t, []byte(s.expected), v, fmt.Sprintf("key: %s, expected: %s", s.key, s.expected))
		}
	})

	t.Run("more nodes, more paths, some clashing", func(t *testing.T) {
		trie := NewMerklePatriciaTrie()
		testCases := []struct {
			key, expected string
		}{
			{"firstpath", "first"},
			{"secondpath", "second"},
			{"thirdpath", "third"},
			{"fourthpath", "fourth"},
			{"fifthpath", "fifth"},
			{"sixthpath", "sixth"},
			{"seventhpath", "seventh"},
			{"eighthpath", "eighth"},
			{"ninthpath", "ninth"},
		}
		for i, tc := range testCases {
			t.Run(fmt.Sprintf("testCase %d Put", i+1), func(t *testing.T) {
				assert.NoError(t, trie.Put([]byte(tc.key), []byte(tc.expected)))
			})
		}
		for i, tc := range testCases {
			t.Run(fmt.Sprintf("testCase %d Get", i+1), func(t *testing.T) {
				v, err := trie.Get([]byte(tc.key))
				assert.NoError(t, err)
				assert.Equal(t, []byte(tc.expected), v, fmt.Sprintf("key: %s, expected: %s", tc.key, tc.expected))
			})
		}
	})
}

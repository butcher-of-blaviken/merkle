package patricia

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	gethTrie "github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/assert"
)

func TestMPT_PutGet(t *testing.T) {
	t.Run("empty trie", func(t *testing.T) {
		trie := New()
		_, err := trie.Get([]byte("not-there"))
		assert.Error(t, err)
	})

	t.Run("simple put/get", func(t *testing.T) {
		trie := New()
		assert.NoError(t, trie.Put([]byte("hello"), []byte("world")))
		val, err := trie.Get([]byte("hello"))
		assert.NoError(t, err)
		assert.Equal(t, []byte("world"), val)
	})

	t.Run("multiple nodes in tree - leaf to extension to branch", func(t *testing.T) {
		trie := New()
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
		trie := New()
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
		trie := New()
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

func TestMPT_Root(t *testing.T) {
	t.Run("empty trie", func(t *testing.T) {
		trie := New()
		actual := trie.Root()
		assert.Equal(t, "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421", hexutil.Encode(actual))
	})

	t.Run("some elements", func(t *testing.T) {
		trie := New()
		trie.Put([]byte{1, 2, 3, 4}, []byte("hello"))
		actual := trie.Root()
		assert.Equal(t, "0x6764f7ad0efcbc11b84fe7567773aa4b12bd6b4d35c05bbc3951b58dedb6c8e8", hexutil.Encode(actual))

		trie.Put([]byte{1, 2}, []byte("world"))
		actual = trie.Root()
		assert.Equal(t, "0xd0efbf92d7ff7c9cc38807248d85407e1b68d3e934d879ca4aa02308ca4bd824", hexutil.Encode(actual))

		trie.Put([]byte{1, 2}, []byte("trie"))
		actual = trie.Root()
		assert.Equal(t, "0x50dc8dca4b79c361cbef2678fa230de5e40e7d00201af9e71881cf2fbdb82487", hexutil.Encode(actual))
	})

	t.Run("geth cross test", func(t *testing.T) {
		trie := New()
		kvs := []struct {
			key, value []byte
		}{
			{[]byte{1, 2, 3, 4}, []byte("hello")},
			{[]byte{1, 2, 5, 4}, []byte("world")},
			{[]byte{1, 2, 6, 4}, []byte("haha")},
			{[]byte{1, 7, 3, 4}, []byte("yessir")},
			{[]byte{9, 2, 3, 4}, []byte("tweet it")},
		}
		gTrie := gethTrie.NewEmpty(gethTrie.NewDatabase(rawdb.NewMemoryDatabase()))
		for _, kv := range kvs {
			assert.NoError(t, trie.Put(kv.key, kv.value))
			gTrie.Update(kv.key, kv.value)

			myRoot := trie.Root()
			gethRoot := gTrie.Hash()
			assert.Equal(t, gethRoot.Hex(), hexutil.Encode(myRoot))
		}
	})
}

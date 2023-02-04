package merkle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixTrie_PutGet(t *testing.T) {
	trie := NewPrefixTrie()
	assert.NoError(t, trie.Put([]byte("hello"), []byte("world")))
	assert.NoError(t, trie.Put([]byte("hell"), []byte("heaven")))

	value, err := trie.Get([]byte("he"))
	assert.Error(t, err)
	assert.Nil(t, value)

	value, err = trie.Get([]byte("hell"))
	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, []byte("heaven"), value)

	value, err = trie.Get([]byte("hello"))
	assert.NoError(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, []byte("world"), value)
}

func TestPrefixTrie_Delete(t *testing.T) {
	trie := NewPrefixTrie()
	assert.NoError(t, trie.Put([]byte("hello"), []byte("world")))
	assert.NoError(t, trie.Put([]byte("hell"), []byte("heaven")))

	trie.Delete([]byte("hell"))
	value, err := trie.Get([]byte("hell"))
	assert.Error(t, err)
	assert.Nil(t, value)
}

package merkle

import "fmt"

var (
	ErrKeyNotFound = fmt.Errorf("key not found")
)

var _ Trie = &prefixTrie{}

// Trie defines the interface for a trie data structure.
// A trie is a search tree that implements a key-value store.
type Trie interface {
	// Get returns the value set for the provided key, if the key is in the trie.
	Get(key []byte) (value []byte, err error)
	// Put inserts a key-value pair into the trie.
	Put(key, value []byte) error
	// Delete removes the values associated with the provided key from the trie.
	Delete(key []byte)
}

type prefixTrieNode struct {
	children [16]*prefixTrieNode // hexadecimal alphabet, so 16 possible children (0-f inclusive)
	terminal bool                // whether we're a terminal node or not
	value    []byte              // only set if terminal, nil otherwise
}

type prefixTrie struct {
	root *prefixTrieNode
}

// NewPrefixTrie returns a trie that is implemented using a basic prefix trie.
// The alphabet is the hexadecimal one, from 0-f.
func NewPrefixTrie() Trie {
	return &prefixTrie{
		root: &prefixTrieNode{},
	}
}

// Delete implements Trie
func (p *prefixTrie) Delete(key []byte) {
	nibbleKey := bytesToNibbles(key)
	p.deleteHelper(p.root, nibbleKey)
}

func (p *prefixTrie) deleteHelper(x *prefixTrieNode, key []byte) *prefixTrieNode {
	if x == nil {
		return x
	}
	if key == nil {
		if x.terminal {
			x.terminal = false
			x.value = nil
		}
		for _, c := range x.children {
			if c != nil {
				return c
			}
		}
		return nil
	}
	if len(key) > 1 {
		x.children[key[0]] = p.deleteHelper(x.children[key[0]], key[1:])
	} else {
		x.children[key[0]] = p.deleteHelper(x.children[key[0]], nil)
	}

	return x
}

// Get implements Trie
func (p *prefixTrie) Get(key []byte) (value []byte, err error) {
	root := p.root
	nibbleKey := bytesToNibbles(key)
	for _, b := range nibbleKey {
		if root.children[b] == nil {
			return nil, ErrKeyNotFound
		}
		root = root.children[b]
	}
	if root.value == nil {
		return nil, ErrKeyNotFound
	}
	return root.value, nil
}

// Put implements Trie
func (p *prefixTrie) Put(key []byte, value []byte) error {
	root := p.root
	nibbleKey := bytesToNibbles(key)
	for _, b := range nibbleKey {
		if root.children[b] == nil {
			root.children[b] = &prefixTrieNode{}
		}
		root = root.children[b]
	}
	root.value = value
	root.terminal = true
	return nil
}

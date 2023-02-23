package common

import (
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethdb"
)

var (
	ErrKeyNotFound = fmt.Errorf("key not found")
)

// Trie defines the interface for a trie data structure.
// A trie is a search tree that implements a key-value store.
type Trie interface {
	// Get returns the value set for the provided key, if the key is in the trie.
	Get(key []byte) (value []byte, err error)
	// Put inserts a key-value pair into the trie.
	Put(key, value []byte) error
	// Delete removes the values associated with the provided key from the trie.
	Delete(key []byte) error
}

// MPT stands for "Merkle-Patricia Trie", which is a fusion of
// the merkle tree and patricia trie data structures.
//
// The intent is to provide a key-value store that provides
// merkle proofs of membership.
type MPT interface {
	Trie
	types.TrieHasher
	// Root returns the merkle root (i.e hash) of the entire MPT.
	Root() []byte
	ProofFor(key []byte) (proofDB ethdb.KeyValueReader)
}

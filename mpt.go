package merkle

import (
	"bytes"
	"encoding/hex"
)

var (
	// emptyNodeHash is the known keccak256 root hash of an empty trie.
	emptyNodeHash, _ = hex.DecodeString("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

// MPT stands for "Merkle-Patricia Trie", which is a fusion of
// the merkle tree and patricia trie data structures.
//
// The intent is to provide a key-value store that provides
// merkle proofs of membership.
type MPT interface {
	Trie
	// Root returns the merkle root (i.e hash) of the entire MPT.
	Root() []byte
}

type mpt struct {
	root mptNode
}

// Delete implements MPT
func (m *mpt) Delete(key []byte) {
	panic("unimplemented")
}

// Get implements MPT
func (m *mpt) Get(key []byte) (value []byte, err error) {
	node := m.root
	nibbles := bytesToNibbles(key)
	for {
		if node == nil {
			return nil, ErrKeyNotFound
		}
		switch node.kind() {
		case nodeKindBranch:
			// check if we have a path for the first nibble
			// and recursively continue
			branch := node.(*branchNode)
			if len(nibbles) > 0 {
				// the case where node is set to nil is handled above,
				// no need to handle it here again.
				node = branch.children[nibbles[0]] // jump one level down
				nibbles = nibbles[1:]              // nibble off first nibble
				continue
			}

			if branch.value != nil {
				return branch.value, nil
			}
			return nil, ErrKeyNotFound
		case nodeKindExtension:
			// extract the common prefix from the nibbles that
			// remain and the extension path.
			extension := node.(*extensionNode)
			commonPrefix := extractCommonPrefix(extension.path, nibbles)
			if len(commonPrefix) < len(extension.path) {
				return nil, ErrKeyNotFound
			}
			// "skip" through all the common nibbles and jump to the next node.
			// this is where the optimization kicks in.
			nibbles = nibbles[len(commonPrefix):]
			node = extension.next
		case nodeKindLeaf:
			leaf := node.(*leafNode)
			// if the remaining nibbles match the path in the leaf then we've found
			// the value.
			if bytes.Equal(nibbles, leaf.path) {
				return leaf.value, nil
			}
			// otherwise, we're at a leaf (i.e no more child nodes) and we haven't
			// found the provided path.
			return nil, ErrKeyNotFound
		default:
			panic("unexpected node kind - bug?")
		}
	}
}

// Put implements MPT
func (m *mpt) Put(key []byte, value []byte) error {
	node := &m.root
	nibbles := bytesToNibbles(key)
	for {
		// case: NULL node
		if *node == nil {
			*node = &leafNode{
				path:  nibbles,
				value: value,
			}
			return nil
		}

		switch (*node).kind() {
		case nodeKindBranch:
			branch := (*node).(*branchNode)
			if len(nibbles) > 0 {
				node = &branch.children[nibbles[0]]
				nibbles = nibbles[1:]
				continue
			} else {
				// store the value in the branch
				branch.value = value
				return nil
			}
		case nodeKindExtension:
			extension := (*node).(*extensionNode)
			commonPrefix := extractCommonPrefix(extension.path, nibbles)
			// only two cases we care about here:
			// 1. common prefix length is less than the extension path length.
			//   a. in this case we reduce the path size of this extension node and add
			//      a new branch and a new leaf node.
			// 2. common prefix length is greater than or equal to extension path length.
			//   a. in this case we can trim off the matching nibbles and continue down
			//      the trie.
			if len(commonPrefix) < len(extension.path) {
				// case 1.
				newExtPath := extension.path[:len(commonPrefix)]
				branchNibble := extension.path[len(commonPrefix)]
				remainingPath := extension.path[len(commonPrefix)+1:]
				branch := &branchNode{}
				if len(remainingPath) == 0 {
					branch.children[branchNibble] = extension.next
				} else {
					branch.children[branchNibble] = &extensionNode{
						path: remainingPath,
						next: extension.next,
					}
				}

				if len(commonPrefix) < len(nibbles) {
					branchNibble, remaining := nibbles[len(commonPrefix)], nibbles[len(commonPrefix)+1:]
					branch.children[branchNibble] = &leafNode{
						path:  remaining,
						value: value,
					}
				} else if len(commonPrefix) == len(nibbles) {
					branch.value = value
				} else {
					panic("invariant violated: len(commonPrefix) > len(nibbles)") // should be impossible
				}

				if len(newExtPath) == 0 {
					*node = branch
				} else {
					*node = &extensionNode{
						path: newExtPath,
						next: branch,
					}
				}
				return nil // insert done, no nibbles left
			}

			// case 2.
			nibbles = nibbles[len(commonPrefix):]
			node = &extension.next
			continue
		case nodeKindLeaf:
			leaf := (*node).(*leafNode)
			commonPrefix := extractCommonPrefix(leaf.path, nibbles)

			// if the common prefix matches both the remaining nibbles and
			// the leaf path, then we can update the leaf value in-place.
			if len(commonPrefix) == len(nibbles) && len(commonPrefix) == len(leaf.path) {
				leaf.value = value
				return nil
			}

			branch := &branchNode{}
			// only one of the cases below will be true, since the third possibility is
			// checked above.
			if len(commonPrefix) == len(leaf.path) {
				branch.value = leaf.value
			}

			if len(commonPrefix) == len(nibbles) {
				branch.value = value
			}

			if len(commonPrefix) > 0 {
				// create an extension node that will store the common prefix
				// between the leaf and the remaining nibbles
				extension := &extensionNode{
					path: commonPrefix,
					next: branch,
				}
				*node = extension
			} else {
				// when there is no common prefix, we'll be replacing the leaf node
				// with a branch node.
				*node = branch
			}

			if len(commonPrefix) < len(leaf.path) {
				branch.children[leaf.path[len(commonPrefix)]] = &leafNode{
					path:  leaf.path[len(commonPrefix)+1:],
					value: leaf.value,
				}
			}

			if len(commonPrefix) < len(nibbles) {
				branch.children[nibbles[len(commonPrefix)]] = &leafNode{
					path:  nibbles[len(commonPrefix)+1:],
					value: value,
				}
			}

			return nil
		default:
			panic("unexpected node kind - bug?")
		}
	}
}

func (m *mpt) Root() []byte {
	panic("unimplemented")
}

func NewMerklePatriciaTrie() MPT {
	return &mpt{
		root: nil,
	}
}

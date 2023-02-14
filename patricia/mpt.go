package patricia

import (
	"bytes"

	"github.com/butcher-of-blaviken/merkle/common"
	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

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
	nibbles := common.BytesToNibbles(key)
	for {
		if node == nil {
			return nil, common.ErrKeyNotFound
		}
		switch n := node.(type) {
		case *branchNode:
			// check if we have a path for the first nibble
			// and recursively continue
			if len(nibbles) > 0 {
				// the case where node is set to nil is handled above,
				// no need to handle it here again.
				node = n.children[nibbles[0]] // jump one level down
				nibbles = nibbles[1:]         // nibble off first nibble
				continue
			}

			if n.value != nil {
				return n.value, nil
			}
			return nil, common.ErrKeyNotFound
		case *extensionNode:
			// extract the common prefix from the nibbles that
			// remain and the extension path.
			commonPrefix := common.ExtractCommonPrefix(n.path, nibbles)
			if len(commonPrefix) < len(n.path) {
				return nil, common.ErrKeyNotFound
			}
			// "skip" through all the common nibbles and jump to the next node.
			// this is where the optimization kicks in.
			nibbles = nibbles[len(commonPrefix):]
			node = n.next
		case *leafNode:
			// if the remaining nibbles match the path in the leaf then we've found
			// the value.
			if bytes.Equal(nibbles, n.path) {
				return n.value, nil
			}
			// otherwise, we're at a leaf (i.e no more child nodes) and we haven't
			// found the provided path.
			return nil, common.ErrKeyNotFound
		default:
			panic("unexpected node kind - bug?")
		}
	}
}

// Put implements MPT
func (m *mpt) Put(key []byte, value []byte) error {
	node := &m.root
	nibbles := common.BytesToNibbles(key)
	for {
		// case: NULL node
		if *node == nil {
			*node = &leafNode{
				path:  nibbles,
				value: value,
			}
			return nil
		}

		switch n := (*node).(type) {
		case *branchNode:
			if len(nibbles) > 0 {
				node = &n.children[nibbles[0]]
				nibbles = nibbles[1:]
				continue
			} else {
				// store the value in the branch
				n.value = value
				return nil
			}
		case *extensionNode:
			var (
				commonPrefix    = common.ExtractCommonPrefix(n.path, nibbles)
				commonPrefixLen = len(commonPrefix)
			)

			// only two cases we care about here:
			// 1. common prefix length is less than the extension path length.
			//   a. in this case we reduce the path size of this extension node and add
			//      a new branch and a new leaf node.
			// 2. common prefix length is greater than or equal to extension path length.
			//   a. in this case we can trim off the matching nibbles and continue down
			//      the trie.
			if commonPrefixLen < len(n.path) {
				// case 1.
				newExtPath := n.path[:commonPrefixLen]
				branchNibble := n.path[commonPrefixLen]
				remainingPath := n.path[commonPrefixLen+1:]
				branch := &branchNode{}
				if len(remainingPath) == 0 {
					branch.children[branchNibble] = n.next
				} else {
					branch.children[branchNibble] = &extensionNode{
						path: remainingPath,
						next: n.next,
					}
				}

				if commonPrefixLen < len(nibbles) {
					branchNibble, remaining := nibbles[commonPrefixLen], nibbles[commonPrefixLen+1:]
					branch.children[branchNibble] = &leafNode{
						path:  remaining,
						value: value,
					}
				} else if commonPrefixLen == len(nibbles) {
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
			nibbles = nibbles[commonPrefixLen:]
			node = &n.next
			continue
		case *leafNode:
			var (
				commonPrefix    = common.ExtractCommonPrefix(n.path, nibbles)
				commonPrefixLen = len(commonPrefix)
			)

			// if the common prefix matches both the remaining nibbles and
			// the leaf path, then we can update the leaf value in-place.
			if commonPrefixLen == len(nibbles) && commonPrefixLen == len(n.path) {
				n.value = value
				return nil
			}

			branch := &branchNode{}
			// only one of the cases below will be true, since the third possibility is
			// checked above.
			if commonPrefixLen == len(n.path) {
				branch.value = n.value
			}

			if commonPrefixLen == len(nibbles) {
				branch.value = value
			}

			if commonPrefixLen > 0 {
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

			if commonPrefixLen < len(n.path) {
				branch.children[n.path[commonPrefixLen]] = &leafNode{
					path:  n.path[commonPrefixLen+1:],
					value: n.value,
				}
			}

			if commonPrefixLen < len(nibbles) {
				branch.children[nibbles[commonPrefixLen]] = &leafNode{
					path:  nibbles[commonPrefixLen+1:],
					value: value,
				}
			}

			return nil
		default:
			panic("unexpected node kind - bug?")
		}
	}
}

// Root returns the merkle root of this MPT
func (m *mpt) Root() []byte {
	if m.root == nil {
		return crypto.Keccak256(rlp.EmptyString)
	}
	return hash(m.root)
}

// Reset implements types.TrieHasher
func (m *mpt) Reset() {
	m.root = nil
}

// Update implements types.TrieHasher
func (m *mpt) Update(key, value []byte) {
	m.Put(key, value)
}

// Hash implements types.TrieHasher
func (m *mpt) Hash() gethCommon.Hash {
	return gethCommon.BytesToHash(m.Root())
}

// New returns an empty Merkle-Patricia trie ready for use.
func New() common.MPT {
	return &mpt{
		root: nil,
	}
}

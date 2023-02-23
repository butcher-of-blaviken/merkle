package patricia

import (
	"bytes"

	"github.com/butcher-of-blaviken/merkle/common"
	gethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

type mpt struct {
	root mptNode
}

// Delete implements MPT
// Delete deletes the value associated with the provided key from the trie.
// Note that Del _does not_ return an error if the key is not in the trie.
func (m *mpt) Delete(key []byte) error {
	_, newRoot, err := m.delete(m.root, nil, common.BytesToNibbles(key))
	if err != nil {
		return err
	}

	m.root = newRoot

	return nil
}

// delete deletes the provided key from the trie rooted at n.
// it returns whether a delete occurred (called "dirty"), the new root node of the
// subtree, and any error that occurred.
// delete is recursive, unlike Put and Get, since we need to fix up the
// tree structure after removing a key on the way _up_ the tree rather than
// on the way _down_.
func (m *mpt) delete(n mptNode, prefix, key []byte) (dirty bool, newRoot mptNode, err error) {
	switch n := n.(type) {
	case nil:
		return false, nil, nil
	case *branchNode:
		// Case 1. n.children[key[0]] == nil, in which case the key is not present in the trie.
		// Case 2. n.children[key[0]] != nil, in which case we recursively delete.
		// The returned root is the _new_ root of the subtree previously rooted at
		// n.children[key[0]].
		dirty, newRoot, err = m.delete(n.children[key[0]], append(prefix, key[0]), key[1:])
		if !dirty || err != nil {
			return false, n, err
		}

		// update the subtree reference.
		n.children[key[0]] = newRoot

		// Because n is a branch node, it must've contained at least two children
		// before the delete operation. Otherwise, it would just be an extension node.
		// Case 1. newRoot != nil, in which case n still has at least 2 children,
		// and can remain a branch node.
		// Case 2. newRoot == nil, in which case n has one less child, and we should
		// check if we can reduce it to an extension node.
		if newRoot != nil {
			return true, n, nil
		}

		nonNilIndex := nonNilOnlyChildIndex(n.children[:])
		if nonNilIndex >= 0 {
			if nonNilIndex < 16 {
				cnode := n.children[nonNilIndex]
				switch cn := cnode.(type) {
				case *extensionNode:
					// If the only child of this branch is an extension node,
					// merge them together to form a single extension node.
					k := append([]byte{byte(nonNilIndex)}, cn.path...)
					return true, &extensionNode{
						path: k,
						next: cn.next,
					}, nil
				case *leafNode:
					// If the only child of this branch is a leaf node,
					// merge them together to form a single leaf node.
					return true, &leafNode{
						path:  append([]byte{byte(nonNilIndex)}, cn.path...),
						value: cn.value,
					}, nil
				}
			}

			return true, &extensionNode{[]byte{byte(nonNilIndex)}, n.children[nonNilIndex]}, nil
		}
		// n still contains at least two values and cannot be reduced.
		return true, n, nil
	case *extensionNode:
		prefixLength := len(common.ExtractCommonPrefix(key, n.path))
		// Case 1. len(n.path) > prefixLength.
		// Case 2. len(key) == prefixLength
		if len(n.path) > prefixLength {
			return false, n, nil
		}

		// The key is longer than n.path. Remove the remaining suffix
		// from the subtrie. Child can never be nil here since the
		// subtrie must contain at least two other values with keys
		// longer than n.path.
		dirty, child, err := m.delete(n.next, append(prefix, key[:len(n.path)]...), key[len(n.path):])
		if !dirty || err != nil {
			return false, n, err
		}
		switch child := child.(type) {
		case *extensionNode:
			// merge two extension nodes into one by stitching their paths
			// together.
			return true, &extensionNode{common.Concat(n.path, child.path), child.next}, nil
		case *leafNode:
			panic("insertion variant inviolated")
		default:
			// possible cases are nil and branch node.
			// in both cases we can just point to the subtree without having
			// to merge it.
			// Note that it's impossible that child is a leaf node, since this
			// violates the insertion invariant.
			return true, &extensionNode{n.path, child}, nil
		}
	case *leafNode:
		return true, nil, nil
	default:
		panic("unknown node type") // impossible
	}
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

// ProofFor constructs a merkle proof for the provided key.
// The result contains all encoded nodes on the path to the
// value at key. The value itself is also included in the last node.
func (m *mpt) ProofFor(key []byte) ethdb.KeyValueReader {
	var (
		proofDB = rawdb.NewMemoryDatabase()
		nibbles = common.BytesToNibbles(key)
		node    = m.root
	)
	for {
		proofDB.Put(hash(node), serialize(node))

		if node == nil {
			return proofDB
		}

		switch n := node.(type) {
		case *leafNode:
			// if the remaining nibbles match the path in the leaf then we've found
			// the value.
			if bytes.Equal(nibbles, n.path) {
				return proofDB
			}
			// key not found
			return nil
		case *extensionNode:
			// extract the common prefix from the nibbles that
			// remain and the extension path.
			commonPrefix := common.ExtractCommonPrefix(n.path, nibbles)
			if len(commonPrefix) < len(n.path) {
				// key not found
				return nil
			}
			// "skip" through all the common nibbles and jump to the next node.
			// this is where the optimization kicks in.
			nibbles = nibbles[len(commonPrefix):]
			node = n.next
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
				return proofDB
			}
			return nil // key not found
		default:
			panic("unexpected node kind - bug?")
		}
	}
}

// New returns an empty Merkle-Patricia trie ready for use.
func New() common.MPT {
	return &mpt{
		root: nil,
	}
}

// nonNilOnlyChildIndex returns the index of the only non-nil
// child in the given slice, or -1 if more than one non-nil child
// exists.
// in the event that no non-nil children exist, -2 is returned.
func nonNilOnlyChildIndex(children []mptNode) (pos int) {
	var nonNilChildren int
	for i, child := range children {
		if child != nil {
			nonNilChildren++
			pos = i
		}
	}
	// if there is more than one non-nil child we can't
	// collapse the branch node any further.
	if nonNilChildren > 1 {
		return -1
	}
	if nonNilChildren == 0 {
		return -2
	}
	return pos
}

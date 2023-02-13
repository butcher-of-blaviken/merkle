package patricia

import (
	"fmt"

	"github.com/butcher-of-blaviken/merkle/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
)

// mptNode is an interface that is implemented by all MPT node types.
type mptNode interface {
	preRLP() []any
}

// MPT have four kinds of nodes.
// See: https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/#optimization
// 1. NULL: (represented as the empty string)
// 2. branch: A 17-item node [ v0 ... v15, vt ]
// 3. leaf: A 2-item node [ encodedPath, value ]
// 4. extension: A 2-item node [ encodedPath, key ]
// In ethereum, paths are 64 characters long (64 nibbles in a 32 byte hash).

var (
	_ mptNode = &branchNode{}
	_ mptNode = &leafNode{}
	_ mptNode = &extensionNode{}
)

// leafNode is a node in an mpt that has no children. They contain
// what remains of the path (from the root) and an rlp-encoded value
// which could mean e.g the account state (in ethereum).
type leafNode struct {
	path  []byte // all nibbles
	value []byte
}

// preRLP implements mptNode
func (l *leafNode) preRLP() []any {
	ce := common.CompactEncode(l.path, true)
	return []any{
		ce,
		l.value,
	}
}

// extensionNode is an optimization in mpt's which allows us to "shortcut"
// a potentially long path into a single node.
// Since 64 character paths in ethereum are unlikely to have many collisions,
// this saves on a lot of space (otherwise, your tree will be much deeper).
type extensionNode struct {
	path []byte // all nibbles
	next mptNode
}

// preRLP implements mptNode
func (e *extensionNode) preRLP() []any {
	ce := common.CompactEncode(e.path, false)
	r := []any{
		ce,
	}
	if nextRLP := serialize(e.next); len(nextRLP) >= 32 {
		r = append(r, crypto.Keccak256(nextRLP))
	} else {
		r = append(r, e.next.preRLP())
	}
	return r
}

type branchNode struct {
	children [16]mptNode // 16 nodes + value = 17 items total
	value    []byte
}

// preRLP implements mptNode
func (b *branchNode) preRLP() (r []any) {
	for _, c := range b.children {
		if c == nil {
			r = append(r, []byte{})
		} else {
			if cRLP := serialize(c); len(cRLP) >= 32 {
				r = append(r, crypto.Keccak256(cRLP))
			} else {
				r = append(r, c.preRLP())
			}
		}
	}
	r = append(r, b.value)
	if len(r) != 17 {
		panic(fmt.Sprintf("invariant violated: branch preRLP must be length 17, got %d", len(r)))
	}
	return
}

func hash(node mptNode) []byte {
	return crypto.Keccak256(serialize(node))
}

func serialize(node mptNode) []byte {
	var preRLP any

	if node == nil {
		preRLP = []byte{}
	} else {
		preRLP = node.preRLP()
	}

	rlpEncoded, err := rlp.EncodeToBytes(preRLP)
	if err != nil {
		panic(err) // should never happen
	}

	return rlpEncoded
}

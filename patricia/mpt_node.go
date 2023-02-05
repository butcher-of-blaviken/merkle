package patricia

type nodeKind int

const (
	nodeKindInvalid = iota
	nodeKindBranch
	nodeKindLeaf
	nodeKindExtension
)

// mptNode is an interface that is implemented by all MPT node types.
type mptNode interface {
	kind() nodeKind
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

// kind implements mptNode
func (*leafNode) kind() nodeKind {
	return nodeKindLeaf
}

// extensionNode is an optimization in mpt's which allows us to "shortcut"
// a potentially long path into a single node.
// Since 64 character paths in ethereum are unlikely to have many collisions,
// this saves on a lot of space (otherwise, your tree will be much deeper).
type extensionNode struct {
	path []byte // all nibbles
	next mptNode
}

// kind implements mptNode
func (*extensionNode) kind() nodeKind {
	return nodeKindExtension
}

type branchNode struct {
	children [16]mptNode // 16 nodes + value = 17 items total
	value    []byte
}

// kind implements mptNode
func (*branchNode) kind() nodeKind {
	return nodeKindBranch
}

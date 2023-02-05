package hashtree

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/butcher-of-blaviken/merkle/common"
)

// Bytes32 is a convenience type to represent a 32 byte slice.
type Bytes32 [sha256.Size]byte

// String returns the string representation of the byte slice in hex.
func (b Bytes32) String() string {
	return hex.EncodeToString(b[:])
}

// ProofNode is a merkle tree node with the required data needed to use
// in a merkle proof.
type ProofNode struct {
	Hash  Bytes32
	Index int
}

func (p ProofNode) String() string {
	return fmt.Sprintf("ProofNode(hash=%x, index=%d)", p.Hash[:], p.Index)
}

// Proof represents a merkle proof, which is a sequence of hashes
// and indexes into levels of the merkle tree, that proves that a particular
// piece of data belongs to the tree.
type Proof []ProofNode

// level represents a level in the complete binary tree that
// is the merkle tree.
type level []Bytes32

// Tree represents a merkle tree.
type Tree struct {
	levels []level // starting from bottom, going to the top
}

func (t *Tree) Height() int {
	return len(t.levels)
}

// Root returns the root of the merkle tree.
// This is a single hash.
func (t *Tree) Root() Bytes32 {
	// by construction, the last level should only have one node.
	rootLevel := t.levels[len(t.levels)-1]
	if len(rootLevel) != 1 {
		panic(fmt.Sprintf("merkle invariant violated, expected 1 node in root level, got %d", len(rootLevel)))
	}
	return rootLevel[0]
}

func (t *Tree) String() string {
	// print from the root downwards
	var builder strings.Builder
	for i := len(t.levels) - 1; i >= 0; i-- {
		builder.WriteString(fmt.Sprintf("level %d: ", i+1))
		for _, b := range t.levels[i] {
			builder.WriteString(hex.EncodeToString(b[:]))
			builder.WriteString(", ")
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

// Verify verifies that a provided piece of data is contained
// within the merkle tree.
// True is returned if and only if the leaf is a member of this merkle
// tree.
func (t *Tree) Verify(proof Proof, leaf, root Bytes32) bool {
	hash := leaf
	for _, p := range proof {
		if p.Index%2 == 0 {
			// sibling is a left node, so concat that hash first then the proof node
			hash = sha256.Sum256(common.Concat(p.Hash[:], hash[:]))
		} else {
			// sibling is a right node, so concat proof node first then the sibling
			hash = sha256.Sum256(common.Concat(hash[:], p.Hash[:]))
		}
	}

	return bytes.Equal(hash[:], root[:])
}

// ProofFor returns the merkle proof for the leaf node at index
// i, or an error if that index does not exist.
func (t *Tree) ProofFor(i int) (p Proof, err error) {
	if i > len(t.levels[0])-1 || i < 0 {
		return nil, errors.New("leaf node index out of bounds")
	}

	var (
		siblingIndex = t.getSiblingIndex(i, 0)
		sibling      = t.levels[0][siblingIndex]
	)
	p = append(p, ProofNode{
		Hash:  sibling,
		Index: siblingIndex,
	})

	// now that we have the element and it's sibling, we walk up the levels
	// of the tree to get the nodes that are needed to complete the proof.
	var (
		currIndex = i
		currLevel = 1
	)
	for currLevel < len(t.levels)-1 {
		parentIndex := currIndex / 2
		parentSiblingIndex := t.getSiblingIndex(parentIndex, currLevel)
		parentSibling := t.levels[currLevel][parentSiblingIndex]
		p = append(p, ProofNode{
			Hash:  parentSibling,
			Index: parentSiblingIndex,
		})
		currIndex = parentSiblingIndex
		currLevel++
	}
	return p, nil
}

func (t *Tree) getSiblingIndex(i, level int) int {
	if i%2 == 0 {
		return i + 1
	}
	return i - 1
}

// New constructs a new merkle tree given some data.
func New(data [][]byte) (*Tree, error) {
	if len(data)&(len(data)-1) != 0 {
		return nil, fmt.Errorf("data length must be exact power of two, got: %d", len(data))
	}

	// build the bottom-most level of the tree by hashing the passed in data
	var bottom level
	for i := 0; i < len(data); i += 2 {
		left := sha256.Sum256(data[i])
		right := sha256.Sum256(data[i+1])
		bottom = append(bottom, left)
		bottom = append(bottom, right)
	}

	var (
		allLevels = []level{bottom}
		exponent  = math.Log2(float64(len(data)))
	)
	// build the tree in a bottom up fashion, starting
	// from the deepest level.
	// level i + 1 is constructing by pairwise hashing the nodes
	// on level i.
	for l := 1; l <= int(exponent); l++ {
		var (
			prevLevel = allLevels[l-1]
			currLevel level
		)
		for n := 0; n < len(prevLevel); n += 2 {
			currLevel = append(currLevel, sha256.Sum256(common.Concat(prevLevel[n][:], prevLevel[n+1][:])))
		}
		allLevels = append(allLevels, currLevel)
	}
	return &Tree{
		levels: allLevels,
	}, nil
}

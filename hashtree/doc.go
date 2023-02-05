// package hashtree implements a binary Merkle tree.
//
// Merkle trees provide a simple way to verify the integrity
// of a large amount of data by way of "Merkle proofs" - a sequence
// of carefully selected hashes, one for each level of the tree.
//
// Merkle trees can be n-ary trees, but here we implement binary Merkle
// trees, since they give the smallest sized proofs.
package hashtree

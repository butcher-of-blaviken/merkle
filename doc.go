// package merkle contains a simple and naive implementation
// of the Merkle tree and Merkle Patricia Trie (MPT) data structures.
//
// Merkle trees provide a simple way to verify the integrity
// of a large amount of data by way of "Merkle proofs" - a sequence
// of carefully selected hashes, one for each level of the tree.
//
// Merkle trees can be n-ary trees, but here we implement binary Merkle
// trees, since they give the smallest sized proofs.
//
// Merkle Patricia Tries are a fusion between prefix tries (also known as radix
// tries) and Merkle Trees. They optimize searches and insertions in radix tries
// for long keys that don't have much in common (which is usually the case in
// applications such as Ethereum).
package merkle

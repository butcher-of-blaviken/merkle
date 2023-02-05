// package patricia implements a Merkle-Patricia tree (MPT for short)
//
// Merkle Patricia Tries are a fusion between prefix tries (also known as radix
// tries) and Merkle Trees. They optimize searches and insertions in radix tries
// for long keys that don't have much in common (which is usually the case in
// applications such as Ethereum).
package patricia

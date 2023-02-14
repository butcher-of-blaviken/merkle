// package patricia implements a Merkle-Patricia tree (MPT for short)
//
// Merkle Patricia Tries are a fusion between prefix tries (also known as radix
// tries) and Merkle Trees. They optimize searches and insertions in radix tries
// for long keys that don't have much in common (which is usually the case in
// applications such as Ethereum).
//
// There are a couple of different kinds of merkle patricia tries in Ethereum.
//
// Taken from https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/#tries-in-ethereum:
//
//  1. State trie: There is one global state trie, and it updates
//     as each new block is included in the chain. In it, a path is always:
//     keccak256(ethereumAddress) and a value is always: rlp(ethereumAccount).
//     More specifically an ethereum account is a 4 item array of [nonce,balance,storageRoot,codeHash].
//     At this point it's worth noting that this storageRoot is the root of another patricia trie:
//
//  2. Storage trie: is where all contract data lives. There is a separate storage trie
//     for each account. To retrieve values at specific storage positions at a given
//     address the storage address, integer position of the stored data in the storage,
//     and the block ID are required.
//
//  3. Transactions trie: There is a separate transactions trie for every block, again storing (key, value) pairs.
//     A path here is: rlp(transactionIndex) which represents the key that corresponds to a
//     value determined by:
//
//     IF legacyTx THEN value = rlp(tx) ELSE value = TxType | encode(tx)
//
//  4. Every block has its own Receipts trie. A path here is: rlp(transactionIndex).
//     transactionIndex is its index within the block it's mined. he receipts trie never updates. Similarly to the
//     Transactions trie, there are current and legacy receipts. To query a specific receipt in the
//     Receipts trie the index of the transaction in its block, the receipt payload and the transaction type
//     are required. The Returned receipt can be of type Receipt which is defined as the concatenation of
//     transaction type and transaction payload or it can be of type LegacyReceipt which is defined as
//     rlp([status, cumulativeGasUsed, logsBloom, logs]).
package patricia

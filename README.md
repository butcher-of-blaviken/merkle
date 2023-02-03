# merkle

`merkle` is a small Go library that implements a [Merkle tree](https://en.wikipedia.org/wiki/Merkle_tree).

Written primarily for educational purposes rather than production `:-)`.

## Install

```bash
go get github.com/butcher-of-blaviken/merkle
```

## Usage

```golang
tree, err := New([][]byte{
    []byte("hello"),
    []byte("world"),
    []byte("today"),
    []byte("yes"),
})
if err != nil {
    panic(err)
}

// Get the merkle proof for "hello"
proof, err := tree.ProofFor(0)
if err != nil {
    panic(err)
}

// verify the proof
if !tree.Verify(proof, sha256.Sum256([]byte("hello")), tree.Root()) {
    panic("couldn't verify merkle proof")
}
```

## Testing

```bash
go test -cover
```

# Some python helpers for interactive usage in the Python REPL
# NOT tested

# Cribbed from
# https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/#specification
def compact_encode(hexarray):
    term = 1 if hexarray[-1] == 16 else 0
    if term: hexarray = hexarray[:-1]
    oddlen = len(hexarray) % 2
    flags = 2 * term + oddlen
    if oddlen:
        hexarray = [flags] + hexarray
    else:
        hexarray = [flags] + [0] + hexarray
    #  hexarray now has an even length whose first nibble is the flags.
    print(hexarray)
    o = ''
    for i in range(0,len(hexarray),2):
        o += chr(16 * hexarray[i] + hexarray[i+1])
    return o


# Ported from the go implementation in utils.go
def bytes_to_nibbles(b: bytes):
    nibbles = [0]*(len(b)*2)
    for i, b in enumerate(b):
        nibbles[2*i] = b >> 4
        nibbles[2*i+1] = b % 16
    return nibbles

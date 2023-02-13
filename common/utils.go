package common

func Concat(a, b []byte) (r []byte) {
	r = append(r, a...)
	r = append(r, b...)
	return
}

// bytesToNibbles transforms the given []byte-ified hex string
// to a byte slice where each byte is exactly in the range
// of 0-f (i.e, a nibble - half a byte or 4 bits).
func BytesToNibbles(s []byte) (nibbles []byte) {
	nibbles = make([]byte, len(s)*2)
	for i, b := range s {
		nibbles[2*i] = b >> 4
		nibbles[2*i+1] = b % 16
	}
	return
}

func ExtractCommonPrefix(a, b []byte) (r []byte) {
	i := 0
	for i < len(a) && i < len(b) {
		if a[i] == b[i] {
			r = append(r, a[i])
		} else {
			break
		}
		i++
	}
	return
}

// CompactEncode prepends the appropriate flags to the provided path.
// See the table below for details.
// It returns the path in bytes instead of nibbles.
//
// hexchar  |  bits    |    node type partial  |   path length
//
//	0       |  0000    |       extension       |       even
//	1       |  0001    |       extension       |       odd
//	2       |  0010    |   terminating (leaf)  |       even
//	3       |  0011    |   terminating (leaf)  |       odd
func CompactEncode(b []byte, isLeaf bool) (r []byte) {
	// prefix the provided nibbles depending on the number of nibbles
	if len(b)%2 == 0 {
		// even
		b = append([]byte{0, 0}, b...)
	} else {
		// odd
		b = append([]byte{1}, b...)
	}
	// result must always be of even length
	if len(b)%2 != 0 {
		panic("invariant violated")
	}
	if isLeaf {
		b[0] += 2
	}
	// transform back to bytes
	for i := 0; i < len(b); i += 2 {
		r = append(r, 16*b[i]+b[i+1])
	}
	return
}

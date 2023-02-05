package merkle

func concat(a, b []byte) (r []byte) {
	r = append(r, a...)
	r = append(r, b...)
	return
}

// bytesToNibbles transforms the given []byte-ified hex string
// to a byte slice where each byte is exactly in the range
// of 0-f (i.e, a nibble - half a byte or 4 bits).
func bytesToNibbles(s []byte) (nibbles []byte) {
	nibbles = make([]byte, len(s)*2)
	for i, b := range s {
		nibbles[2*i] = b >> 4
		nibbles[2*i+1] = b % 16
	}
	return
}

func extractCommonPrefix(a, b []byte) (r []byte) {
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

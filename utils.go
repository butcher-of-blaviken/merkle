package merkle

func concat(a, b []byte) (r []byte) {
	r = append(r, a...)
	r = append(r, b...)
	return
}

// keyToHex transforms the given []byte-ified hex string
// to a byte slice where each byte is exactly in the range
// of 0-f (i.e, a nibble - half a byte or 4 bits).
func keybytesToHex(s []byte) []byte {
	l := len(s)*2 + 1
	var nibbles = make([]byte, l)
	for i, b := range s {
		nibbles[i*2] = b / 16
		nibbles[i*2+1] = b % 16
	}
	nibbles[l-1] = 16
	return nibbles
}

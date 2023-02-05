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

func CompactEncode(b []byte) []byte {
	// b is nibbles
	var term int
	if b[len(b)-1] == 0xf {
		term = 1
	}
	if term == 1 {
		b = b[:len(b)-1]
	}
	oddlen := len(b) % 2
	flags := 2*term + oddlen
	if oddlen == 1 {
		// odd length path
		b = append([]byte{byte(flags)}, b...)
	} else {
		// even length path
		b = append([]byte{byte(flags), 0}, b...)
	}
	return b
}

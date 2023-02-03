package merkle

func concat(a, b []byte) (r []byte) {
	r = append(r, a...)
	r = append(r, b...)
	return
}

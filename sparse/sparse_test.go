package sparse_test

import (
	"testing"

	"github.com/butcher-of-blaviken/merkle/sparse"
	"github.com/stretchr/testify/assert"
)

func TestSparse(t *testing.T) {
	a := sparse.New()
	a.Set(0)
	a.Set(32)
	a.Set(63)
	assert.True(t, a.Get(0))
	assert.True(t, a.Get(32))
	assert.True(t, a.Get(63))
	a.Unset(32)
	assert.False(t, a.Get(32))
}

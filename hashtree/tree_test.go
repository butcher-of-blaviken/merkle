package hashtree

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeHash(s string) (r Bytes32) {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	copy(r[:], b)
	return
}

func TestNew(t *testing.T) {
	type args struct {
		data [][]byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Tree
		wantErr bool
	}{
		{
			name: "basic",
			args: args{
				data: [][]byte{
					[]byte("hello"),
					[]byte("world"),
					[]byte("today"),
					[]byte("yes"),
				},
			},
			want: &Tree{
				levels: []level{
					{
						sha256.Sum256([]byte("hello")),
						sha256.Sum256([]byte("world")),
						sha256.Sum256([]byte("today")),
						sha256.Sum256([]byte("yes")),
					},
					{
						makeHash("7305db9b2abccd706c256db3d97e5ff48d677cfe4d3a5904afb7da0e3950e1e2"),
						makeHash("aa543d2950bc2e1274e337e7064b4999538d7de1bd57ec0e9c8df011550367c1"),
					},
					{
						makeHash("ce1883478771a5921ef628afbbfd222f36dda955d2b311a2a722fe314f825046"),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "not a power of 2",
			args: args{
				data: [][]byte{
					[]byte("hello"),
					[]byte("world"),
					[]byte("today"),
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.data)
			if tt.wantErr {
				require.Error(t, err)
				return
			} else {
				require.NoError(t, err)
			}
			if !reflect.DeepEqual(got.levels, tt.want.levels) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProofFor(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		tree, err := New([][]byte{
			[]byte("hello"),
			[]byte("world"),
			[]byte("today"),
			[]byte("yes"),
		})
		require.NoError(t, err)
		require.Equal(t, 3, tree.Height())
		require.Equal(t, "ce1883478771a5921ef628afbbfd222f36dda955d2b311a2a722fe314f825046", tree.Root().String())

		// get proof for "hello"
		proof, err := tree.ProofFor(0)
		require.NoError(t, err)
		require.Len(t, proof.Hashes, 2)
		require.Equal(t, "486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7", proof.Hashes[0].String())
		require.Equal(t, "aa543d2950bc2e1274e337e7064b4999538d7de1bd57ec0e9c8df011550367c1", proof.Hashes[1].String())

		// get proof for "world"
		proof, err = tree.ProofFor(1)
		require.NoError(t, err)
		require.Len(t, proof.Hashes, 2)
		assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", proof.Hashes[0].String())
		assert.Equal(t, "aa543d2950bc2e1274e337e7064b4999538d7de1bd57ec0e9c8df011550367c1", proof.Hashes[1].String())

		// get proof for "today"
		proof, err = tree.ProofFor(2)
		require.NoError(t, err)
		require.Len(t, proof.Hashes, 2)
		assert.Equal(t, "8a798890fe93817163b10b5f7bd2ca4d25d84c52739a645a889c173eee7d9d3d", proof.Hashes[0].String())
		assert.Equal(t, "7305db9b2abccd706c256db3d97e5ff48d677cfe4d3a5904afb7da0e3950e1e2", proof.Hashes[1].String())

		// get proof for "yes"
		proof, err = tree.ProofFor(3)
		require.NoError(t, err)
		require.Len(t, proof.Hashes, 2)
		assert.Equal(t, "e0f4f767ac88a9303e7317843ac20be980665a36f52397e5b26d4cc2bf54011d", proof.Hashes[0].String())
		assert.Equal(t, "7305db9b2abccd706c256db3d97e5ff48d677cfe4d3a5904afb7da0e3950e1e2", proof.Hashes[1].String())
	})
	t.Run("bit bigger", func(t *testing.T) {
		tree, err := New([][]byte{
			[]byte("hello"),
			[]byte("world"),
			[]byte("today"),
			[]byte("yes"),
			[]byte("no"),
			[]byte("idontknow"),
			[]byte("maybe"),
			[]byte("so"),
		})
		require.NoError(t, err)
		require.Equal(t, 4, tree.Height())
		require.Equal(t, "478d6126254478195bfdb902295a429e1fa9f99cdca928b629959f40ee521c2e", tree.Root().String())

		// get proof for "hello"
		proof, err := tree.ProofFor(0)
		require.NoError(t, err)
		require.Len(t, proof.Hashes, 3)
		assert.Equal(t, "486ea46224d1bb4fb680f34f7c9ad96a8f24ec88be73ea8e5a6c65260e9cb8a7", proof.Hashes[0].String())
		assert.Equal(t, "aa543d2950bc2e1274e337e7064b4999538d7de1bd57ec0e9c8df011550367c1", proof.Hashes[1].String())
		assert.Equal(t, "31789922ed43ddc635da50f3155b836fed13514afaeacebb743d57e258e1b12f", proof.Hashes[2].String())
	})
}

func TestGetSiblingIndex(t *testing.T) {
	tree, err := New([][]byte{
		[]byte("hello"),
		[]byte("world"),
		[]byte("today"),
		[]byte("yes"),
	})
	require.NoError(t, err)
	require.Equal(t, 3, tree.Height())
	require.Equal(t, "ce1883478771a5921ef628afbbfd222f36dda955d2b311a2a722fe314f825046", tree.Root().String())

	assert.Equal(t, 1, getSiblingIndex(0, 0))
	assert.Equal(t, 0, getSiblingIndex(1, 0))
}

func TestVerify(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		tree, err := New([][]byte{
			[]byte("hello"),
			[]byte("world"),
			[]byte("today"),
			[]byte("yes"),
		})
		require.NoError(t, err)
		require.Equal(t, 3, tree.Height())
		require.Equal(t, "ce1883478771a5921ef628afbbfd222f36dda955d2b311a2a722fe314f825046", tree.Root().String())

		proof, err := tree.ProofFor(0)
		require.NoError(t, err)
		require.Len(t, proof, 2)

		require.True(t, Verify(proof, tree.levels[0][0], tree.Root()))
	})

	t.Run("bit bigger", func(t *testing.T) {
		tree, err := New([][]byte{
			[]byte("hello"),
			[]byte("world"),
			[]byte("today"),
			[]byte("yes"),
			[]byte("no"),
			[]byte("idontknow"),
			[]byte("maybe"),
			[]byte("so"),
		})
		require.NoError(t, err)
		require.Equal(t, 4, tree.Height())
		require.Equal(t, "478d6126254478195bfdb902295a429e1fa9f99cdca928b629959f40ee521c2e", tree.Root().String())

		proof, err := tree.ProofFor(0)
		require.NoError(t, err)
		require.Len(t, proof, 3)

		require.True(t, Verify(proof, tree.levels[0][0], tree.Root()))
	})
}

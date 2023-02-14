package patricia_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/butcher-of-blaviken/merkle/patricia"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	gethTrie "github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"
)

func receiptsFromJSON(t *testing.T, receiptsJSONPath string) (r types.Receipts) {
	f, err := os.Open(receiptsJSONPath)
	require.NoError(t, err)
	defer f.Close()
	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(contents, &r))
	return
}

func TestReceiptsRootEIP1559(t *testing.T) {
	header := headerFromJSON(t, "testdata/16614538/header.json")
	receipts := receiptsFromJSON(t, "testdata/16614538/receipts.json")
	mpt := patricia.New()
	gtrie := gethTrie.NewEmpty(gethTrie.NewDatabase(rawdb.NewMemoryDatabase()))
	// Using DeriveSha is quite convenient, from the geth API.
	receiptsRoot := types.DeriveSha(receipts, mpt)
	gethRoot := types.DeriveSha(receipts, gtrie)
	require.Equal(t, gethRoot, receiptsRoot, "my root doesn't match geth root")
	require.Equal(t, header.ReceiptHash, gethRoot)
	require.Equal(t, header.ReceiptHash, receiptsRoot)
}

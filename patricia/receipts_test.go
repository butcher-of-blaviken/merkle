package patricia_test

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/butcher-of-blaviken/merkle/patricia"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	gethTrie "github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"
)

func receiptsFromJSON(t *testing.T, receiptsJSONPath string) (r []*types.Receipt) {
	f, err := os.Open(receiptsJSONPath)
	require.NoError(t, err)
	defer f.Close()
	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(contents, &r))
	return
}

func TestReceiptsRootEIP1559(t *testing.T) {
	// TODO: fix
	t.Skip()

	header := headerFromJSON(t, "testdata/16614538/header.json")
	receipts := receiptsFromJSON(t, "testdata/16614538/receipts.json")
	mpt := patricia.New()
	gtrie := gethTrie.NewEmpty(gethTrie.NewDatabase(rawdb.NewMemoryDatabase()))
	for i, receipt := range receipts {
		// encode into RLP
		indexRLP, err := rlp.EncodeToBytes(uint64(i))
		require.NoError(t, err)
		receiptRLP, err := rlp.EncodeToBytes(receipt)
		require.NoError(t, err)

		// insert into tries
		require.NoError(t, mpt.Put(indexRLP, receiptRLP))
		gtrie.Update(indexRLP, receiptRLP)
	}

	txRoot := mpt.Root()
	gethRoot := gtrie.Hash()
	require.Equal(t, gethRoot.Hex(), hexutil.Encode(txRoot), "my root doesn't match geth root")
	require.Equal(t, header.TxHash.Hex(), gethRoot.Hex())
	require.Equal(t, header.TxHash.Hex(), hexutil.Encode(txRoot))
}

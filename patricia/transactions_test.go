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

func headerFromJSON(t *testing.T, headerJSONPath string) (h *types.Header) {
	h = new(types.Header)
	f, err := os.Open(headerJSONPath)
	require.NoError(t, err)
	defer f.Close()
	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(contents, h))
	return
}

func transactionsFromJSON(t *testing.T, txsJSONPath string) (r []*types.Transaction) {
	f, err := os.Open(txsJSONPath)
	require.NoError(t, err)
	defer f.Close()
	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(contents, &r))
	return
}

func TestUnmarshalTransaction(t *testing.T) {
	f, err := os.Open("testdata/16614538/tx.json")
	require.NoError(t, err)
	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	var tx types.Transaction
	require.NoError(t, json.Unmarshal(contents, &tx))
	t.Log(tx)
}

func TestUnmarshalTransactions(t *testing.T) {
	f, err := os.Open("testdata/16614538/txs.json")
	require.NoError(t, err)
	contents, err := io.ReadAll(f)
	require.NoError(t, err)
	var txs []*types.Transaction
	require.NoError(t, json.Unmarshal(contents, &txs))
	t.Log(txs)
}

func TestTransactionRootEIP1559(t *testing.T) {
	t.Skip("not working - need to fix")

	header := headerFromJSON(t, "testdata/16614538/header.json")
	txs := transactionsFromJSON(t, "testdata/16614538/txs.json")
	mpt := patricia.New()
	gtrie := gethTrie.NewEmpty(gethTrie.NewDatabase(rawdb.NewMemoryDatabase()))
	for i, tx := range txs {
		// encode into RLP
		indexRLP, err := rlp.EncodeToBytes(uint64(i))
		require.NoError(t, err)
		txRLP, err := rlp.EncodeToBytes(tx)
		require.NoError(t, err)

		// insert into tries
		require.NoError(t, mpt.Put(indexRLP, txRLP))
		gtrie.Update(indexRLP, txRLP)
	}
	txRoot := mpt.Root()
	gethRoot := gtrie.Hash()
	require.Equal(t, gethRoot.Hex(), hexutil.Encode(txRoot), "my root doesn't match geth root")
	require.Equal(t, header.TxHash.Hex(), gethRoot.Hex())
	require.Equal(t, header.TxHash.Hex(), hexutil.Encode(txRoot))
}

func TestTransactionsRootLegacy(t *testing.T) {
	header := headerFromJSON(t, "testdata/10467135/header.json")
	txs := transactionsFromJSON(t, "testdata/10467135/txs.json")
	mpt := patricia.New()
	gtrie := gethTrie.NewEmpty(gethTrie.NewDatabase(rawdb.NewMemoryDatabase()))
	for i, tx := range txs {
		// encode into RLP
		indexRLP, err := rlp.EncodeToBytes(uint64(i))
		require.NoError(t, err)
		txRLP, err := rlp.EncodeToBytes(tx)
		require.NoError(t, err)
		// t.Log("RLP encoded tx", tx.Hash, "is:", hexutil.Encode(txRLP))

		// insert into tries
		require.NoError(t, mpt.Put(indexRLP, txRLP))
		gtrie.Update(indexRLP, txRLP)
	}
	txRoot := mpt.Root()
	gethRoot := gtrie.Hash()
	require.Equal(t, gethRoot.Hex(), hexutil.Encode(txRoot), "my root doesn't match geth root")
	require.Equal(t, header.TxHash.Hex(), gethRoot.Hex())
	require.Equal(t, header.TxHash.Hex(), hexutil.Encode(txRoot))
}

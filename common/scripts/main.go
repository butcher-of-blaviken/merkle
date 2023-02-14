package main

import (
	"context"
	"encoding/json"
	"flag"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	// See https://chainlist.org/chain/1, don't spam!
	rpcURL = "https://eth.llamarpc.com"
)

func main() {
	rpcClient, err := rpc.Dial(rpcURL)
	if err != nil {
		panic(err)
	}

	ethClient, err := ethclient.Dial(rpcURL)
	if err != nil {
		panic(err)
	}

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "fetch-receipts":
			cmd := flag.NewFlagSet("fetch-receipts", flag.ExitOnError)
			blockNumber := cmd.Int64("block-number", -1, "block number to fetch receipts for")
			out := cmd.String("out", "receipts.json", "output JSON path of response")
			if err := cmd.Parse(os.Args[2:]); err != nil {
				panic(err)
			}

			block, err := ethClient.BlockByNumber(context.Background(), big.NewInt(*blockNumber))
			if err != nil {
				panic(err)
			}

			txs := block.Transactions()
			var batch []rpc.BatchElem
			var receipts []*types.Receipt
			for _, tx := range txs {
				out := new(types.Receipt)
				batch = append(batch, rpc.BatchElem{
					Method: "eth_getTransactionReceipt",
					Args:   []any{tx.Hash().String()},
					Result: out,
				})
				receipts = append(receipts, out)
			}

			err = rpcClient.BatchCall(batch)
			if err != nil {
				panic(err)
			}

			f, err := os.Create(*out)
			defer f.Close()
			if err != nil {
				panic(err)
			}

			err = json.NewEncoder(f).Encode(receipts)
			if err != nil {
				panic(err)
			}
		}
	}
}

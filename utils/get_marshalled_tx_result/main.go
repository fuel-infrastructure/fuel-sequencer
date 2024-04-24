package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/cometbft/cometbft/libs/bytes"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cometbft/cometbft/types"
)

func main() {
	if len(os.Args) != 6 {
		fmt.Fprintf(os.Stderr, "Expected exactly 3 args: [node] [height] [tx-index] [start-block] [end-block]")
		os.Exit(1)
	}

	node := os.Args[1]
	height, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		panic(err)
	}
	txIndex, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		panic(err)
	}
	startBlock, err := strconv.ParseUint(os.Args[4], 10, 64)
	if err != nil {
		panic(err)
	}
	endBlock, err := strconv.ParseUint(os.Args[5], 10, 64)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Establishing connection to %s to query tx %d @ block %d...\n", node, txIndex, height)

	httpClient, err := libclient.DefaultHTTPClient(node)
	if err != nil {
		panic(err)
	}

	httpClient.Timeout = 10 * time.Second
	rpcClient, err := rpchttp.NewWithClient(node, "/websocket", httpClient)
	if err != nil {
		panic(err)
	}

	result, err := rpcClient.BlockResults(context.Background(), &height)
	if err != nil {
		panic(err)
	}

	deterministicExecTxResults := types.NewResults(result.TxsResults)
	if txIndex >= int64(len(deterministicExecTxResults)) {
		panic(fmt.Sprintf("block only has %d txs; cannot get tx at index %d", len(deterministicExecTxResults), txIndex))
	}

	marshalled, err := deterministicExecTxResults[txIndex].Marshal()
	if err != nil {
		panic(err)
	}

	marshalledHexBytes := bytes.HexBytes(marshalled)
	fmt.Println(marshalledHexBytes.String())

	proofs, err := rpcClient.BridgeCommitmentInclusionProof(context.Background(), height+1, txIndex, startBlock, endBlock)
	if err != nil {
		panic(err)
	}
	fmt.Printf("txResultProof: \n")
	for _, aunt := range proofs.LastResultsMerkleProof.Aunts {
		auntBytes := bytes.HexBytes(aunt)
		fmt.Println(auntBytes.String())
	}

	fmt.Printf("bridgeCommitmentLeafProof: \n")
	for _, aunt := range proofs.BridgeCommitmentMerkleProof.Aunts {
		auntBytes := bytes.HexBytes(aunt)
		fmt.Println(auntBytes.String())
	}
}

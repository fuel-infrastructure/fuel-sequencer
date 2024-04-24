package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	libclient "github.com/cometbft/cometbft/rpc/jsonrpc/client"
	"github.com/cometbft/cometbft/types"
)

func exit(format string, a ...any) {
	fmt.Fprintf(os.Stderr, format, a)
	os.Exit(1)
}

func main() {
	expectedArgs := 6
	if len(os.Args) != expectedArgs {
		exit("Expected exactly %d args: [node] [height] [tx-index] [start-block] [end-block]", expectedArgs)
	}

	node := os.Args[1]
	height, err := strconv.ParseInt(os.Args[2], 10, 64)
	if err != nil {
		exit(err.Error())
	}
	txIndex, err := strconv.ParseInt(os.Args[3], 10, 64)
	if err != nil {
		exit(err.Error())
	}
	startBlock, err := strconv.ParseUint(os.Args[4], 10, 64)
	if err != nil {
		exit(err.Error())
	}
	endBlock, err := strconv.ParseUint(os.Args[5], 10, 64)
	if err != nil {
		exit(err.Error())
	}

	fmt.Printf("Configuration: "+
		"height=%d, "+
		"tx-index=%d, "+
		"start-block=%d, "+
		"end-block=%d\n",
		height, txIndex, startBlock, endBlock,
	)
	fmt.Printf("Establishing connection to %s...\n", node)

	httpClient, err := libclient.DefaultHTTPClient(node)
	if err != nil {
		exit(err.Error())
	}

	httpClient.Timeout = 10 * time.Second
	rpcClient, err := rpchttp.NewWithClient(node, "/websocket", httpClient)
	if err != nil {
		exit(err.Error())
	}

	// ---------------------------- Get txResultMarshalled...

	blockResults, err := rpcClient.BlockResults(context.Background(), &height)
	if err != nil {
		exit(err.Error())
	}

	deterministicExecTxResults := types.NewResults(blockResults.TxsResults)
	if txIndex >= int64(len(deterministicExecTxResults)) {
		exit("block only has %d txs; cannot get tx at index %d", len(deterministicExecTxResults), txIndex)
	}

	txResultMarshalled, err := deterministicExecTxResults[txIndex].Marshal()
	if err != nil {
		exit(err.Error())
	}
	fmt.Println("\ntxResultMarshalled: ")
	fmt.Printf("%X\n", txResultMarshalled)

	// ---------------------------- Get txResultProof...

	proofs, err := rpcClient.BridgeCommitmentInclusionProof(context.Background(), height+1, txIndex, startBlock, endBlock)
	if err != nil {
		exit(err.Error())
	}
	fmt.Println("\ntxResultProof: ")
	for _, aunt := range proofs.LastResultsMerkleProof.Aunts {
		fmt.Printf("%X\n", aunt)
	}
	// Alternative: proofs.LastResultsMerkleProof.String()

	// ---------------------------- Get bridgeCommitmentLeafProof...

	fmt.Println("\nbridgeCommitmentLeafProof: ")
	for _, aunt := range proofs.BridgeCommitmentMerkleProof.Aunts {
		fmt.Printf("%X\n", aunt)
	}
	// Alternative: proofs.BridgeCommitmentMerkleProof.String()
}

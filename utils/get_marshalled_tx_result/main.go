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
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "Expected exactly 3 args: [node] [height] [tx-index]")
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
}

#!/usr/bin/env bash

RPC_URL="https://rpc-seq.simplystaking.xyz"
TX_HEIGHT=46299
TX_INDEX=1

go run main.go "$RPC_URL" "$TX_HEIGHT" "$TX_INDEX"

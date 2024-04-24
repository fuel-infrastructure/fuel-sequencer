#!/usr/bin/env bash

RPC_URL="https://rpc-seq.simplystaking.xyz"
TX_HEIGHT=45979
TX_INDEX=1
START_BLOCK=45910
END_BLOCK=46010

go run main.go "$RPC_URL" "$TX_HEIGHT" "$TX_INDEX" "$START_BLOCK" "$END_BLOCK"

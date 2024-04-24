#!/usr/bin/env bash

NODE="https://rpc-seq.simplystaking.xyz"
HEIGHT=46299
TX_INDEX=1

go run main.go "$NODE" "$HEIGHT" "$TX_INDEX"

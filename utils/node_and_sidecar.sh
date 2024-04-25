#!/usr/bin/env bash

# Set Sidecar variables
HOST="0.0.0.0"
COSMOS_NODE_RPC="127.0.0.1:9090"
TENDERMINT_NODE_RPC="http://127.0.0.1:26657"
ETH_RPC="http://ethereum:8545"
CONTRACT_ADDRESS="0xa513E6E4b8f2a923D98304ec87F64353C4D5C853"
ETH_MAX_BLOCK_RANGE="100"
UNSAFE_ETHEREUM_BLOCK="1"

# Start FuelSequencer node (TODO: make customisable)
fuelsequencerd start \
  --sidecar.enabled \
  &

# Start Sidecar (TODO: make customisable).
# NOTE: Here we are assuming that we are running an Anvil node. If this
# is no longer the case we should consider setting development to false.
fuelsequencerd start-sidecar \
  --host "$HOST" \
  --cosmos_node_rpc "$COSMOS_NODE_RPC" \
  --tendermint_node_rpc "$TENDERMINT_NODE_RPC" \
  --eth_node_rpc "$ETH_RPC" \
  --contract_address "$CONTRACT_ADDRESS" \
  --eth_max_block_range "$ETH_MAX_BLOCK_RANGE" \
  --unsafe_ethereum_block "$UNSAFE_ETHEREUM_BLOCK" \
  --development=true \
  &

# Wait for all background jobs to finish
wait

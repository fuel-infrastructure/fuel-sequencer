#!/usr/bin/env bash

# Set Sidecar variables
SIDECAR_HOST="0.0.0.0"
SIDECAR_PORT="8080"
SEQUENCER_GRPC_URL="127.0.0.1:9090"
SEQUENCER_RPC_URL="http://127.0.0.1:26657"
ETH_WS_URL="ws://ethereum:8545"
ETH_CONTRACT_ADDRESS="0xa513E6E4b8f2a923D98304ec87F64353C4D5C853"
ETH_MAX_BLOCK_RANGE="100"
ETH_MIN_LOGS_QUERY_INTERVAL="10s"
ETH_UNSAFE_START_BLOCK="1"
DEVELOPMENT="true"

# Start FuelSequencer node (TODO: make customisable)
fuelsequencerd start \
  --sidecar.enabled \
  &

# Start Sidecar (TODO: make customisable).
# NOTE: Here we are assuming that we are running an Anvil node. If this
# is no longer the case we should consider setting development to false.
fuelsequencerd start-sidecar \
  --host "$SIDECAR_HOST" \
  --port "$SIDECAR_PORT" \
  --sequencer_rpc_url "$SEQUENCER_RPC_URL" \
  --sequencer_grpc_url "$SEQUENCER_GRPC_URL" \
  --eth_ws_url "$ETH_WS_URL" \
  --eth_contract_address "$ETH_CONTRACT_ADDRESS" \
  --eth_max_block_range "$ETH_MAX_BLOCK_RANGE" \
  --eth_min_logs_query_interval "$ETH_MIN_LOGS_QUERY_INTERVAL" \
  --unsafe_eth_start_block "$ETH_UNSAFE_START_BLOCK" \
  --development "$DEVELOPMENT" \
  &

# Wait for all background jobs to finish
wait

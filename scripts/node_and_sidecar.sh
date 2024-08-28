#!/usr/bin/env bash

# Set Sidecar variables
SIDECAR_HOST="0.0.0.0"
SIDECAR_PORT="8080"
SEQUENCER_GRPC_URL="127.0.0.1:9090"
ETH_WS_URL="ws://ethereum-node:8545" # set to the same port as RPC because e2e testing uses anvil nodes
ETH_RPC_URL="http://ethereum-node:8545"
ETH_CONTRACT_ADDRESS="0x0165878A594ca255338adfa4d48449f69242Eb8F"
ETH_MAX_BLOCK_RANGE="100"
ETH_MIN_LOGS_QUERY_INTERVAL="1s" # this is low because this script is used for E2E purposes where the block time is 1s
ETH_UNSAFE_START_BLOCK="1"
DEVELOPMENT="true"
PROMETHEUS_ENABLED="true"

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
  --sequencer_grpc_url "$SEQUENCER_GRPC_URL" \
  --eth_ws_url "$ETH_WS_URL" \
  --eth_rpc_url "$ETH_RPC_URL" \
  --eth_contract_address "$ETH_CONTRACT_ADDRESS" \
  --eth_max_block_range "$ETH_MAX_BLOCK_RANGE" \
  --eth_min_logs_query_interval "$ETH_MIN_LOGS_QUERY_INTERVAL" \
  --unsafe_eth_start_block "$ETH_UNSAFE_START_BLOCK" \
  --development "$DEVELOPMENT" \
  --prometheus_enabled "$PROMETHEUS_ENABLED" \
  &

# Wait for all background jobs to finish
wait

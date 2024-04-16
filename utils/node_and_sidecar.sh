#!/usr/bin/env bash

# Start FuelSequencer node (TODO: make customisable)
fuelsequencerd start \
  --sidecar.enabled \
  &

# Start Sidecar (TODO: make customisable).
# NOTE: Here we are assuming that we are running an Anvil node. If this is no longer the case we should consider setting
# development to false.
fuelsequencerd start-sidecar \
  --host "0.0.0.0" \
  --eth_node_rpc "http://ethereum:8545" \
  --contract_address "0xa513E6E4b8f2a923D98304ec87F64353C4D5C853" \
  --development=true \
  &

# Wait for all background jobs to finish
wait

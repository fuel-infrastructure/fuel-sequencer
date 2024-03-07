#!/usr/bin/env bash

# Start FuelSequencer node (TODO: make customisable)
fuelsequencerd start \
  --sidecar.enabled \
  &

# Start Sidecar (TODO: make customisable)
fuelsequencerd start-sidecar \
  --eth_node_rpc "http://ethereum:8545" \
  --contract_address "0xDc64a140Aa3E981100a9becA4E685f962f0cF6C9" \
  &

# Wait for all background jobs to finish
wait

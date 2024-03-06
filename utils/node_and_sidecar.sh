#!/usr/bin/env bash

# Start FuelSequencer node (TODO: make customisable)
fuelsequencerd start \
  --sidecar.enabled \
  &

# Start Sidecar (TODO: make customisable)
fuelsequencerd start-sidecar \
  --eth_node_rpc "http://ethereum:8545" \
  --contract_address "0x5FbDB2315678afecb367f032d93F642f64180aa3" \
  &

# Wait for all background jobs to finish
wait

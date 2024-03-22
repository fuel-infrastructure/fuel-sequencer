#!/usr/bin/env bash

# Start FuelSequencer node (TODO: make customisable)
fuelsequencerd start \
  --sidecar.enabled \
  &

# Start Sidecar (TODO: make customisable)
fuelsequencerd start-sidecar \
  --host "0.0.0.0" \
  --eth_node_rpc "http://ethereum:8545" \
  --contract_address "0x101E64349abe34E53e3E6AAbE009197240AaE1cD" \
  --development=true \
  &

# Wait for all background jobs to finish
wait

#!/bin/bash

ETH_DEPLOYMENT_CONTAINER_NAME=deploy

set -e

logs=$(docker logs $ETH_DEPLOYMENT_CONTAINER_NAME)

get_address_from_first_match() {
    local reString="$1"
    if [[ $logs =~ $reString ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        echo "No match found for regex: $reString" >&2
        exit 1
    fi
}

FuelStreamXContractAddress=$(get_address_from_first_match "Deployed FuelStreamX at (0x[a-fA-F0-9]{40})")
SequencerInterfaceContractAddress=$(get_address_from_first_match "Deployed SequencerInterface at (0x[a-fA-F0-9]{40})")
MigratedTokenContractAddress=$(get_address_from_first_match "deploying \"MigratedToken\" \(tx: 0x[a-fA-F0-9]{64}\)\.\.\.: deployed at (0x[a-fA-F0-9]{40})")
VaultContractAddress=$(get_address_from_first_match "Deployed Vault at (0x[a-fA-F0-9]{40})")
SequencerProxyContractAddress=$(get_address_from_first_match "Deployed SequencerProxy at (0x[a-fA-F0-9]{40})")
TokenMigratorContractAddress=$(get_address_from_first_match "Deployed TokenMigrator at (0x[a-fA-F0-9]{40})")
TokenContractAddress=$(get_address_from_first_match "Deployed Token at (0x[a-fA-F0-9]{40})")
FaucetContractAddress=$(get_address_from_first_match "deploying \"TokenFaucet\" \(tx: 0x[a-fA-F0-9]{64}\)\.\.\.: deployed at (0x[a-fA-F0-9]{40})")

echo "$FuelStreamXContractAddress :: FuelStreamX contract"
echo "$SequencerInterfaceContractAddress :: SequencerInterface contract"
echo "$MigratedTokenContractAddress :: MigratedToken contract"
echo "$VaultContractAddress :: Vault contract"
echo "$SequencerProxyContractAddress :: SequencerProxy contract"
echo "$TokenMigratorContractAddress :: TokenMigrator contract"
echo "$TokenContractAddress :: Token contract"
echo "$FaucetContractAddress :: Faucet contract"

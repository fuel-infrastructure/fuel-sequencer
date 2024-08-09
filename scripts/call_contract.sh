#!/usr/bin/env bash

# -------------------------------- Get contract addresses from running container

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

SequencerInterfaceContractAddress=$(get_address_from_first_match "Deployed SequencerInterface at (0x[a-fA-F0-9]{40})")
TokenContractAddress=$(get_address_from_first_match "Deployed Token at (0x[a-fA-F0-9]{40})")

echo "$SequencerInterfaceContractAddress :: SequencerInterface contract"
echo "$TokenContractAddress :: Token contract"

# -------------------------------- Call the contracts

PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
TOKEN_CONTRACT="$TokenContractAddress"
SEQUENCER_INTERFACE_CONTRACT="$SequencerInterfaceContractAddress"
RPC_URL=http://localhost:8545

echo ""
echo "----------------------------------------------------------------------------------------------------------------------------------"
echo "Minting 100 utest to 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 on Ethereum..."

cast send --private-key $PRIVATE_KEY $TOKEN_CONTRACT --rpc-url $RPC_URL "mint(address,uint256)" 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 100

echo ""
echo "----------------------------------------------------------------------------------------------------------------------------------"
echo "Sending a deposit of 100 utest to 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266..."

cast send --private-key $PRIVATE_KEY $TOKEN_CONTRACT --rpc-url $RPC_URL "transferAndCall(address,uint256)" $SEQUENCER_INTERFACE_CONTRACT 100

echo ""
echo "----------------------------------------------------------------------------------------------------------------------------------"
echo "Authorizing a MsgSend of 10 utest from 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266 to 0xd447066a8ba9cb15a862a0f6de961f27be86fc0a..."

cast send --private-key $PRIVATE_KEY $SEQUENCER_INTERFACE_CONTRACT --rpc-url $RPC_URL "batchAuthorize(bytes[])" [0a85010a1c2f636f736d6f732e62616e6b2e763162657461312e4d736753656e6412650a2a307866333946643665353161616438384636463463653661423838323732373963666646623932323636122a3078643434373036366138626139636231356138363261306636646539363166323762653836666330611a0b0a05757465737412023130]

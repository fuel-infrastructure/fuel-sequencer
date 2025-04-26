#!/bin/bash

# Function to get module account address
get_module_address() {
    local module_name=$1
    fuelsequencerd query auth module-accounts --output json | jq -r ".accounts[] | select(.value.name == \"$module_name\") | .value.address"
}

# Function to get validator addresses
get_validator_addresses() {
    fuelsequencerd keys list --output json | jq -r '.[].address'
}

# Get chain ID
CHAIN_ID=$(fuelsequencerd status | jq -r '.node_info.network')
echo "Using chain ID: $CHAIN_ID"

# Get governance module address
GOV_ADDRESS=$(get_module_address "gov")
if [ -z "$GOV_ADDRESS" ]; then
    echo "Error: Could not get governance module address"
    exit 1
fi
echo "Governance module address: $GOV_ADDRESS"

# Get validator addresses
VALIDATOR_ADDRESSES=($(get_validator_addresses))
if [ ${#VALIDATOR_ADDRESSES[@]} -lt 2 ]; then
    echo "Error: Need at least 2 validators for voting"
    exit 1
fi
echo "Found ${#VALIDATOR_ADDRESSES[@]} validators"

# Create the proposal JSON file in the current directory
cat > bond.json << EOF
{
    "messages": [
        {
            "@type": "/fuelsequencer.bond.v1.MsgUpdateParams",
            "authority": "$GOV_ADDRESS",
            "params": {
                "inflation": "0.200000000000000000",
                "authority": "$GOV_ADDRESS"
            }
        }
    ],
    "deposit": "6000000000utest",
    "title": "Update Bond Module Parameters",
    "summary": "Update bond module inflation to 0.2 and set authority to governance module"
}
EOF

# Submit the proposal
echo "Submitting proposal..."
fuelsequencerd tx gov submit-proposal bond.json \
    --from ${VALIDATOR_ADDRESSES[0]} \
    --fees 2000utest \
    --chain-id $CHAIN_ID \
    -y

# Wait for proposal to be processed
echo "Waiting for proposal to be processed..."
sleep 10

# Get the latest proposal ID
PROPOSAL_ID=$(fuelsequencerd query gov proposals --output json | jq -r '.proposals[-1].id')
echo "Proposal ID: $PROPOSAL_ID"

# Wait for proposal to be active
echo "Waiting for proposal to be active..."
start_time=$(date +%s)
while true; do
    current_time=$(date +%s)
    if [ $((current_time - start_time)) -gt 30 ]; then
        echo "Error: Timeout waiting for proposal to become active"
        exit 1
    fi
    PROPOSAL_STATUS=$(fuelsequencerd query gov proposal $PROPOSAL_ID --output json | jq -r '.proposal.status')
    if [ "$PROPOSAL_STATUS" == "2" ]; then
        break
    fi
    sleep 2
done

# Have validators vote
echo "Voting on proposal..."
for i in {0..1}; do
    echo "Validator ${VALIDATOR_ADDRESSES[$i]} voting..."
    fuelsequencerd tx gov vote $PROPOSAL_ID yes \
        --from ${VALIDATOR_ADDRESSES[$i]} \
        --fees 2000utest \
        --chain-id $CHAIN_ID \
        -y
done

# Wait for voting period to end and votes to be counted
echo "Waiting for voting period to end and votes to be counted..."
sleep 30

# Verify proposal status
echo "Verifying proposal status..."
PROPOSAL_STATUS=$(fuelsequencerd query gov proposal $PROPOSAL_ID --output json | jq -r '.proposal.status')
if [ "$PROPOSAL_STATUS" != "3" ]; then
    echo "Error: Proposal did not pass. Status: $PROPOSAL_STATUS"
    exit 1
fi
echo "Proposal passed successfully!"

# Verify the updated parameters
echo "Verifying updated parameters..."
BOND_PARAMS=$(fuelsequencerd query bond params --output json)
INFLATION=$(echo $BOND_PARAMS | jq -r '.params.inflation')
AUTHORITY=$(echo $BOND_PARAMS | jq -r '.params.authority')

# Convert inflation to decimal for comparison
EXPECTED_INFLATION="200000000000000000"
if [ "$INFLATION" != "$EXPECTED_INFLATION" ]; then
    echo "Error: Inflation not updated correctly. Got: $INFLATION, Expected: $EXPECTED_INFLATION"
    exit 1
fi

if [ "$AUTHORITY" != "$GOV_ADDRESS" ]; then
    echo "Error: Authority not updated correctly. Got: $AUTHORITY, Expected: $GOV_ADDRESS"
    exit 1
fi

echo "All parameters updated successfully!"
echo "Inflation: $INFLATION"
echo "Authority: $AUTHORITY"

# Clean up
rm -f bond.json 
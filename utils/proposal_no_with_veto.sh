# This script is meant to be run

PROPOSAL_ID=1

fuelsequencerd tx gov submit-proposal utils/proposal.json --from alice --home ./data/fuelsequencer --chain-id fuelsequencer-1 -y
sleep 1
fuelsequencerd q gov proposal $PROPOSAL_ID

fuelsequencerd tx gov vote $PROPOSAL_ID no_with_veto --from alice --home ./data/fuelsequencer --chain-id fuelsequencer-1 -y
sleep 1
fuelsequencerd q gov proposal $PROPOSAL_ID

sleep 10
fuelsequencerd q bridge show-supply-delta-info

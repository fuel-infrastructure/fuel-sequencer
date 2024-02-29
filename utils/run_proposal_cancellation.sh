PROPOSAL_ID=1

fuelsequencerd tx gov submit-proposal utils/proposal_community_spend.json --from alice --home ./data/fuelsequencer --chain-id fuelsequencer-1 -y
sleep 1
fuelsequencerd q gov proposal $PROPOSAL_ID

fuelsequencerd tx gov cancel-proposal $PROPOSAL_ID --from alice --home ./data/fuelsequencer --chain-id fuelsequencer-1 -y
sleep 1
fuelsequencerd q bridge show-supply-delta-info

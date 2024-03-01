PROPOSAL_ID=2

fuelsequencerd tx gov submit-proposal utils/proposal_community_spend.json --from alice --home ./data/fuelsequencer --chain-id fuelsequencer-1 -y
sleep 1
fuelsequencerd q gov proposal $PROPOSAL_ID

fuelsequencerd tx gov vote $PROPOSAL_ID no_with_veto --from alice --home ./data/fuelsequencer --chain-id fuelsequencer-1 -y
sleep 1
fuelsequencerd q gov proposal $PROPOSAL_ID

sleep 10
fuelsequencerd q bridge show-supply-delta-info

# After running this script, you can look up "proposal tallied" in the node's logs.
# If you have debug logging enabled, you will notice that the change in supply is less.

package main

import (
	"fmt"
	"os"

	"cosmossdk.io/math"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/spf13/cobra"
)

const (
	// The upgrade name as defined in the upgrade handler
	UpgradeName = "bond-module"
	// The height at which the upgrade should occur
	UpgradeHeight = 1000000 // Replace with your desired upgrade height
)

func main() {
	cmd := &cobra.Command{
		Use:   "submit-bond-upgrade",
		Short: "Submit a proposal to upgrade the chain with the bond module",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Get the proposer address
			proposer := clientCtx.GetFromAddress()

			// Create the upgrade plan
			plan := upgradetypes.Plan{
				Name:   UpgradeName,
				Height: UpgradeHeight,
				Info:   "Adding bond module for FUEL token staking",
			}

			// Create the software upgrade proposal message
			msg := &upgradetypes.MsgSoftwareUpgrade{
				Authority: proposer.String(),
				Plan:      plan,
			}

			// Create the proposal message
			proposalMsg, err := govv1.NewMsgSubmitProposal(
				[]sdk.Msg{msg},
				sdk.NewCoins(sdk.NewCoin("ufuel", math.NewInt(1000000))), // Replace with your desired deposit amount
				proposer.String(),
				"Bond Module Upgrade", // title
				"Proposal to add the bond module for FUEL token staking", // summary
				"",    // metadata
				false, // expedited
			)
			if err != nil {
				return err
			}

			// Set up the transaction factory
			txf, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf = txf.WithTxConfig(clientCtx.TxConfig).
				WithAccountRetriever(clientCtx.AccountRetriever).
				WithKeybase(clientCtx.Keyring).
				WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

			// Broadcast the transaction
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, proposalMsg)
		},
	}

	// Add flags
	cmd.Flags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")
	cmd.Flags().String(flags.FlagChainID, "", "The network chain ID")
	cmd.Flags().String(flags.FlagNode, "tcp://localhost:26657", "RPC node address")
	cmd.Flags().String(flags.FlagKeyringBackend, keyring.BackendFile, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flags.FlagHome, os.ExpandEnv("$HOME/.fuelsequencer"), "Directory for config and data")

	// Execute the command
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

/*
CLI Approach:

1. First, create a proposal JSON file (proposal.json):
{
  "title": "Bond Module Upgrade",
  "description": "Proposal to add the bond module for FUEL token staking",
  "plan": {
    "name": "bond-module",
    "height": "1000000",
    "info": "Adding bond module for FUEL token staking"
  },
  "deposit": "1000000ufuel"
}

2. Submit the proposal using the CLI:
fuelsequencer tx gov submit-proposal software-upgrade proposal.json --from <your-key-name> --chain-id <chain-id> --node <node-address>

3. Vote on the proposal:
fuelsequencer tx gov vote <proposal-id> yes --from <your-key-name> --chain-id <chain-id> --node <node-address>

4. Wait for the proposal to pass and reach the upgrade height
5. Update the binary to the new version
6. Restart the chain
*/

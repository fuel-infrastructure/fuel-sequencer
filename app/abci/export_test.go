package abci

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bridgetypes "github.com/fuel-infrastructure/fuel-sequencer/x/bridge/types"
)

// AuthenticateTx is an export of authenticateTx for testing.
func (h *FuelSequencerProposalHandler) AuthenticateTx(
	sender string, msgs []sdk.Msg, params *bridgetypes.Params, blockedAddresses map[string]bool,
) error {
	return h.authenticateTx(sender, msgs, params, blockedAddresses)
}

package abci

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuthenticateTx is an export of authenticateTx for testing.
func (h *FuelSequencerProposalHandler) AuthenticateTx(
	sender string, msgs []sdk.Msg, blockedAddresses map[string]bool,
) error {
	return h.authenticateTx(sender, msgs, blockedAddresses)
}

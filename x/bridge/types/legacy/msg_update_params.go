package legacy

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.Msg = &MsgUpdateParams{}

// ValidateBasic does a sanity check on the provided data.
func (m *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}

	// Basic validation - we don't need full params validation for legacy messages
	// since they're only used for decoding existing proposals
	if m.Params.BridgeDenom == "" {
		return errorsmod.New("bridge", 1, "bridge denom cannot be empty")
	}

	return nil
}

// GetSigners returns the expected signers for the message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	authority, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{authority}
}

// Route returns the message route.
func (m *MsgUpdateParams) Route() string {
	return "bridge"
}

// Type returns the message type.
func (m *MsgUpdateParams) Type() string {
	return "legacy_update_params"
}

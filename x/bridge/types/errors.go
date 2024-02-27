package types

// DONTCOVER

import (
	sdkerrors "cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidSigner  = sdkerrors.Register(ModuleName, 1100, "expected gov account as only signer for proposal message")
	ErrInvalidAddress = sdkerrors.Register(ModuleName, 1101, "invalid address")
)
